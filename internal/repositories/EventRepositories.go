package repositories

import (
	"context"
	"fmt"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"gorm.io/gorm"

	"strconv"
)

func Eventcreate(event *models.Event) error {
	return database.DB.Create(&event).Error
}

func EventChangeNotiDayOne(event *models.Event) error {
	return database.DB.Save(&event).Error
}

func EventDelete(ctx context.Context, eventID int) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) ลบ favorite
		if err := tx.Where("event_id = ?", eventID).
			Delete(&models.Favorite{}).Error; err != nil {
			return fmt.Errorf("delete favorites failed: %w", err)
		}

		// 2) ลบ media (ถ้ามีตารางลูกของ media เช่น media_like, media_comment ให้ลบก่อน media)
		if err := tx.Where("event_id = ?", eventID).
			Delete(&models.Media{}).Error; err != nil {
			return fmt.Errorf("delete media failed: %w", err)
		}

		// 3) ลบ chat messages (ถ้ามี message_reads ฯลฯ ให้ลบก่อน)
		if err := tx.Where("event_id = ?", eventID).
			Delete(&models.ChatMessage{}).Error; err != nil {
			return fmt.Errorf("delete chat messages failed: %w", err)
		}

		// 4) ลบการเข้าร่วม
		if err := tx.Where("event_id = ?", eventID).
			Delete(&models.Join{}).Error; err != nil {
			return fmt.Errorf("delete joins failed: %w", err)
		}

		// 5) ลบการแจ้งเตือน
		if err := tx.Where("event_id = ?", eventID).
			Delete(&models.Notification{}).Error; err != nil {
			return fmt.Errorf("delete favorites failed: %w", err)
		}

		// **สำคัญ**: ตรวจว่ามีตารางอื่นที่อ้าง event_id อีกไหม เช่น Notification, RoomKey, Invite ฯลฯ
		if err := tx.Where("event_id = ?", eventID).
			Delete(&models.Notification{}).Error; err != nil {
			// ถ้าไม่มีคอนสตรัคนี้ในระบบคุณ ลบทิ้งบรรทัดนี้ได้
			return fmt.Errorf("delete notifications failed: %w", err)
		}

		// 5) ลบ event (hard delete)
		res := tx.Unscoped().Where("id = ?", eventID).Delete(&models.Event{})
		if res.Error != nil {
			return fmt.Errorf("delete event failed: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func EventGetAll() (*[]models.Event, error) {
	var events []models.Event

	if err := database.DB.Find(&events).Error; err != nil {
		return nil, err
	}

	return &events, nil
}

func EventFindByID(id int) (*models.Event, error) {
	var Event models.Event

	if err := database.DB.Where("id = ?", id).First(&Event).Error; err != nil {
		return nil, err
	}

	return &Event, nil
}

func EventUpdateImageCover(id int, imagePath string) (*models.Event, error) {
	var event models.Event

	// ค้นหา Event ตาม ID
	if err := database.DB.First(&event, id).Error; err != nil {
		return nil, err
	}

	event.EventImage = imagePath

	if err := database.DB.Save(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func GetEventByUID(uid string) ([]models.Event, error) {
	var events []models.Event
	if err := database.DB.Where("user_id = ?", uid).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func GetEventsWithAttendees(uid string, isCalendar bool, isJoin bool) ([]response.GetallEvent, error) {
	events := make([]response.GetallEvent, 0)

	// --- SELECT/JOIN พื้นฐาน (มี my_join ให้เช็คว่าเราเข้าร่วมหรือยัง) ---
	base := `
		SELECT
			e.id,
			e.event_name,
			e.event_start_time,
			e.event_location_name,
			e.latitude,
			e.longitude,
			COALESCE(e.event_image, '') AS event_image,
			e.type_id,
			COALESCE(attendee_count.count, 0) AS attendee_count,
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM favorite f
					WHERE f.event_id = e.id
					AND f.user_id = ?
				) THEN 1 ELSE 0
			END AS is_favorite,
			IF(my_join.event_id IS NULL, 0, 1) AS is_joined
		FROM event e
		LEFT JOIN (
			SELECT event_id, COUNT(*) AS count
			FROM joins
			GROUP BY event_id
		) attendee_count
			ON e.id = attendee_count.event_id
		LEFT JOIN (
			SELECT DISTINCT event_id
			FROM joins
			WHERE user_id = ?
		) my_join
			ON my_join.event_id = e.id
	`

	// --- เงื่อนไขตามโหมดต่าง ๆ ---
	// โหมด Calendar: ยังไม่หมดเวลา OR ยังไม่เริ่ม (รองรับ end_time เป็น NULL) OR เราเคยเข้าร่วม → แสดงได้
	whereCalendar := `
		WHERE
			( e.event_end_time IS NOT NULL AND e.event_end_time >= NOW() )
			OR ( e.event_end_time IS NULL AND e.event_start_time >= NOW() )
			OR ( my_join.event_id IS NOT NULL )
	`

	// โหมดทั่วไป: ใช้ end_time ถ้ามี ไม่งั้นใช้ start_time
	whereNonCalendar := `
		WHERE COALESCE(e.event_end_time, e.event_start_time) >= NOW()
	`

	// โหมดเฉพาะอีเวนต์ที่เราเข้าร่วมเท่านั้น
	whereOnlyJoined := `
		WHERE my_join.event_id IS NOT NULL
	`

	order := `
		ORDER BY
			CASE WHEN my_join.event_id IS NULL THEN 0 ELSE 1 END DESC,
			e.id ASC;
	`

	// เลือก WHERE ตามพารามิเตอร์
	var query string
	switch {
	case isJoin:
		// ถ้าขอ “เอาเฉพาะที่เราเข้าร่วม” ให้ใช้ whereOnlyJoined ตรง ๆ
		query = base + whereOnlyJoined + order
	case isCalendar:
		query = base + whereCalendar + order
	default:
		query = base + whereNonCalendar + order
	}

	rows, err := database.DB.Raw(query, uid, uid).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ev response.GetallEvent
		if err := rows.Scan(
			&ev.EventId,
			&ev.EventTitle,
			&ev.EventDate,
			&ev.EventLocation,
			&ev.EventLatitude,
			&ev.EventLongitude,
			&ev.EventImage,
			&ev.EventCategory,
			&ev.AttendeeCount,
			&ev.IsFavorite,
			&ev.IsJoin,
		); err != nil {
			return nil, err
		}

		avatars, err := getLastThreeAttendeeAvatars(ev.EventId)
		if err != nil {
			ev.AttendeeAvatars = make([]string, 0)
		} else {
			ev.AttendeeAvatars = avatars
		}

		events = append(events, ev)
	}

	return events, nil
}

func GetEventDetailWithAttendees(eventID int) (*models.Event, error) {
	var event models.Event
	if err := database.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func getLastThreeAttendeeAvatars(eventID int) ([]string, error) {
	var results []struct {
		UserImage string `json:"user_image"`
	}

	query := database.DB.Table("joins").
		Select("user.user_image").
		Joins("JOIN user ON joins.user_id = user.user_id").
		Where("joins.event_id = ?", eventID).
		Order("joins.user_id DESC").
		Limit(3)

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	var imageURLs []string
	for _, result := range results {
		imageURLs = append(imageURLs, result.UserImage)
	}

	for len(imageURLs) < 3 {
		imageURLs = append(imageURLs, "")
	}

	return imageURLs, nil
}

func EventGetMediaByID(eid string) (*[]models.Media, *[]models.Media, error) {
	var mediaImage []models.Media
	var mediaVideo []models.Media

	if err := database.DB.Where("event_id = ? and file_type = 1", eid).Preload("User").Find(&mediaImage).Error; err != nil {
		return nil, nil, err
	}

	if err := database.DB.Where("event_id = ? and file_type = 2", eid).Preload("User").Find(&mediaVideo).Error; err != nil {
		return nil, nil, err
	}

	return &mediaImage, &mediaVideo, nil
}

func EventGetListJoinUser(eid string, currentUserID string) ([]response.EventGetListJoin, error) {
	var queryData []response.EventGetListJoin

	query := `
		SELECT
		u.user_id, u.user_image, u.name, u.email, u.phone, j.status,
		CASE
			WHEN j.status = -1           THEN 99         -- บล็อก ส่งไปท้ายสุดเสมอ
			WHEN u.user_id = e.user_id   THEN 1          -- เจ้าของก่อน
			WHEN u.user_id = ?           THEN 2          -- แล้ว current user
			ELSE 3
		END AS priority
		FROM joins j
		JOIN user u ON j.user_id = u.user_id
		JOIN event e ON j.event_id = e.id
		WHERE j.event_id = ?
		ORDER BY priority, u.name;
	`

	if err := database.DB.Raw(query, currentUserID, eid).Scan(&queryData).Error; err != nil {
		return nil, err
	}

	return queryData, nil
}

func EventGetFavorites(uid string) ([]response.EventGetFavorites, error) {
	var events []response.EventGetFavorites

	err := database.DB.
		Table("event AS e").
		Select(`
			e.id, 
			e.event_name, 
			e.event_image, 
			e.type_id, 
			e.max_media,
			e.event_start_time,
			f.favorite_date
		`).
		Joins("JOIN favorite AS f ON f.event_id = e.id").
		Where("f.user_id = ?", uid).
		Group("e.id").
		Order("MAX(f.favorite_date) ASC").
		Scan(&events).Error

	if err != nil {
		return nil, err
	}
	return events, nil
}

func EventCheckJoinUser(eid, uid string) (bool, error) {
	var join models.Join

	if err := database.DB.Where("event_id = ? AND user_id = ?", eid, uid).First(&join).Error; err != nil {
		return false, nil
	}

	return true, nil
}

func EventCheckFavorite(req request.FavoriteReq) (bool, error) {
	var favorite models.Favorite

	if err := database.DB.Where("event_id = ? AND user_id = ?", req.EID, req.UID).First(&favorite).Error; err != nil {
		return false, err
	}

	return true, nil
}

func AddFavorite(req request.FavoriteReq) (*models.Favorite, error) {
	eventID, err := strconv.ParseUint(req.EID, 10, 64)
	userID, err := strconv.ParseUint(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}

	var favorite models.Favorite
	favorite = models.Favorite{
		UserID:       uint(userID),
		EventID:      uint(eventID),
		FavoriteDate: req.Date,
	}
	err = database.DB.Create(&favorite).Error
	return &favorite, err
}

func RemoveFavorite(req request.FavoriteReq) (*models.Favorite, error) {
	eventID, err := strconv.ParseUint(req.EID, 10, 64)
	userID, err := strconv.ParseUint(req.UID, 10, 64)
	if err != nil {
		return nil, err
	}

	unfavorite := &models.Favorite{
		UserID:  uint(userID),
		EventID: uint(eventID),
	}

	err = database.DB.Where("event_id = ? AND user_id = ?", eventID, userID).Delete(unfavorite).Error
	return unfavorite, err
}

func FindEventBetween(now, next24h time.Time) ([]models.Event, error) {
	var events []models.Event

	err := database.DB.
		Model(&models.Event{}).
		Where("event_start_time BETWEEN ? AND ? AND is_noti_one_day = 0", now, next24h).
		Order("event_start_time ASC").
		Find(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}
