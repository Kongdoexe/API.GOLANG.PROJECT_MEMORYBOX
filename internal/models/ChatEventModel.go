package models

import "time"

type ChatMessage struct {
	MessageID uint      `gorm:"primaryKey" json:"message_id"`
	UserID    uint      `json:"user_id"`
	EventID   uint      `json:"event_id"`                 // ใช้เป็น room ID
	Message   string    `gorm:"type:text" json:"message"` // เนื้อหาข้อความ
	ChatDate  time.Time `json:"chat_date"`

	User  User  `gorm:"foreignKey:UserID;references:ID"`
	Event Event `gorm:"foreignKey:EventID;references:ID"`
}

func (ChatMessage) TableName() string {
	return "chatmessage"
}
