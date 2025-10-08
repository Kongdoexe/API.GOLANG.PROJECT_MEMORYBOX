package models

import "time"

// import "time"

// type Media struct {
// 	ID         uint      `json:"id" gorm:"primaryKey;column:media_id"`
// 	UserID     uint      `json:"user_id"`
// 	EventID    uint      `json:"event_id"`
// 	FileURL    string    `json:"file_url"`
// 	FileType   int       `json:"file_type"`
// 	DetailTime time.Time `json:"detail_time"`
// 	UploadTime time.Time `json:"upload_time"`
// }

type Media struct {
	MediaID    uint      `gorm:"primaryKey" json:"media_id"`
	UserID     uint      `json:"user_id"`
	EventID    uint      `json:"event_id"`
	FileURL    string    `json:"file_url"`
	FileType   int       `json:"file_type"`
	DetailTime time.Time `json:"detail_time"`
	UploadTime time.Time `json:"upload_time"`

	User  User  `gorm:"foreignKey:UserID;references:ID"`
	Event Event `gorm:"foreignKey:EventID;references:ID"`
}

func (Media) TableName() string {
	return "media"
}

