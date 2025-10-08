package models

import "time"

// import "time"

// type Event struct {
// 	ID                 uint       `json:"id" gorm:"primaryKey;column:event_id"`
// 	UserID             uint       `json:"user_id"`
// 	EventImage         string     `json:"event_image" gorm:"column:event_image"`
// 	EventName          string     `json:"event_name" gorm:"column:event_name;size:500"`
// 	EventDetail        string     `json:"event_detail" gorm:"column:event_detail"`
// 	EventLocationName  string     `json:"event_location_name" gorm:"column:event_location_name;size:500"`
// 	Latitude           float64    `json:"latitude"`
// 	Longitude          float64    `json:"longitude"`
// 	EventStartDateTime time.Time  `json:"event_start_date_time" gorm:"column:event_start_date_time"` // วันและเวลาในการเริ่มอีเวนต์
// 	EventEndDateTime   time.Time  `json:"event_end_date_time" gorm:"column:event_end_date_time"`     // วันและเวลาการจบของอีเวนต์
// 	MaxMedia           int        `json:"max_media" gorm:"column:max_media"`                         // จำนวนมีเดียสูงสุดที่อัปได้ต่อคน
// 	TypeId             int        `json:"type_id" gorm:"column:type_id"`                             // ประเภทของอีเวนต์
// 	EventStatus        int        `json:"event_status" gorm:"column:event_status"`                   // สถานนะของอีเวนต์ ว่าจบไปแล้ว หรือยังถ้า 1 ยังไม่จบ 0 คือจบแล้ว ถ้ายังไม่จบยังสามารถให้คนเข้าร่วมได้อยู่
// 	AccessModifiers    int        `json:"access_modifiers" gorm:"colum:access_modifiers"`            // การตั้งค่า 1 รับอัตโนมัติ 2 รับแบบให้ผู้สร้างกดที่ละคน หรือทีละมากๆ
// 	EventQRCode        string     `json:"event_qrcode" gorm:"column:event_qrcode"`                   // ลิงค์รูปภาพ QRCode ใช้ในการลิงค์เข้าไปที่อีเวนต์
// 	Joins              []Join     `json:"joins" gorm:"foreignKey:EventID;onDelete:CASCADE"`
// 	Favorites          []Favorite `json:"favorites" gorm:"foreignKey:EventID;onDelete:CASCADE"`
// 	Media              []Media    `json:"media" gorm:"foreignKey:EventID;onDelete:CASCADE"`
// }

type Event struct {
	ID                uint      `gorm:"primaryKey" json:"event_id"`
	UserID            uint      `json:"user_id"`
	EventName         string    `json:"event_name"`
	EventDetail       string    `json:"event_detail"`
	EventImage        string    `json:"event_image"`
	EventLocationName string    `json:"event_location_name"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	EventStartTime    time.Time `json:"event_start_time"`
	EventEndTime      time.Time `json:"event_end_time"`
	MaxMedia          int       `json:"max_media"`
	TypeID            uint      `json:"type_id"`
	EventStatus       int       `json:"event_status"`
	AccessModifiers   int       `json:"access_modifiers"`
	EventQrcode       string    `json:"event_qrcode"`

	Notification []Notification `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Media        []Media        `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (Event) TableName() string {
	return "event"
}
