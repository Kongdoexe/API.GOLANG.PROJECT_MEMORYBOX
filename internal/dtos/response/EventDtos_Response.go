package response

import (
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

type GetallEvent struct {
	EventId         int      `json:"event_id"`
	EventTitle      string   `json:"event_title"`
	EventDate       string   `json:"event_date"`
	EventLocation   string   `json:"event_location"`
	EventLatitude   float64  `json:"event_latitude"`
	EventLongitude  float64  `json:"event_longitude"`
	EventImage      string   `json:"event_image"`
	EventCategory   int      `json:"event_category"`
	AttendeeCount   int      `json:"attendee_count"`
	AttendeeAvatars []string `json:"attendee_avatars"`
	IsFavorite      bool     `json:"isFavorite"`
	IsJoin          bool     `json:"isJoin"`
}

type WsMessageRequest struct {
	EventId string `json:"event_id"`
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

type WSMessage struct {
	ActionEvent string        `json:"action_event"`
	Data        WSChatMessage `json:"data"`
}

type WSChatMessage struct {
	UserID    string `json:"userId"`
	EventId   string `json:"eventId"` // RoomId
	DataUser  *models.User
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type EventDetail struct {
	EventId            int64     `json:"event_id"`
	EventTitle         string    `json:"event_title"`
	EventStartDateTime time.Time `json:"event_start_date_time"`
	EventEndDateTime   time.Time `json:"event_end_date_time"`
	EventLocationName  string    `json:"event_location_name"`
	EventImage         string    `json:"event_image"`
	EventAttendeeCount int       `json:"event_attend_count"`
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

type EventGetFavorite struct {
	EventId       int    `json:"event_id"`
	EventTitle    string `json:"event_title"`
	EventDeatil   string `json:"event_detail"`
	EventStart    string `json:"event_start_date"`
	EventEnd      string `json:"event_end_date"`
	EventLocation string `json:"event_location"`
	EventImage    string `json:"event_image"`
}

type EventGetMainProfile struct {
	EventId       int    `json:"id"`
	EventImage    string `json:"image"`
	EventTitle    string `json:"title"`
	EventDate     string `json:"datetime"`
	EventLocation string `json:"loction"`
}

type EventGetFavorites struct {
	ID             int       `json:"id"`
	EventName      string    `json:"title"`
	EventImage     string    `json:"image"`
	TypeID         int       `json:"type_id"`
	MaxMedia       int       `json:"max_media"`
	EventStartTime time.Time `json:"dateT"`
	EventDate      string    `json:"date"`
	FavoriteDate   string    `json:"favorite_date"`
}

type EventGetListJoin struct {
	ID        uint   `gorm:"primaryKey;column:user_id" json:"id"`
	UserImage string `json:"user_image" gorm:"column:user_image"`
	Name      string `json:"name" gorm:"size:255;column:name"`
	Email     string `json:"email" gorm:"unique;size:255;not null;column:email"`
	Phone     string `json:"phone" gorm:"column:phone"`
	Status    int    `json:"status"`
}
