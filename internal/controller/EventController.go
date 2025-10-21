package controller

import (
	// "fmt"
	"strconv"

	// "time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"

	// "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
)

func GetAllEvent(c *fiber.Ctx) error {
	response, err := services.EventGetAll()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่พบอีเวนต์",
		})
	}

	return c.JSON(fiber.Map{
		"data": response,
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
			"message": "Unable to process",
		})
	}

	response, err := services.EventCreate(&event)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var req request.JoinRequest

	req.EID = strconv.FormatUint(uint64(response.ID), 10)
	req.UID = strconv.FormatUint(uint64(response.UserID), 10)

	_, err = services.JoinCreate(&req)
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

func EventDelete(c *fiber.Ctx) error {
	eid := c.Params("eid")

	_, err := services.EventDelete(eid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ลบอีเวนต์สำเร็จ",
	})
}

func GetEventsWithAttendees(c *fiber.Ctx) error {
	uid := c.Params("uid")

	response, err := services.GetEventsWithAttendees(uid)
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

func GetEventMainProfile(c *fiber.Ctx) error {
	uid := c.Params("uid")

	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด UserId",
		})
	}

	response, err := services.GetEventMainProfile(uid)
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

func EventGetListJoinUser(c *fiber.Ctx) error {
	eid := c.Params("eid")
	uid := c.Params("uid")

	if eid == "" || uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด EventID และ UserID",
		})
	}

	response, err := services.EventGetListJoinUser(eid, uid)
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

func EventGetFavorites(c *fiber.Ctx) error {
	uid := c.Params("uid")

	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด UserID",
		})
	}

	response, err := services.EventGetFavorites(uid)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลเสร็จสิ้น",
		"data":    response,
	})
}

func GetEventCalendar(c *fiber.Ctx) error {
	uid := c.Params("uid")

	response, err := services.GetEventCalendar(uid)
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

func GetEventJoin(c *fiber.Ctx) error {
	uid := c.Params("uid")

	response, err := services.GetEventJoin(uid)
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

func EventCheckJoinUser(c *fiber.Ctx) error {
	eid := c.Params("eid")
	uid := c.Params("uid")

	if eid == "" || uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ขาด EventID และ UserID",
		})
	}

	response, err := services.EventCheckJoinUser(eid, uid)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": response,
	})
}

func EventFavorite(c *fiber.Ctx) error {
	var req request.FavoriteReq

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unable to process",
		})
	}

	msg, err := services.EventFavorite(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  true,
		"message": msg,
	})
}

// func EventGetFavorite(c *fiber.Ctx) error {
// 	uid := c.Params("uid")

// 	if uid == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "ไม่สามารถดำเนินการได้",
// 		})
// 	}

// 	res, err := services.EventGetFavorite(uid)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"message": err.Error(),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"data":   res,
// 		"mssage": "ดึงข้อมูลสำเร็จ",
// 	})
// }

// func EventGetFavoriteWS(ws *websocket.Conn, c *fiber.Ctx) {
// 	uid := c.Params("uid") // รับ user id จาก path parameter
// 	if uid == "" {
// 		ws.WriteMessage(1, []byte(`{"error":"uid is required"}`))
// 		return
// 	}

// 	ticker := time.NewTicker(5 * time.Second)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		res, err := services.EventGetFavorite(uid)
// 		if err != nil {
// 			ws.WriteMessage(1, fmt.Appendf(nil, `{"error":"%s"}`, err.Error()))
// 			continue
// 		}

// 		data := fmt.Sprintf(`{"message":"อัปเดตข้อมูล","data":%v}`, res)
// 		if err := ws.WriteMessage(1, []byte(data)); err != nil {
// 			fmt.Println("WebSocket write error:", err)
// 			return
// 		}
// 	}
// }
