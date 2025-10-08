package models

import "time"

// import "time"

// type Favorite struct {
// 	ID           uint      `json:"id" gorm:"primaryKey;column:favorite_id"`
// 	UserID       uint      `json:"user_id"`
// 	EventID      uint      `json:"event_id"`
// 	FavoriteDate time.Time `json:"favorite_date"`
// }

type Favorite struct {
	FavoriteID   uint      `gorm:"primaryKey" json:"favorite_id"`
	UserID       uint      `json:"user_id"`
	EventID      uint      `json:"event_id"`
	FavoriteDate time.Time `json:"favorite_date"`

	User  User  `gorm:"foreignKey:UserID;references:ID"`
	Event Event `gorm:"foreignKey:EventID;references:ID"`
}

func (Favorite) TableName() string {
	return "favorite"
}
