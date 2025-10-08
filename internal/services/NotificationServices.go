package services

import (
	"errors"
	"fmt"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
)

func InsertTokenNotification(req request.RequestTokenNotification) (bool, error) {
	_, err := repositories.UserFindByID(req.UserID)
	if err != nil {
		return false, errors.New("User not found")
	}

	status, err := repositories.InsertTokenNotification(req)
	if err != nil {
		return false, errors.New("You can update the information, try again.")
	}

	return status, nil
}

func SendNotification(req request.RequestSendNotificaiton) (bool, error) {
	response, err := repositories.GetUserTokenNotificationInEvent(req.EventID)
	if err != nil {
		return false, errors.New("Event Not Found")
	}

	fmt.Print(response)

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

	return true, nil
}
