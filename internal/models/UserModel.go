package models

// type User struct {
// 	ID             uint       `json:"id" gorm:"primaryKey;column:user_id"`
// 	UserImage      string     `json:"user_image" gorm:"column:user_image"`
// 	Name           string     `json:"name" gorm:"size:500"`
// 	Email          string     `json:"email" gorm:"unique;size:500;not null"`
// 	Phone          string     `json:"phone"`
// 	Password       string     `json:"password"`
// 	IsNotification	int		`json:"is_notification"`
// 	TokenNotification string  `json:"token_notification"`
// 	GoogleID       *string    `json:"google_id" gorm:"column:google_id"`
// 	Events         []Event    `json:"events" gorm:"foreignKey:UserID;onDelete:CASCADE"`
// 	Joins          []Join     `json:"joins" gorm:"foreignKey:UserID;onDelete:CASCADE"`
// 	Favorites      []Favorite `json:"favorites" gorm:"foreignKey:UserID;onDelete:CASCADE"`
// 	Media          []Media    `json:"media" gorm:"foreignKey:UserID;onDelete:CASCADE"`
// }

type User struct {
	ID                uint    `gorm:"primaryKey;column:user_id" json:"id"`
	UserImage         string  `json:"user_image" gorm:"column:user_image"`
	Name              string  `json:"name" gorm:"size:255;column:name"`
	Email             string  `json:"email" gorm:"unique;size:255;not null;column:email"`
	Phone             string  `json:"phone" gorm:"column:phone"`
	Password          string  `json:"password" gorm:"column:password"`
	IsNotification    int     `json:"is_notification" gorm:"column:is_notification"`
	TokenNotification string  `json:"token_notification" gorm:"column:token_notification"`
	GoogleID          *string `json:"google_id" gorm:"column:google_id"`

	// relations
	ResetPassToken ResetPassToken `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // 1-1
	Notification   []Notification `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // 1-M
	Media          []Media        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Event          []Event        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (User) TableName() string { return "user" }
