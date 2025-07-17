package repositories

import (
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
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

	if err := database.DB.Where("event_id = ?", id).First(&Event).Error; err != nil {
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

func GetEventsWithAttendees() ([]response.GetallEvent, error) {
	var events []response.GetallEvent

	query := `
		SELECT 
			e.event_id,
			e.event_name,
			e.event_start_date_time,
			e.event_location_name,
			e.event_image,
			COALESCE(attendee_count.count, 0) as attendee_count
		FROM events e
		LEFT JOIN (
			SELECT event_id, COUNT(*) as count
			FROM joins
			GROUP BY event_id
		) attendee_count ON e.event_id = attendee_count.event_id
		ORDER BY e.event_id ASC
	`

	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event response.GetallEvent
		err := rows.Scan(
			&event.EventId,
			&event.EventTitle,
			&event.EventDate,
			&event.EventLocation,
			&event.EventImage,
			&event.AttendeeCount,
		)
		if err != nil {
			return nil, err
		}

		avatars, err := getLastThreeAttendeeAvatars(event.EventId)
		if err != nil {
			event.AttendeeAvatars = []string{}
		} else {
			event.AttendeeAvatars = avatars
		}

		events = append(events, event)
	}

	return events, nil
}

func GetEventDetailWithAttendees(eventID int) (*models.Event, error) {
	var event models.Event
	if err := database.DB.Where("event_id = ?", eventID).First(&event).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func getLastThreeAttendeeAvatars(eventID int) ([]string, error) {
	var results []struct {
		UserImage string `json:"user_image"`
	}

	query := database.DB.Table("joins").
		Select("users.user_image").
		Joins("JOIN users ON joins.user_id = users.user_id").
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
	var mediaImage *[]models.Media
	var mediaVideo *[]models.Media

	if err := database.DB.Where("event_id = ? and file_type = 1", eid).Find(&mediaImage).Error; err != nil {
		return nil, nil, err
	}

	if err := database.DB.Where("event_id = ? and file_type = 2", eid).Find(&mediaVideo).Error; err != nil {
		return nil, nil, err
	}

	return mediaImage, mediaVideo, nil
}
