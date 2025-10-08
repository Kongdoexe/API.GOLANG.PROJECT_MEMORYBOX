package models

// type Join struct {
// 	ID              uint `json:"id" gorm:"primaryKey;column:join_id"`
// 	UserID          uint `json:"user_id" gorm:"event_id"`
// 	EventID         uint `json:"event_id" gorm:"event_id"`
// 	Enable          bool `json:"enable" gorm:"enable"`                           // เช็คสถานนะว่าโดนแบน หรือไม่โดน
// 	AccessModifiers int  `json:"access_modifiers" gorm:"colum:access_modifiers"` // ยอมรับการเข้าร่วม 1 ยังไม่ได้รับการเข้าร่วม 0
// 	JoinLevel       int  `json:"join_level"`                                     //  0 = ไม่มีสิทธิ์, 1 = อ่านได้, 2 = ดูวิดีโอย้อนหลังได้
// }

type Join struct {
	JoinID  uint `gorm:"primaryKey" json:"join_id"`
	UserID  uint `json:"user_id"`
	EventID uint `json:"event_id"`
	Status  int  `json:"status"`

	User  User  `gorm:"foreignKey:UserID;references:ID"`
	Event Event `gorm:"foreignKey:EventID;references:ID"`
}

func (Join) TableName() string {
	return "joins"
}
