package controller

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/models"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/rwcarlsen/goexif/exif"

	// "github.com/skip2/go-qrcode"
	"google.golang.org/api/option"
)

type Config struct {
	BucketName      string
	ProjectID       string
	CredentialsFile string
}

type FileUpload struct {
	bucket *storage.BucketHandle
	config Config
}

type ImageMetadata struct {
	DateTaken   string `json:"dateTaken"`
	CameraModel string `json:"cameraModel"`
}

type UploadResult struct {
	URL      string `json:"url"`
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
	Detail   textproto.MIMEHeader
	Type     string        `json:"type"`
	Error    string        `json:"error,omitempty"`
	Metadata ImageMetadata `json:"metadata,omitempty"`
}

const (
	maxVideoSize = 400 * 1024 * 1024
)

func extractMetadata(file multipart.File) (ImageMetadata, error) {
	var metadata ImageMetadata

	x, err := exif.Decode(file)
	if err != nil {
		return metadata, fmt.Errorf("failed to decode EXIF: %v", err)
	}

	if dt, err := x.DateTime(); err == nil {
		metadata.DateTaken = dt.Format(time.RFC3339)
	}

	if model, err := x.Get(exif.Model); err == nil {
		if str, err := model.StringVal(); err == nil {
			metadata.CameraModel = str
		}
	}

	return metadata, nil
}

func NewFileUpload(config Config) (*FileUpload, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(config.CredentialsFile)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: config.ProjectID,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating storage client: %v", err)
	}

	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("error accessing bucket: %v", err)
	}

	return &FileUpload{
		bucket: bucket,
		config: config,
	}, nil
}

func isAllowedFileType(filename string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4", ".mov", ".avi":
		return "video", true
	case ".jpg", ".jpeg", ".png", ".gif", ".heic":
		return "image", true
	default:
		return "", false
	}
}

func generateFilePath(fileType, ID string, isSingle, isProfileUser, isQRCode bool, filename string) string {
	timestamp := time.Now().UnixNano()
	if isProfileUser {
		return fmt.Sprintf("User/u%s/%d_%s", ID, timestamp, filename)
	}
	if isSingle {
		return fmt.Sprintf("Event/e%s/CoverImage/%d_%s", ID, timestamp, filename)
	}
	if isQRCode {
		return fmt.Sprintf("Event/e%s/QRCode/%d_%s", ID, timestamp, filename)
	}
	return fmt.Sprintf("Event/e%s/%s/%d_%s", ID, fileType, timestamp, filename)
}

func (fu *FileUpload) uploadFile(file *multipart.FileHeader, ID string, isSingle, isProfileUser, isQRCode bool) (*UploadResult, error) {
	result := &UploadResult{
		FileName: file.Filename,
		Size:     file.Size,
		Detail:   file.Header,
	}

	fileType, allowed := isAllowedFileType(file.Filename)
	if !allowed {
		result.Error = "unsupported file type"
		return result, fmt.Errorf(result.Error)
	}
	result.Type = fileType

	if fileType == "video" && file.Size > maxVideoSize {
		result.Error = "video file exceeds 400MB limit"
		return result, fmt.Errorf(result.Error)
	}

	src, err := file.Open()
	if err != nil {
		result.Error = fmt.Sprintf("error opening file: %v", err)
		return result, fmt.Errorf(result.Error)
	}
	defer src.Close()

	if fileType == "image" {
		if metadata, err := extractMetadata(src); err == nil {
			result.Metadata = metadata
		}
		if _, err := src.Seek(0, 0); err != nil {
			result.Error = fmt.Sprintf("error resetting file pointer: %v", err)
			return result, fmt.Errorf(result.Error)
		}
	}

	filename := generateFilePath(fileType, ID, isSingle, isProfileUser, isQRCode, filepath.Base(file.Filename))
	ctx := context.Background()
	obj := fu.bucket.Object(filename)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, src); err != nil {
		writer.Close()
		result.Error = fmt.Sprintf("error copying file to bucket: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	if err := writer.Close(); err != nil {
		result.Error = fmt.Sprintf("error closing bucket writer: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		result.Error = fmt.Sprintf("error making file public: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	result.URL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", fu.config.BucketName, filename)
	return result, nil
}

func (fu *FileUpload) HandleMultipleUpload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to process multipart form",
		})
	}

	eid := c.Params("eid")
	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no files uploaded",
		})
	}

	var wg sync.WaitGroup
	results := make([]*UploadResult, len(files))
	errChan := make(chan error, len(files))

	for i, file := range files {
		wg.Add(1)
		go func(index int, fileHeader *multipart.FileHeader) {
			defer wg.Done()
			result, err := fu.uploadFile(fileHeader, eid, false, false, false)
			if err != nil {
				errChan <- err
			}
			results[index] = result
		}(i, file)
	}

	wg.Wait()
	close(errChan)

	var successful, failed []*UploadResult
	for _, result := range results {
		if result != nil {
			if result.Error != "" {
				failed = append(failed, result)
			} else {
				successful = append(successful, result)
			}
		}
	}

	return c.JSON(fiber.Map{
		"successful": successful,
		"failed":     failed,
		"total":      len(files),
		"succeeded":  len(successful),
		"fails":      len(failed),
	})
}

func (fu *FileUpload) HandleSingleUpload(c *fiber.Ctx) error {

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to retrieve file",
		})
	}

	eid := c.Params("eid")

	var event models.Event
	if err := database.DBconn.Where("event_id = ?", eid).First(&event).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Event Not Found",
		})
	}

	result, err := fu.uploadFile(file, eid, true, false, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	event.EventImage = result.URL
	if err := database.DBconn.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Not Save In Event",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"result":  result,
	})
}

func (fu *FileUpload) HandleSingleUploadCoverImageUser(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to retrieve file",
		})
	}

	uid := c.Params("uid")
	result, err := fu.uploadFile(file, uid, true, true, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"result":  result,
	})
}

// func (fu *FileUpload) GenerateQR(c *fiber.Ctx) error {
// 	var request struct {
// 		Text string `json:"text"`
// 	}

// 	eid := c.Params("eid")
// 	if err := c.BodyParser(&request); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "Failed to parse JSON",
// 		})
// 	}

// 	qrCode, err := qrcode.New(request.Text, qrcode.Medium)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "Failed to generate QR code",
// 		})
// 	}

// 	png, err := qrCode.PNG(256)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "Failed to create PNG",
// 		})
// 	}

// 	fileHeader := &multipart.FileHeader{
// 		Filename: "qrcode.png",
// 		Size:     int64(len(png)),
// 		Header:   textproto.MIMEHeader{},
// 	}
// 	fileHeader.Header.Set("Content-Type", "image/png")

// 	result, err := fu.uploadFile(fileHeader, eid, false, false, true)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "QR code generated successfully",
// 		"result":  result,
// 	})
// }

func (fu *FileUpload) DeleteFile(c *fiber.Ctx) error {
	var request struct {
		Filename string `json:"filename"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "filename is required",
		})
	}

	request.Filename = strings.Replace(request.Filename, "https://storage.googleapis.com/project-memorybox.firebasestorage.app/", "", 1)

	ctx := context.Background()
	obj := fu.bucket.Object(request.Filename)

	if err := obj.Delete(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to delete file: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("file %s deleted successfully", request.Filename),
	})
}
