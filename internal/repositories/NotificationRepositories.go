package repositories

import (
	"fmt"

	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

func CreateNotification(req *models.Notification) error {
	return database.DB.Create(req).Error
}

func CreateNotificationsBulk(notis []models.Notification, batchSize int) error {
	if len(notis) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 200
	}
	return database.DB.CreateInBatches(notis, batchSize).Error
}

func CreateNotificationsLoopTx(notis []models.Notification) error {
	if len(notis) == 0 {
		return nil
	}
	tx := database.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	for i := range notis {
		if err := tx.Create(&notis[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func PutNotification(req *models.Notification) error {
	return database.DB.Save(req).Error
}

func GetUserIDsInEvent(eventID int) ([]models.User, error) {
	var uids []models.User
	err := database.DB.
		Table("joins").
		Select("user_id").
		Where("event_id = ?", eventID).
		Scan(&uids).Error
	if err != nil {
		return nil, err
	}
	return uids, nil
}

func GetUserNamesByIDs(userIDs []string) ([]string, error) {
	var names []string
	err := database.DB.
		Table("user").
		Select("name").
		Where("user_id IN ?", userIDs).
		Scan(&names).Error
	return names, err
}

func InsertTokenNotification(req request.RequestTokenNotification) (bool, error) {
	var user models.User

	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		return false, err
	}

	user.TokenNotification = req.TokenNotification

	if err := database.DB.Save(&user).Error; err != nil {
		return false, err
	}

	return true, nil
}

func GetUserNotification(uid string) ([]models.Notification, error) {
	var notification []models.Notification

	err := database.DB.
		Where("user_id = ?", uid).
		Order("notification_time DESC").
		Find(&notification).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %v", err)
	}

	return notification, nil
}

func GetUserTokenNotificationInEvent(req request.RequestSendNotiChat) ([]string, error) {
	var tokens []string
	err := database.DB.
		Table("event AS e").
		Select("u.token_notification").
		Joins("LEFT JOIN joins j ON e.id = j.event_id").
		Joins("LEFT JOIN user u ON j.user_id = u.user_id").
		Where("e.id = ? AND u.token_notification IS NOT NULL AND u.token_notification <> '' AND u.user_id <> ?", req.EventID, req.UserID).
		Scan(&tokens).Error

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func GetUserTokenNotificationInEvent24Hr(eventId string) ([]string, error) {
	var tokens []string
	err := database.DB.
		Table("event AS e").
		Select("u.token_notification").
		Joins("LEFT JOIN joins j ON e.id = j.event_id").
		Joins("LEFT JOIN user u ON j.user_id = u.user_id").
		Where("e.id = ? AND u.token_notification IS NOT NULL AND u.token_notification <> ''", eventId).
		Scan(&tokens).Error

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func CheckNotiUser(userIDs []string, eventID string) (map[string]bool, error) {
	if len(userIDs) == 0 {
		return map[string]bool{}, nil
	}

	result := make(map[string]bool)

	for _, uid := range userIDs {
		result[uid] = false
	}

	type NotiUser struct {
		UserID            string `json:"user_id"`
		TokenNotification string `json:"token_notification"`
	}

	var notiUsers []NotiUser

	err := database.DB.
		Table("notification").
		Select("user_id, token_notification").
		Where("user_id IN ? AND event_id = ?", userIDs, eventID).
		Where("token_notification IS NOT NULL AND token_notification != ''").
		Find(&notiUsers).Error
	if err != nil {
		return nil, err
	}

	for _, n := range notiUsers {
		result[n.UserID] = true
	}

	return result, nil
}

func GetTokensByUserIDs(userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	var tokens []string

	err := database.DB.
		Table("user").
		Select("token_notification").
		Where("user_id IN ?", userIDs).
		Where("token_notification IS NOT NULL AND token_notification != ''").
		Scan(&tokens).Error

	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(tokens))
	uniqueTokens := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		uniqueTokens = append(uniqueTokens, t)
	}

	return uniqueTokens, nil
}

func GetTokensUserByEventId(eventId string) ([]string, error) {
	var tokens []string

	err := database.DB.
		Table("user").
		Select("token_notification").
		Where("user_id IN ?", eventId).
		Where("token_notification IS NOT NULL AND token_notification != ''").
		Scan(&tokens).Error

	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(tokens))
	uniqueTokens := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		uniqueTokens = append(uniqueTokens, t)
	}

	return uniqueTokens, nil
}

func UpdateIsReadNotification(nid string) error {
	return database.DB.
		Table("notification").
		Where("notification_id = ?", nid).
		Update("is_read", 1).Error
}
