package services

import (
	"errors"
	"strconv"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
)

func ValidateJoinAndEvent(reqJoin *request.JoinRequest) (*models.Event, error) {
	eid, err := strconv.ParseUint(reqJoin.EID, 10, 32)
	if err != nil {
		return nil, errors.New("ไม่สามารถแปลง EventID ได้")
	}
	if _, err := repositories.UserFindByID(reqJoin.UID); err != nil {
		return nil, errors.New("ไม่พบผู้ใช้")
	}
	ev, err := repositories.EventFindByID(int(eid))
	if err != nil {
		return nil, errors.New("ไม่พบอีเวนต์")
	}
	if _, err := repositories.JoinFindUIDandEID(reqJoin); err != nil {
		return nil, errors.New("คุณยังไม่เข้าร่วมอีเวนต์")
	}
	blocked, err := repositories.CheckBlocked(*reqJoin)
	if err != nil {
		return nil, errors.New("คุณถูกแบนจากอีเวนต์")
	}
	if blocked {
		return nil, errors.New("คุณถูกแบนจากอีเวนต์")
	}
	return ev, nil
}

func InsertMessage(req request.InsertMessage) error {
	_, err := InsertMessageAndReturn(req)
	return err
}

func InsertMessageAndReturn(req request.InsertMessage) (*models.ChatMessage, error) {
	uid, err := strconv.ParseUint(req.Uid, 10, 32)
	if err != nil {
		return nil, errors.New("ไม่สามารถแปลง UserID ได้")
	}
	eid, err := strconv.ParseUint(req.Eid, 10, 32)
	if err != nil {
		return nil, errors.New("ไม่สามารถแปลง EventID ได้")
	}

	if _, err := ValidateJoinAndEvent(&request.JoinRequest{EID: req.Eid, UID: req.Uid}); err != nil {
		return nil, err
	}

	chat := &models.ChatMessage{
		UserID:   uint(uid),
		EventID:  uint(eid),
		Message:  req.Msg,
		ChatDate: time.Now(),
	}
	if err := repositories.InsertMessage(chat); err != nil {
		return nil, errors.New("ไม่สามารถส่งแชทได้")
	}

	created, err := repositories.FindMessageWithUser(chat.MessageID)
	if err != nil {
		// ถ้าพลาด preload ก็คืนแบบไม่ preload
		return chat, nil
	}
	return created, nil
}

func GetMessage(req request.GetMessage) (*[]models.ChatMessage, error) {
	if _, err := ValidateJoinAndEvent(&request.JoinRequest{EID: req.Eid, UID: req.Uid}); err != nil {
		return nil, err
	}
	return repositories.GetMessage(req)
}
