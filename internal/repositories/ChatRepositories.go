package repositories

import (
	"strconv"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

func InsertMessage(chat *models.ChatMessage) error {
	return database.DB.Create(chat).Error
}

func FindMessageWithUser(messageID uint) (*models.ChatMessage, error) {
	var cm models.ChatMessage
	if err := database.DB.Preload("User").
		Where("message_id = ?", messageID).
		First(&cm).Error; err != nil {
		return nil, err
	}
	return &cm, nil
}

func GetMessage(req request.GetMessage) (*[]models.ChatMessage, error) {
	var list []models.ChatMessage
	eid, err := strconv.ParseUint(req.Eid, 10, 32)
	if err != nil {
		return nil, err
	}
	if err := database.DB.
		Preload("User").
		Where("event_id = ?", uint(eid)).
		Order("chat_date ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return &list, nil
}
