package controller

import (
	"strconv"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
)

func GetAllEvent(c *fiber.Ctx) error {
	response, err := services.EventGetAll()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    response,
	})
}

func EventUploadImageCover(c *fiber.Ctx) error {
	eid := c.Params("eid")

	if eid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด EventID",
		})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่พบไฟล์ที่อัปโหลด",
		})
	}

	imageurl, err := services.EventUploadImageCover(file, eid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":   "อัปโหลดสำเร็จ",
		"image_url": imageurl,
	})
}

func EventCreate(c *fiber.Ctx) error {
	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	response, err := services.EventCreate(&event)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "สร้างอีเวนต์สำเร็จ",
		"data":    response,
	})
}

func GetEventsWithAttendees(c *fiber.Ctx) error {
	response, err := services.GetEventsWithAttendees()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    response,
	})
}
func GetEventDetailWithAttendees(c *fiber.Ctx) error {
	eidStr := c.Params("eid")

	if eidStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด EventID",
		})
	}

	eid, err := strconv.Atoi(eidStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "EventID ต้องเป็นตัวเลข",
		})
	}

	response, err := services.GetEventDetailWithAttendees(eid)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    response,
	})
}

func EventGetMediaByID(c *fiber.Ctx) error {
	eid := c.Params("eid")

	if eid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด EventID",
		})
	}

	responseImage, responseVideo, err := services.EventGetMediaByID(eid)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":   "ดึงข้อมูลสำเร็จ",
		"dataImage": responseImage,
		"dataVideo": responseVideo,
	})
}
