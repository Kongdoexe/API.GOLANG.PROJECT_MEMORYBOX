package models

type User struct {
	ID             uint       `json:"id" gorm:"primaryKey;column:user_id"`
	UserImage      string     `json:"user_image" gorm:"column:user_image"`
	Name           string     `json:"name" gorm:"size:500"`
	Email          string     `json:"email" gorm:"unique;size:500"`
	Phone          string     `json:"phone" gorm:"unique;not null"`
	Password       string     `json:"password"`
	IsNotification bool       `json:"is_notification" gorm:"column:is_notification"`
	GoogleID       *string    `json:"google_id" gorm:"column:google_id"`
	Events         []Event    `json:"events" gorm:"foreignKey:UserID;onDelete:CASCADE"`
	Joins          []Join     `json:"joins" gorm:"foreignKey:UserID;onDelete:CASCADE"`
	Favorites      []Favorite `json:"favorites" gorm:"foreignKey:UserID;onDelete:CASCADE"`
	Media          []Media    `json:"media" gorm:"foreignKey:UserID;onDelete:CASCADE"`
}
