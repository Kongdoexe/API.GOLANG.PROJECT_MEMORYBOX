package services

import (
	"errors"
	"strconv"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
)

func JoinCreate(req *request.JoinRequest) (*models.Join, error) {
	var join models.Join
	user, err := repositories.UserFindByID(req.UID)
	if err != nil {
		return nil, errors.New("ไม่พบผู้ใช้")
	}
	eid, err := strconv.Atoi(req.EID)
	if err != nil {
		return nil, errors.New("รหัสอีเวนต์ไม่ถูกต้อง")
	}
	event, err := repositories.EventFindByID(eid)
	if err != nil {
		return nil, errors.New("ไม่พบอีเวนต์")
	}

	check, _ := repositories.JoinFindUIDandEID(req)
	if check != nil {
		return nil, errors.New("คุณเข้าร่วมอีเวนต์อยู่แล้ว")
	}

	join.UserID = user.ID
	join.EventID = event.ID
	join.Status = 1

	if err = repositories.Joincreate(&join); err != nil {
		return nil, errors.New("ไม่สามารถสร้างผู้ใช้ได้")
	}

	return &join, nil
}
