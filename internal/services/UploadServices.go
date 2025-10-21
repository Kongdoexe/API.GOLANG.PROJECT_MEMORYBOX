package services

import (
	"fmt"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
)

// ตรวจว่าอีเวนต์ยังไม่จบ
func CheckDateEvent(eventId uint) error {
	var event struct {
		EndTime time.Time `gorm:"column:event_end_time"`
	}

	// ดึงเวลาสิ้นสุดของอีเวนต์จาก DB
	if err := database.DB.
		Table("event").
		Select("event_end_time").
		Where("id = ?", eventId).
		Limit(1).
		Scan(&event).Error; err != nil {
		return fmt.Errorf("ไม่สามารถดึงข้อมูลอีเวนต์ได้: %v", err)
	}

	now := time.Now() // ใช้ local time ของเซิร์ฟเวอร์
	if now.After(event.EndTime) {
		return fmt.Errorf("อีเวนต์นี้จบไปแล้ว (จบเมื่อ %s)", event.EndTime.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func CheckMaxMedia(eventId uint, userId uint) error {
	eventData, err := repositories.EventFindByID(int(eventId))
	if err != nil {
		return fmt.Errorf("ไม่พบข้อมูลอีเวนต์: %v", err)
	}

	var count int64
	q := database.DB.
		Model(&models.Media{}).
		Where("event_id = ? AND user_id = ?", eventId, userId)

	if err := q.Count(&count).Error; err != nil {
		return fmt.Errorf("ไม่สามารถนับจำนวนมีเดียได้: %v", err)
	}

	if count >= int64(eventData.MaxMedia) {
		return fmt.Errorf("จำนวนมีเดียถึงขีดจำกัดสูงสุดแล้ว (%d/%d)", count, eventData.MaxMedia)
	}
	return nil
}
