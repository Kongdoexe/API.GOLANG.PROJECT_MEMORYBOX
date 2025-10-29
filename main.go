package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/firebase"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/controller"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/routers"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const (
	uploadDir = "./uploads"
)

func init() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error in loading .env file.")
	}

	database.InitDB()
}

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 4048 * 1024 * 1024,
	})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	sqlDb, err := database.DB.DB()

	if err != nil {
		panic("Error in sql connection.")
	}

	firebase.InitFirebase()

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Cannot create upload directory:", err)
	}

	app.Static("/uploads", "./uploads")
	app.Static("/public", "./public")
	app.Static("/userImage", "./userImage")

	app.Get("/ws", websocket.New(func(C *websocket.Conn) {
		controller.Hub.Register <- C.Conn
		defer func() { controller.Hub.Unregister <- C.Conn }()

		for {
			_, _, err := C.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	app.Post("/api/sendmessage", func(c *fiber.Ctx) error {
		var messageData response.WsMessageRequest

		if err := c.BodyParser(&messageData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "Invalid request format",
			})
		}

		// ดึงข้อมูล user จากฐานข้อมูล
		responseUserData, _ := repositories.UserFindByID(messageData.UserId)

		//สร้าง message สำหรับ WebSocket
		msg := response.WSMessage{
			ActionEvent: "new_message",
			Data: response.WSChatMessage{
				UserID:    messageData.UserId,
				EventId:   messageData.EventId,
				Message:   messageData.Message,
				Timestamp: time.Now(),
				DataUser:  responseUserData,
			},
		}

		jsonBytes, _ := json.Marshal(msg)
		controller.Hub.Broadcast <- jsonBytes

		reqInsert := &request.InsertMessage{
			Eid: messageData.EventId,
			Uid: messageData.UserId,
			Msg: messageData.Message,
		}

		err := services.InsertMessage(*reqInsert)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "ส่งข้อความสำเร็จ",
			"status":  "success",
		})
	})

	go controller.Hub.Run()

	c := cron.New(
		cron.WithLocation(time.FixedZone("Asia/Bangkok", 7*60*60)),  // แนะนำระบุโซนเวลา
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)), // กันงานค้างทับกัน
	)
	_, err = c.AddFunc("@hourly", controller.CheckUpCommingEvents)
	if err != nil {
		log.Fatal("Failed to add cron job:", err)
	}
	c.Start()
	defer c.Stop()

	defer sqlDb.Close()
	routers.SetupRouter(app)

	app.Listen(":5000")
}
