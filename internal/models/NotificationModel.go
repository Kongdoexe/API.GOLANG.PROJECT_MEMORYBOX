package models

import "time"

// import "time"

// type Notification struct {
// 	ID               uint      `json:"id" gorm:"primaryKey;column:notification_id"`
// 	UserID           uint      `json:"user_id"`
// 	EventID          uint      `json:"event_id"`
// 	NotificationTime time.Time `json:"notification_time" gorm:"column:notification_time"`
// 	NotificationType string    `json:"notification_type" gorm:"column:notification_type;size:500"` // 1 System, 2 Message, 3 Invite
// 	IsRead           bool      `json:"is_read"`
// }

type Notification struct {
	NotificationID   uint      `gorm:"primaryKey" json:"notification_id"`
	UserID           int       `json:"user_id"`
	EventID          *int      `json:"event_id"`
	Title            string    `json:"title"`
	Detail           string    `json:"detail"`
	NotificationType int       `json:"notification_type"` // 1 System, 2 Message, 3 Invite
	NotificationTime time.Time `json:"notification_time"`
	IsRead           bool      `json:"is_read"`
}

func (Notification) TableName() string {
	return "notification"
}
