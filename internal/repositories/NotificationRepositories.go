package repositories

import (
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

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

func GetUserTokenNotificationInEvent(eventID string) ([]string, error) {
	var tokens []string
	q := `
		SELECT u.token_notification
		FROM event e
		LEFT JOIN joins j ON e.id = j.event_id
		LEFT JOIN user  u ON j.user_id = u.user_id
		WHERE e.id = ?
		  AND u.token_notification IS NOT NULL
		  AND u.token_notification <> ''
	`
	if err := database.DB.Raw(q, eventID).Scan(&tokens).Error; err != nil {
		return nil, err
	}

	// กันเหนียว dedupe ใน Go
	// uniq := make(map[string]struct{}, len(tokens))
	// out := make([]string, 0, len(tokens))
	// for _, t := range tokens {
	// 	if _, ok := uniq[t]; ok {
	// 		continue
	// 	}
	// 	uniq[t] = struct{}{}
	// 	out = append(out, t)
	// }
	return tokens, nil
}
