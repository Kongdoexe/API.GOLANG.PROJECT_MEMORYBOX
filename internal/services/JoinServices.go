package services

import (
	"errors"
	"strconv"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
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

	if err = SendNotificationEventUserJoin(req.EID, req.UID); err != nil {
		return nil, errors.New("ไม่สามารถส่งการแจ้งเตือนไปยังผู้ใช้คนอื่นได้")
	}

	if err = repositories.Joincreate(&join); err != nil {
		return nil, errors.New("ไม่สามารถสร้างผู้ใช้ได้")
	}

	return &join, nil
}

func JoinBlocked(req request.BlockedJoin) (*[]response.EventGetListJoin, error) {
	_, err := repositories.JoinBlocked(req)
	if err != nil {
		return nil, err
	}

	res, err := repositories.EventGetListJoinUser(req.Eid, req.Current)
	if err != nil {
		return nil, err
	}

	if len(res) < 1 {
		res = make([]response.EventGetListJoin, 0)
	}

	return &res, nil

}

func CheckBlocked(req request.JoinRequest) (bool, error) {
	res, _ := repositories.CheckBlocked(req)

	return res, nil
}
