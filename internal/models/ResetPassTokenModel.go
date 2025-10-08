package models

import "time"

// import "time"

// type ResetPassToken struct {
// 	ID     uint      `json:"id" gorm:"primaryKey;column:token_id"`
// 	UserID uint      `json:"user_id"`
// 	Token  string    `json:"token" gorm:"size:500"`
// 	Expire time.Time `json:"expire"`
// }

type ResetPassToken struct {
	ID     uint      `json:"id" gorm:"primaryKey;column:token_id"`
	UserID uint      `json:"user_id" gorm:"column:user_id"`
	Token  string    `json:"token" gorm:"size:255; column:token"`
	Expire time.Time `json:"expire" gorm:"column:expire"`
}

func (ResetPassToken) TableName() string {
	return "resetpasstoken"
}
