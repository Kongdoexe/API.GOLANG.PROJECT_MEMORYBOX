package repositories

import (
	"API.GOLANG.PROJECT_MEMORYBOX/database"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
)

func Joincreate(join *models.Join) error {
	return database.DB.Create(join).Error
}

func JoinFindlastUID() (*models.Join, error) {
	var Join models.Join

	if err := database.DB.Last(&Join).Error; err != nil {
		return nil, err
	}

	return &Join, nil
}

func JoinFindUIDandEID(req *request.JoinRequest) (*models.Join, error) {
	var Join models.Join

	if err := database.DB.Where("user_id = ? and event_id = ?", req.UID, req.EID).First(&Join).Error; err != nil {
		return nil, err
	}

	return &Join, nil
}
