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
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

type ImageResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	UploadAt string `json:"upload_at"`
}

// เพิ่มรูป / วิดีโอ
func readExifDateTime(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	// Register Canon and Nikon makernote parsers
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(file)
	if err != nil {
		return time.Time{}, err
	}

	// ลองอ่านจาก DateTime tag ต่างๆ
	dateTimeTags := []exif.FieldName{
		exif.DateTimeOriginal,  // วันที่ถ่ายจริง
		exif.DateTime,          // วันที่แก้ไข
		exif.DateTimeDigitized, // วันที่ digitize
	}

	for range dateTimeTags {
		if dateTime, err := x.DateTime(); err == nil && !dateTime.IsZero() {
			return dateTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("no valid datetime found in EXIF")
}

func readExifGPS(filePath string) (lat, lng float64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		return 0, 0, err
	}

	lat, lng, err = x.LatLong()
	if err != nil {
		return 0, 0, err
	}

	return lat, lng, nil
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

	if err := services.CheckDateEvent(uint(eventID)); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
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

	// Get mediaTimes from form if present (e.g., from Flutter)
	mediaTimes := []string{}
	if mtimes, ok := form.Value["mediaTime"]; ok {
		mediaTimes = mtimes
	}

	var imageResponses []ImageResponse
	var videoResponses []ImageResponse
	var errors []string

	for i, file := range files {
		contentType := file.Header.Get("Content-Type")

		if err = services.CheckMaxMedia(uint(eventID), uint(userID)); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

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

		// กำหนดค่า DetailTime
		var detailTime time.Time
		var latitude, longitude float64

		// ถ้าเป็นรูปภาพ ให้ลองอ่าน EXIF
		if fileType == 1 { // Image file
			fmt.Printf("Attempting to read EXIF from file %d: %s\n", i+1, newFilename)

			// อ่านวันที่จาก EXIF
			if exifDateTime, err := readExifDateTime(savePath); err == nil {
				detailTime = exifDateTime
				fmt.Printf("✓ EXIF DateTime found for file %d: %s\n", i+1, exifDateTime.Format(time.RFC3339))
			} else {
				fmt.Printf("⚠ No EXIF DateTime for file %d: %v\n", i+1, err)
				// ลองใช้ mediaTime จาก Flutter
				if i < len(mediaTimes) && mediaTimes[i] != "" && mediaTimes[i] != "unknown" {
					if parsedTime, parseErr := time.Parse(time.RFC3339, mediaTimes[i]); parseErr == nil {
						detailTime = parsedTime
						fmt.Printf("✓ Using Flutter mediaTime for file %d: %s\n", i+1, parsedTime.Format(time.RFC3339))
					} else {
						detailTime = currentTime
						fmt.Printf("⚠ Using upload time for file %d\n", i+1)
					}
				} else {
					detailTime = currentTime
					fmt.Printf("⚠ Using upload time for file %d (no mediaTime)\n", i+1)
				}
			}

			// อ่าน GPS coordinates (ถ้าต้องการ)
			if lat, lng, err := readExifGPS(savePath); err == nil {
				latitude = lat
				longitude = lng
				fmt.Printf("✓ GPS coordinates found for file %d: %.6f, %.6f\n", i+1, lat, lng)
			} else {
				fmt.Printf("⚠ No GPS data for file %d: %v\n", i+1, err)
			}

		} else if fileType == 2 { // Video file
			if i < len(mediaTimes) && mediaTimes[i] != "" && mediaTimes[i] != "unknown" {
				if parsedTime, err := time.Parse(time.RFC3339, mediaTimes[i]); err == nil {
					detailTime = parsedTime
					fmt.Printf("✓ Using Flutter mediaTime for video %d: %s\n", i+1, parsedTime.Format(time.RFC3339))
				} else {
					detailTime = currentTime
					fmt.Printf("⚠ Using upload time for video %d\n", i+1)
				}
			} else {
				detailTime = currentTime
				fmt.Printf("⚠ Using upload time for video %d (no mediaTime)\n", i+1)
			}
		} else {
			detailTime = currentTime
		}

		// สร้าง Media object (เพิ่ม GPS fields ถ้า model รองรับ)
		media := models.Media{
			UserID:     uint(userID),
			EventID:    uint(eventID),
			FileURL:    fileURL,
			FileType:   fileType,
			DetailTime: detailTime,
			UploadTime: currentTime,
			// เพิ่ม GPS fields ถ้า model มี
			// Latitude:   latitude,
			// Longitude:  longitude,
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

		fmt.Printf("✅ Successfully processed file %d: %s\n", i+1, newFilename)
		fmt.Printf("   - DetailTime: %s\n", detailTime.Format(time.RFC3339))
		if latitude != 0 && longitude != 0 {
			fmt.Printf("   - GPS: %.6f, %.6f\n", latitude, longitude)
		}
	}

	result := fiber.Map{
		"images":  imageResponses,
		"videos":  videoResponses,
		"success": len(imageResponses) + len(videoResponses),
		"fails":   errors,
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

// ลบรูป / วิดีโอ
func buildDiskPath(eid, uid, subdir, filename string) string {
	return filepath.Join(
		dtos.UploadDir,
		"event_"+eid,
		"userID_"+uid,
		subdir,
		filename,
	)
}

func guessSubdirFromURL(url string, fileType int) string {
	if strings.Contains(url, "/images/") || fileType == 1 {
		return "images"
	}
	if strings.Contains(url, "/videos/") || fileType == 2 {
		return "videos"
	}
	return "others"
}

func removeIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		// มีไฟล์ → ลบ
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func DeleteMediaByID(c *fiber.Ctx) error {
	var req request.DeleteMedia
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	// ตรวจสอบเลข
	if _, err := strconv.ParseUint(req.Eid, 10, 32); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "EventID ไม่ถูกต้อง"})
	}
	if _, err := strconv.ParseUint(req.Uid, 10, 32); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "UserID ไม่ถูกต้อง"})
	}
	midUint, err := strconv.ParseUint(req.Mid, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "MediaID ไม่ถูกต้อง"})
	}

	// อ่านเรคคอร์ดก่อน เพื่อรู้ทั้ง file_url / filetype / filename
	var media models.Media
	if err := database.DB.
		Where("media_id = ? AND event_id = ? AND user_id = ?", midUint, req.Eid, req.Uid).
		First(&media).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบสื่อในระบบ"})
	}

	// ดึง filename จาก URL (หลังสุดของ path)
	parts := strings.Split(media.FileURL, "/")
	filename := parts[len(parts)-1]
	subdir := guessSubdirFromURL(media.FileURL, media.FileType)
	diskPath := buildDiskPath(req.Eid, req.Uid, subdir, filename)

	// ลบไฟล์บนดิสก์
	if err := removeIfExists(diskPath); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("ลบไฟล์ไม่สำเร็จ: %v", err),
			"path":  diskPath,
		})
	}

	// ลบเรคคอร์ด
	if err := database.DB.Delete(&media).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ลบเรคคอร์ดในฐานข้อมูลไม่สำเร็จ",
		})
	}

	return c.JSON(fiber.Map{
		"status":  true,
		"message": "ลบสื่อสำเร็จ",
	})
}
