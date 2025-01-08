package models

import "time"

type Event struct {
	ID                 uint       `json:"id" gorm:"primaryKey;column:event_id"`
	UserID             uint       `json:"user_id"`
	EventImage         string     `json:"event_image" gorm:"column:event_image"`
	EventName          string     `json:"event_name" gorm:"column:event_name;size:500"`
	EventDetail        string     `json:"event_detail" gorm:"column:event_detail"`
	EventLocationName  string     `json:"event_location_name" gorm:"column:event_location_name;size:500"`
	Latitude           float64    `json:"latitude"`
	Longitude          float64    `json:"longitude"`
	EventStartDateTime time.Time  `json:"event_start_date_time" gorm:"column:event_start_date_time"` // วันและเวลาในการเริ่มอีเวนต์
	EventEndDateTime   time.Time  `json:"event_end_date_time" gorm:"column:event_end_date_time"`     // วันและเวลาการจบของอีเวนต์
	MaxMedia           int        `json:"max_media" gorm:"column:max_media"`                         // จำนวนมีเดียสูงสุดที่อัปได้ต่อคน
	TypeName           int        `json:"type_name" gorm:"column:type_name"`                         // ประเภทของอีเวนต์
	EventStatus        int        `json:"event_status" gorm:"column:event_status"`                   // สถานนะของอีเวนต์ ว่าจบไปแล้ว หรือยังถ้า 1 ยังไม่จบ 0 คือจบแล้ว ถ้ายังไม่จบยังสามารถให้คนเข้าร่วมได้อยู่
	AccessModifiers    int        `json:"access_modifiers" gorm:"colum:access_modifiers"`            // การตั้งค่า 1 รับอัตโนมัติ 2 รับแบบให้ผู้สร้างกดที่ละคน หรือทีละมากๆ
	EventQRCode        string     `json:"event_qrcode" gorm:"column:event_qrcode"`                   // ลิงค์รูปภาพ QRCode ใช้ในการลิงค์เข้าไปที่อีเวนต์
	Joins              []Join     `json:"joins" gorm:"foreignKey:EventID;onDelete:CASCADE"`
	Favorites          []Favorite `json:"favorites" gorm:"foreignKey:EventID;onDelete:CASCADE"`
	Media              []Media    `json:"media" gorm:"foreignKey:EventID;onDelete:CASCADE"`
}
