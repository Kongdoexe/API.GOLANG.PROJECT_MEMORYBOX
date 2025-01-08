package controller

import (
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/models"
	"github.com/gofiber/fiber/v2"
)

type ModelSusAndJoin struct {
	UserID  int `json:"user_id"`
	EventID int `json:"event_id"`
}

func SelectAllEvent(c *fiber.Ctx) error {
	var events []models.Event
	result := database.DBconn.
		Preload("Joins").
		Preload("Favorites").
		Preload("Media").
		Find(&events)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch users data",
			"details": result.Error.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   events,
		"count":  len(events),
	})
}

func IncreaseEvent(c *fiber.Ctx) error {
	var data_event models.Event

	if err := c.BodyParser(&data_event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Failed to parse JSON",
		})
	}

	var member models.User
	if err := database.DBconn.Where("user_id = ?", data_event.UserID).First(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Member not found",
		})
	}

	if err := database.DBconn.Create(&data_event).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to Increase Event",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Event Increased Successfully",
		"data":    data_event,
	})
}

func JoinEvent(c *fiber.Ctx) error {

	var JoinData ModelSusAndJoin
	if err := c.BodyParser(&JoinData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	var member models.User
	if err := database.DBconn.Where("user_id = ?", JoinData.UserID).First(&member).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User Not Found",
		})
	}

	var event models.Event
	if err := database.DBconn.Where("event_id = ?", JoinData.EventID).First(&event).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Event Not Found",
		})
	}

	var JoinCheck models.Join
	if err := database.DBconn.Where("user_id = ? and event_id = ? ", member.ID, event.ID).First(&JoinCheck).Error; err == nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "You are already participating in an event.",
		})
	}

	JoinInsert := models.Join{
		UserID:    member.ID,
		EventID:   event.ID,
		Enable:    false,
		JoinLevel: 1,
	}

	if err := database.DBconn.Create(&JoinInsert).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to join event",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Join Successfully",
	})
}

func SuspendedUserInEvent(c *fiber.Ctx) error {

	var Suspended ModelSusAndJoin
	if err := c.BodyParser(&Suspended); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Suspended User %s",
	})
}
