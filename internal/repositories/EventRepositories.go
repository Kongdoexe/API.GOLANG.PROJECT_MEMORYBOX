package repositories

import (
	"fmt"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"

	"strconv"
)

func Eventcreate(event *models.Event) error {
	return database.DB.Create(&event).Error
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

func GetEventsWithAttendees(uid string) ([]response.GetallEvent, error) {
	events := make([]response.GetallEvent, 0)

	query := `
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
		-- กรองเฉพาะอีเวนต์ที่ยังไม่จบ (หรือไม่มี end_time ให้ถือว่าใช้ start_time แทน)
		WHERE COALESCE(e.event_end_time, e.event_start_time) >= NOW()
		ORDER BY
			CASE WHEN my_join.event_id IS NULL THEN 0 ELSE 1 END,
			e.id ASC;
	`

	rows, err := database.DB.Raw(query, uid, uid).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event response.GetallEvent
		if err := rows.Scan(
			&event.EventId,
			&event.EventTitle,
			&event.EventDate,
			&event.EventLocation,
			&event.EventLatitude,
			&event.EventLongitude,
			&event.EventImage,
			&event.EventCategory,
			&event.AttendeeCount,
			&event.IsFavorite,
			&event.IsJoin,
		); err != nil {
			return nil, err
		}

		avatars, err := getLastThreeAttendeeAvatars(event.EventId)
		if err != nil {
			event.AttendeeAvatars = make([]string, 0) // ไม่ให้เป็น nil
		} else {
			event.AttendeeAvatars = avatars
		}

		events = append(events, event)
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

	if err := database.DB.Where("event_id = ? and file_type = 1", eid).Find(&mediaImage).Error; err != nil {
		return nil, nil, err
	}

	if err := database.DB.Where("event_id = ? and file_type = 2", eid).Find(&mediaVideo).Error; err != nil {
		return nil, nil, err
	}

	return &mediaImage, &mediaVideo, nil
}

func EventGetListJoinUser(eid string, currentUserID string) ([]models.User, error) {
	var users []models.User

	err := database.DB.
		Table("joins").
		Select("user.*").
		Joins("JOIN user ON joins.user_id = user.user_id").
		Joins("JOIN event ON joins.event_id = event.id").
		Where("joins.event_id = ?", eid).
		Order(fmt.Sprintf(`
			CASE 
				WHEN user.user_id = event.user_id THEN 1
				WHEN user.user_id = %s THEN 2
				ELSE 3
			END
		`, currentUserID)).
		Scan(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
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
