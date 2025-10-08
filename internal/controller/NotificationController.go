package controller

import (
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
)

func InsertTokenNotification(c *fiber.Ctx) error {
	var req request.RequestTokenNotification

	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unable to process",
		})
	}

	status, err := services.InsertTokenNotification(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Data updated successfully",
		"status":  status,
	})
}

func SendNotification(c *fiber.Ctx) error {
	var req request.RequestSendNotificaiton
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	_, err := services.SendNotification(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// msg := &messaging.Message{
	//     Token: token,
	//     Notification: &messaging.Notification{
	//         Title: "MemoryBox",
	//         Body:  "คุณมีการแจ้งเตือนใหม่",
	//     },
	//     Android: &messaging.AndroidConfig{
	//         CollapseKey: "memorybox_event", // <--- ใส่ collapse_key ตรงนี้
	//         Priority:    "high",
	//     },
	// }

	// // ใช้ client จาก firebase package
	// response, err := firebase.MessagingClient.Send(context.Background(), msg)
	// if err != nil {
	// 	return c.Status(500).SendString(fmt.Sprintf("error sending message: %v", err))
	// }

	return c.JSON(fiber.Map{
		"success": true,
		// "messageID": response,
	})
}
