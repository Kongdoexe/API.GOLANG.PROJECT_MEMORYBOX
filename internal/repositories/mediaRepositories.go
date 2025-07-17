package repositories

import (
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

func GetMediaByID(mediaID uint) (*models.Media, *models.User, error) {
	var media models.Media
	if err := database.DB.Where("media_id = ?", mediaID).First(&media).Error; err != nil {
		return nil, nil, err
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", media.UserID).First(&user).Error; err != nil {
		return &media, nil, err
	}

	return &media, &user, nil
}
