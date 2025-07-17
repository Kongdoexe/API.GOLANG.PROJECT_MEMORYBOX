package models

import "time"

type Favorite struct {
	ID           uint      `json:"id" gorm:"primaryKey;column:favorite_id"`
	UserID       uint      `json:"user_id"`
	EventID      uint      `json:"event_id"`
	FavoriteDate time.Time `json:"favorite_date"`
}
