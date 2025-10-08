package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/firebase"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/routers"
	"firebase.google.com/go/v4/messaging"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

const (
	uploadDir = "./uploads"
	maxSize   = 10 * 1024 * 1024
)

func init() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error in loading .env file.")
	}

	database.InitDB()
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

var hub = Hub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			fmt.Println("Client connected:", client.RemoteAddr())
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				fmt.Println("Client disconnected")
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if err := client.WriteMessage(1, message); err != nil {
					fmt.Println("Send error:", err)
					h.unregister <- client
				}
			}
		}
	}
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

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		hub.register <- c
		defer func() { hub.unregister <- c }()

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	app.Post("/api/sendmessage", func(c *fiber.Ctx) error {
		var messageData struct {
			Username string `json:"username"`
			Message  string `json:"message"`
			RoomID   string `json:"room_id"`
		}

		if err := c.BodyParser(&messageData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "Invalid request format",
			})
		}

		// สร้าง response สำหรับ websocket
		msg := response.WSMessage{
			Event: "new_message",
			UID:   messageData.RoomID,
			Data: response.WSChatMessage{
				Username:  messageData.Username,
				Message:   messageData.Message,
				RoomID:    messageData.RoomID,
				Timestamp: time.Time{},
			},
		}

		jsonBytes, _ := json.Marshal(msg)
		hub.broadcast <- jsonBytes

		return c.JSON(fiber.Map{
			"message": "ส่งข้อความสำเร็จ",
			"status":  "success",
		})
	})

	// API สำหรับจำลองข้อความจากระบบหรือบอท
	app.Post("/api/simulatemessage", func(c *fiber.Ctx) error {
		// ข้อความที่จำลอง
		simulatedMessages := []string{
			"สวัสดีทุกคน! 👋",
			"วันนี้อากาศดีนะครับ",
			"มีใครอยู่ไหม?",
			"ช่วงนี้ยุ่งมากเลย",
			"เย็นนี้ไปกินข้าวกันไหม?",
			"ระบบทำงานปกติครับ ✅",
			"ขอบคุณสำหรับการใช้งาน",
			"หากมีปัญหาแจ้งได้นะครับ",
		}

		// สุ่มเลือกข้อความ
		randomMessage := simulatedMessages[rand.Intn(len(simulatedMessages))]

		// ชื่อผู้ส่งจำลอง
		simulatedUsers := []string{"System", "Admin", "Bot", "Helper", "Manager"}
		randomUser := simulatedUsers[rand.Intn(len(simulatedUsers))]

		// สร้าง response สำหรับ websocket
		msg := response.WSMessage{
			Event: "new_message",
			UID:   "general", // room_id เดียวกัน\
			Data: response.WSChatMessage{
				Username:  randomUser,
				Message:   randomMessage,
				RoomID:    "general",
				Timestamp: time.Now(),
			},
		}

		jsonBytes, _ := json.Marshal(msg)
		hub.broadcast <- jsonBytes

		return c.JSON(fiber.Map{
			"message": "ส่งข้อความจำลองสำเร็จ",
			"status":  "success",
			"data": fiber.Map{
				"username": randomUser,
				"message":  randomMessage,
			},
		})
	})

	// API สำหรับส่งข้อความอัตโนมัติทุก ๆ 30 วินาที
	app.Get("/api/startauto", func(c *fiber.Ctx) error {
		// ใช้ goroutine สำหรับส่งข้อความอัตโนมัติ
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			messages := []string{
				"ระบบทำงานปกติ 🟢",
				"มีผู้ใช้งานใหม่เข้าร่วม",
				"อย่าลืมบันทึกข้อมูลนะครับ",
				"เวลาพัก ☕",
				"ระบบอัพเดตเสร็จแล้ว",
			}

			for range ticker.C {
				randomMessage := messages[rand.Intn(len(messages))]

				msg := response.WSMessage{
					Event: "new_message",
					UID:   "general",
					Data: response.WSChatMessage{
						Username:  "AutoBot",
						Message:   randomMessage,
						RoomID:    "general",
						Timestamp: time.Now(),
					},
				}

				jsonBytes, _ := json.Marshal(msg)
				hub.broadcast <- jsonBytes
			}
		}()

		return c.JSON(fiber.Map{
			"message": "เริ่มส่งข้อความอัตโนมัติแล้ว",
			"status":  "success",
		})
	})

	app.Post("/send", func(c *fiber.Ctx) error {
		type Request struct {
			Token string `json:"token"`
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		msg := &messaging.Message{
			Token: req.Token,
			Notification: &messaging.Notification{
				Title: req.Title,
				Body:  req.Body,
			},
		}

		// ใช้ client จาก firebase package
		response, err := firebase.MessagingClient.Send(context.Background(), msg)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("error sending message: %v", err))
		}

		return c.JSON(fiber.Map{
			"success":   true,
			"messageID": response,
		})
	})

	go hub.run()

	defer sqlDb.Close()
	routers.SetupRouter(app)

	app.Listen(":5000")
}
