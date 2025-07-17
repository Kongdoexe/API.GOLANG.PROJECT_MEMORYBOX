package models

import "time"

type Media struct {
	ID         uint      `json:"id" gorm:"primaryKey;column:media_id"`
	UserID     uint      `json:"user_id"`
	EventID    uint      `json:"event_id"`
	FileURL    string    `json:"file_url"`
	FileType   int       `json:"file_type"`
	DetailTime time.Time `json:"detail_time"`
	UploadTime time.Time `json:"upload_time"`
}
