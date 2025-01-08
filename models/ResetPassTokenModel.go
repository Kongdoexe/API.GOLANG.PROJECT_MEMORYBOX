package models

import "time"

type ResetPassToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;column:token_id"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token" gorm:"size:500"`
	Expire    time.Time `json:"expire"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
