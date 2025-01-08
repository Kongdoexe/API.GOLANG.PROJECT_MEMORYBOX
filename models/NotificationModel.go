package models

import "time"

type Notification struct {
	ID               uint      `json:"id" gorm:"primaryKey;column:notification_id"`
	UserID           uint      `json:"user_id"`
	EventID          uint      `json:"event_id"`
	NotificationTime time.Time `json:"notification_time" gorm:"column:notification_time"`
	NotificationType string    `json:"notification_type" gorm:"column:notification_type;size:500"`
	IsRead           bool      `json:"is_read"`
}
