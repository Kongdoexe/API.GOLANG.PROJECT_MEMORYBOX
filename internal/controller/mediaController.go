package controller

import (
	"strconv"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
)

func GetMediaByID(c *fiber.Ctx) error {
	id := c.Params("id")
	mediaId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "ไม่พบมีเดีย ID"})
	}

	media, err := services.GetMediaInfo(uint(mediaId))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "ไม่พบสื่อ"})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    media,
	})
}
