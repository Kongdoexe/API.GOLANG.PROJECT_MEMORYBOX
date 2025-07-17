package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ImageResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	UploadAt string `json:"upload_at"`
}

func UploadMediaAPI(c *fiber.Ctx) error {
	eid := c.Params("eid")
	uid := c.Params("uid")

	if eid == "" || uid == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ขาด EventID และ UserID",
		})
	}

	eventID, err := strconv.ParseUint(eid, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "EventID ไม่ถูกต้อง",
		})
	}

	userID, err := strconv.ParseUint(uid, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "UserID ไม่ถูกต้อง",
		})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ไม่สามารถรับฟอร์มได้",
			"err":   err.Error(),
		})
	}

	files := form.File["media"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "ไม่มีไฟล์อัปโหลด",
		})
	}

	var imageResponses []ImageResponse
	var videoResponses []ImageResponse
	var errors []string

	for i, file := range files {
		contentType := file.Header.Get("Content-Type")

		if !IsValidMediaType(contentType) {
			errors = append(errors, fmt.Sprintf("ไฟล์ %d: อนุญาตเฉพาะรูปภาพหรือวิดีโอเท่านั้น", i+1))
			continue
		}

		if file.Size > dtos.MaxSize {
			errors = append(errors, fmt.Sprintf("ไฟล์ %d: ไฟล์มีขนาดใหญ่เกินไป ขนาดสูงสุดคือ 10MB", i+1))
			continue
		}

		mediaType := "others"
		var fileType int
		if strings.HasPrefix(contentType, "image/") {
			mediaType = "images"
			fileType = 1
		} else if strings.HasPrefix(contentType, "video/") {
			mediaType = "videos"
			fileType = 2
		} else {
			fileType = 0
		}

		userDir := filepath.Join(dtos.UploadDir, "event_"+eid, "userID_"+uid, mediaType)
		if err := os.MkdirAll(userDir, 0755); err != nil {
			errors = append(errors, fmt.Sprintf("ไฟล์ %d: ไม่สามารถสร้างไดเรกทอรีได้", i+1))
			continue
		}

		ext := filepath.Ext(file.Filename)
		id := uuid.New().String()
		newFilename := fmt.Sprintf("%s%s", id, ext)
		savePath := filepath.Join(userDir, newFilename)

		if err := c.SaveFile(file, savePath); err != nil {
			errors = append(errors, fmt.Sprintf("ไฟล์ %d: ไม่สามารถบันทึกได้", i+1))
			continue
		}

		fileURL := fmt.Sprintf("%s/uploads/event_%s/userID_%s/%s/%s", dtos.BaseURL, eid, uid, mediaType, newFilename)
		currentTime := time.Now()

		media := models.Media{
			UserID:     uint(userID),
			EventID:    uint(eventID),
			FileURL:    fileURL,
			FileType:   fileType,
			DetailTime: currentTime,
			UploadTime: currentTime,
		}

		if database.DB == nil {
			errors = append(errors, fmt.Sprintf("ไฟล์ %d: ไม่มีการเชื่อมต่อฐานข้อมูล", i+1))
			os.Remove(savePath)
			continue
		}

		result := database.DB.Create(&media)
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("ไฟล์ %d: ไม่สามารถบันทึกข้อมูลลงฐานข้อมูลได้ - %v", i+1, result.Error))
			os.Remove(savePath)
			continue
		}

		response := ImageResponse{
			ID:       id,
			Filename: newFilename,
			URL:      fileURL,
			Size:     file.Size,
			UploadAt: currentTime.Format("2006-01-02 15:04:05"),
		}

		if fileType == 1 {
			imageResponses = append(imageResponses, response)
		} else if fileType == 2 {
			videoResponses = append(videoResponses, response)
		}
	}

	result := fiber.Map{
		"images":  imageResponses,
		"videos":  videoResponses,
		"สำเร็จ":  len(imageResponses) + len(videoResponses),
		"ล้มเหลว": errors,
	}

	return c.JSON(result)
}

func IsValidMediaType(contentType string) bool {
	validTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp",
		"video/mp4", "video/quicktime", "video/x-msvideo", "video/x-matroska", // mp4, mov, avi, mkv
	}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}
