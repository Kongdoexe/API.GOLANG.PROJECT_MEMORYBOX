package response

import "time"

type GetallEvent struct {
	EventId         int      `json:"event_id"`
	EventTitle      string   `json:"event_title"`
	EventDate       string   `json:"event_date"`
	EventLocation   string   `json:"event_location"`
	EventImage      string   `json:"event_image"`
	AttendeeCount   int      `json:"attendee_count"`
	AttendeeAvatars []string `json:"attendee_avatars"`
}

type DetailEvent struct {
	UserID            int64  `json:""user_id`
	UserImage         string `json:"organizerAvatar" gorm:"column:user_image"`
	Name              string `json:"organizer" gorm:"size:500"`
	EventImage        string `json:"event_image" gorm:"column:event_image"`
	EventName         string `json:"event_name" gorm:"column:event_name;size:500"`
	EventDetail       string `json:"event_detail" gorm:"column:event_detail"`
	EventLocationName string `json:"event_location_name" gorm:"column:event_location_name;size:500"`
	EventDate         string `json:"event_date"`
	EventTimeDisplay  string `json:"event_time_display"`
	MaxMedia          int    `json:"max_media" gorm:"column:max_media"`
}

type EventcreateResponse struct {
	UserID             int64     `json:"user_id"`
	EventImage         string    `json:"event_image"`
	EventName          string    `json:"event_name"`
	EventDetail        string    `json:"event_detail"`
	EventLocationName  string    `json:"event_location_name"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	EventStartDateTime time.Time `json:"event_start_date_time"`
	EventEndDateTime   time.Time `json:"event_end_date_time"`
	MaxMedia           int64     `json:"max_media"`
	TypeID             int64     `json:"type_id"`
	EventStatus        int64     `json:"event_status"`
	AccessModifiers    int64     `json:"access_modifiers"`
	EventQrcode        string    `json:"event_qrcode"`
}
