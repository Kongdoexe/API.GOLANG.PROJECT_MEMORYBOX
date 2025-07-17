package controller

import (
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
)

func JoinEvent(c *fiber.Ctx) error {
	var req request.JoinRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	response, err := services.JoinCreate(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "เข้าร่วมอีเวนต์สำเร็จ",
		"data":    response,
	})
}
