package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"github.com/jinzhu/copier"
)

func EventGetAll() (*[]models.Event, error) {
	events, err := repositories.EventGetAll()
	if err != nil {
		return nil, errors.New("ไม่สามารถค้นหาได้")
	}

	return events, nil
}

func EventCreate(req *models.Event) (*models.Event, error) {
	if err := repositories.Eventcreate(req); err != nil {
		return nil, errors.New("ไม่สามารถสร้างอีเวนต์ได้")
	}

	return req, nil
}

func EventUploadImageCover(file *multipart.FileHeader, eid string) (string, error) {
	if file.Size > int64(dtos.MaxSize) {
		return "", errors.New("ไฟล์มีขนาดเกินที่กำหนด (10MB)")
	}

	eventFolder := filepath.Join(dtos.UploadDir, fmt.Sprintf("event_%s", eid))
	if _, err := os.Stat(eventFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(eventFolder, os.ModePerm); err != nil {
			return "", err
		}
	}

	filename := "coverImage" + filepath.Ext(file.Filename)
	savePath := filepath.Join(eventFolder, filename)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// แปลง eid จาก string → int
	eidInt, err := strconv.Atoi(eid)
	if err != nil {
		return "", errors.New("event ID ไม่ถูกต้อง")
	}

	imageURL := fmt.Sprintf("%s/uploads/event_%s/%s", dtos.BaseURL, eid, filename)

	// ส่งไปอัปเดตในฐานข้อมูล
	_, err = repositories.EventUpdateImageCover(eidInt, imageURL)
	if err != nil {
		return "", err
	}

	return imageURL, nil
}

func GetEventsWithAttendees() ([]response.GetallEvent, error) {
	events, err := repositories.GetEventsWithAttendees()
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	for i, event := range events {
		t, err := time.Parse(time.RFC3339, event.EventDate)
		if err != nil {
			return nil, fmt.Errorf("ไม่สามารถแปลงวันที่ event_id %d: %v", event.EventId, err)
		}

		events[i].EventDate = t.Format("2 January")
	}

	return events, nil
}

func GetEventDetailWithAttendees(eid int) (*response.DetailEvent, error) {
	var response response.DetailEvent

	event, err := repositories.GetEventDetailWithAttendees(eid)
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	if event == nil {
		return nil, errors.New("ไม่พบอีเวนต์")
	}

	userData, err := repositories.UserFindByID(fmt.Sprintf("%d", event.UserID))
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	if err := copier.Copy(&response, &event); err != nil {
		return nil, err
	}

	if err := copier.Copy(&response, &userData); err != nil {
		return nil, err
	}

	start := event.EventStartDateTime
	end := event.EventEndDateTime

	day := start.Weekday().String()
	startTime := start.Format("15:04")
	endTime := end.Format("15:04")

	response.EventTimeDisplay = fmt.Sprintf("%s, %s - %s", day, startTime, endTime)

	response.EventDate = start.Format("2 January 2006")

	return &response, nil
}

func EventGetMediaByID(eid string) (*[]models.Media, *[]models.Media, error) {
	responseImage, responseVideo, err := repositories.EventGetMediaByID(eid)
	if err != nil {
		return nil, nil, err
	}

	return responseImage, responseVideo, nil
}
