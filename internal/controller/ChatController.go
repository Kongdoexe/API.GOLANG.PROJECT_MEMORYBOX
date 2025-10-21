package controller

import (
	"fmt"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/services"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
)

type HubType struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
}

var Hub = HubType{
	Clients:    make(map[*websocket.Conn]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *websocket.Conn),
	Unregister: make(chan *websocket.Conn),
}

func (h *HubType) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			fmt.Println("Client connected:", client.RemoteAddr())
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
				fmt.Println("Client disconnected")
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				if err := client.WriteMessage(1, message); err != nil {
					fmt.Println("Send error:", err)
					h.Unregister <- client
				}
			}
		}
	}
}

func InsertMessage(c *fiber.Ctx) error {
	var req request.InsertMessage
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	err := services.InsertMessage(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ส่งข้อความสำเร็จ",
	})
}

func GetMessage(c *fiber.Ctx) error {
	var req request.GetMessage
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ไม่สามารถดำเนินการได้",
		})
	}

	response, err := services.GetMessage(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	responseDataUser, _ := repositories.UserFindByID(req.Uid)

	return c.JSON(fiber.Map{
		"message":  "ดึงข้อมูลสำเร็จ",
		"data":     response,
		"dataUser": responseDataUser,
	})
}
