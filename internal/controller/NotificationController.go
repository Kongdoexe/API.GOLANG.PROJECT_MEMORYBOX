package controller

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
)

type NotiWithEventOwner struct {
	models.Notification `json:",inline"`
	EventOwnerID        uint `json:"event_owner_id,omitempty"`
}

func CheckUpCommingEvents() {
	log.Println("เริ่มการตรวจจับอีเวนต์ที่จะเริ่ม")

	now := time.Now()
	next24h := now.Add(24 * time.Hour)

	events, err := repositories.FindEventBetween(now, next24h)
	if err != nil {
		log.Println("ไม่พบอีเวนต์", err)
		return
	}
	if len(events) == 0 {
		log.Println("ไม่มีอีเวนต์ที่จะเริ่มภายใน 24 ชม.")
		return
	}

	for _, e := range events {
		// เปลี่ยนการตั้งค่า สำหรับการตรวจเช็คว่า แจ้งเตือนไปแล้วหรือยัง
		e.IsNotiOneDay = 1
		if err := repositories.EventChangeNotiDayOne(&e); err != nil {
			log.Println("ไม่สามารถบันทึกอีเวนต์ได้:", err)
			return
		}
		if err := services.SendNotificationEvent24Hr(strconv.Itoa(int(e.ID))); err != nil {
			log.Println("ไม่สามารถส่งการแจ้งเตือน 24Hr ได้:", err)
			return
		}
		log.Printf("บันทึกข้อมูลอีเวนต์ '%s' และแจ้งเตือนสำเร็จแล้ว\n", e.EventName)
	}
}

func InsertTokenNotification(c *fiber.Ctx) error {
	var req request.RequestTokenNotification
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unable to process",
		})
	}

	status, err := services.InsertTokenNotification(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Data updated successfully",
		"status":  status,
	})
}

// body: { token|topic, title, body, event_id, subtype?("invited"|"joined"|"reminder") }
func SendNotificationEvent(c *fiber.Ctx) error {
	var req request.RequestSendNotiEvent
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	fmt.Print(req)

	if err := services.SendNotificationEvent(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true})
}

// POST /notify/chat/:eid
func SendNotificationChat(c *fiber.Ctx) error {
	var req request.RequestSendNotiChat
	if err := c.BodyParser(&req); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	if err := services.SendNotificationChat(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

func GetUserNotification(c *fiber.Ctx) error {
	uid := c.Params("uid")
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด UserID",
		})
	}

	notis, events, err := services.GetUserNotification(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	evByID := make(map[uint]models.Event, len(events))
	for _, ev := range events {
		evByID[ev.ID] = ev
	}

	out := make([]NotiWithEventOwner, 0, len(notis))
	for _, n := range notis {
		item := NotiWithEventOwner{
			Notification: n,
		}

		if n.EventID != nil {
			if ev, ok := evByID[uint(*n.EventID)]; ok {
				ownerID := ev.UserID
				if ownerID == 0 && ev.UserID != 0 {
					ownerID = ev.UserID
				}
				item.EventOwnerID = ownerID
			}
		}

		out = append(out, item)
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    out,
	})
}

func UpdateIsReadNotification(c *fiber.Ctx) error {
	var nid = c.Params("nid")
	if nid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด NotificationID",
		})
	}

	if err := services.UpdateIsReadNotification(nid); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "เปลี่ยนสำเร็จ",
	})
}
