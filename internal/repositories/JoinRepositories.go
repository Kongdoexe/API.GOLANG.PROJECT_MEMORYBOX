package repositories

import (
	"fmt"

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

func JoinBlocked(req request.BlockedJoin) (*models.Join, error) {
	var join models.Join

	if err := database.DB.
		Where("user_id = ? AND event_id = ?", req.Uid, req.Eid).
		First(&join).Error; err != nil {
		return nil, err
	}

	if join.Status == -1 {
		join.Status = 1
	} else {
		join.Status = -1
	}

	if err := database.DB.
		Model(&join).
		Update("status", join.Status).Error; err != nil {
		return nil, err
	}

	return &join, nil
}

func CheckBlocked(req request.JoinRequest) (bool, error) {
	var status int
	res := database.DB.
		Raw(`SELECT status 
			 FROM joins 
			 WHERE user_id = ? AND event_id = ? 
			 LIMIT 1`, req.UID, req.EID).
		Scan(&status)

	if res.Error != nil {
		return false, res.Error
	}
	fmt.Print(status)

	return status == -1, nil
}

func GetCountUserJoin(eventId string) (int, error) {
	var length int
	res := database.DB.Raw(`SELECT Count(*) FROM joins WHERE event_id = ?`, eventId).Scan(&length)

	if res.Error != nil {
		return 0, res.Error
	}

	return length, nil
}
