package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/response"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"
	"github.com/jinzhu/copier"
)

func formatThaiDateTime(datetime string) (string, error) {
	// parse ISO8601 string → time.Time
	t, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return "", err
	}

	// ชื่อวันและเดือนภาษาไทย
	weekdays := []string{"อาทิตย์", "จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์"}
	months := []string{
		"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
		"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
	}

	weekdayName := weekdays[int(t.Weekday())]
	day := t.Day()
	monthName := months[int(t.Month())]
	year := t.Year() + 543 // แปลง ค.ศ. → พ.ศ.
	hour := t.Hour()
	minute := t.Minute()

	formatted := fmt.Sprintf("%s, %d %s %d เวลา %02d:%02d", weekdayName, day, monthName, year, hour, minute)
	return formatted, nil
}

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

func EventDelete(eid string) (bool, error) {
	eidInt, err := strconv.Atoi(eid)
	if err != nil {
		return false, errors.New("event ID ไม่ถูกต้อง")
	}

	err = repositories.EventDelete(context.Background(), eidInt)
	if err != nil {
		return false, errors.New("ไม่สามารถลบข้อมูลได้")
	}

	return true, nil
}

func GetEventMainProfile(uid string) ([]response.EventGetMainProfile, error) {
	events, err := repositories.GetEventByUID(uid)
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	var result []response.EventGetMainProfile
	for _, e := range events { // ไม่ต้อง *events แล้ว
		eventDate, _ := formatThaiDateTime(e.EventStartTime.Format(time.RFC3339))
		result = append(result, response.EventGetMainProfile{
			EventId:       int(e.ID),
			EventTitle:    e.EventName,
			EventDate:     eventDate,
			EventLocation: e.EventLocationName,
			EventImage:    e.EventImage,
		})
	}

	if len(result) < 1 {
		result = make([]response.EventGetMainProfile, 0)
	}

	return result, nil

}

func GetEventsWithAttendees(uid string) ([]response.GetallEvent, error) {
	events, err := repositories.GetEventsWithAttendees(uid, false, false)
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	thaiMonths := map[time.Month]string{
		time.January:   "มกราคม",
		time.February:  "กุมภาพันธ์",
		time.March:     "มีนาคม",
		time.April:     "เมษายน",
		time.May:       "พฤษภาคม",
		time.June:      "มิถุนายน",
		time.July:      "กรกฎาคม",
		time.August:    "สิงหาคม",
		time.September: "กันยายน",
		time.October:   "ตุลาคม",
		time.November:  "พฤศจิกายน",
		time.December:  "ธันวาคม",
	}

	for i, event := range events {
		t, err := time.Parse(time.RFC3339, event.EventDate)
		if err != nil {
			return nil, fmt.Errorf("ไม่สามารถแปลงวันที่ event_id %d: %v", event.EventId, err)
		}
		day := t.Day()
		month := thaiMonths[t.Month()]

		formattedDate := fmt.Sprintf("%d %s", day, month)
		events[i].EventDate = formattedDate
	}

	return events, nil
}

func GetEventCalendar(uid string) ([]response.GetallEvent, error) {
	events, err := repositories.GetEventsWithAttendees(uid, true, false)
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	return events, nil
}

func GetEventJoin(uid string) ([]response.GetallEvent, error) {
	events, err := repositories.GetEventsWithAttendees(uid, false, true)
	if err != nil {
		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
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

	start := event.EventStartTime
	end := event.EventEndTime

	day := start.Weekday().String()
	startTime := start.Format("15:04")
	endTime := end.Format("15:04")

	response.EventTimeDisplay = fmt.Sprintf("%s, %s - %s", day, startTime, endTime)

	response.EventDate = start.Format("2 January 2006")

	return &response, nil
}

func EventGetFavorites(uid string) (*[]response.EventGetFavorites, error) {
	responses, err := repositories.EventGetFavorites(uid)
	if err != nil {
		return nil, err
	}

	for i := range responses {
		responses[i].EventDate, _ = formatThaiDateTime(responses[i].EventStartTime.Format(time.RFC3339))
	}

	if len(responses) < 1 {
		responses = make([]response.EventGetFavorites, 0)
	}

	return &responses, nil
}

func EventGetListJoinUser(eid, uid string) (*[]response.EventGetListJoin, error) {
	responses, err := repositories.EventGetListJoinUser(eid, uid)
	if err != nil {
		return nil, err
	}

	if len(responses) < 1 {
		responses = make([]response.EventGetListJoin, 0)
	}

	return &responses, nil
}

func EventGetMediaByID(eid string) (*[]models.Media, *[]models.Media, error) {
	responseImage, responseVideo, err := repositories.EventGetMediaByID(eid)
	if err != nil {
		return nil, nil, err
	}

	// ฟอร์แมตเวลา (UTC + milliseconds)
	layout := "2006-01-02T15:04:05.000Z07:00"

	// จัดการ responseImage
	for i := range *responseImage {
		img := &(*responseImage)[i]
		img.DetailTime, _ = time.Parse(layout, img.DetailTime.UTC().Format(layout))
		img.UploadTime, _ = time.Parse(layout, img.UploadTime.UTC().Format(layout))
	}

	// จัดการ responseVideo
	for i := range *responseVideo {
		vid := &(*responseVideo)[i]
		vid.DetailTime, _ = time.Parse(layout, vid.DetailTime.UTC().Format(layout))
		vid.UploadTime, _ = time.Parse(layout, vid.UploadTime.UTC().Format(layout))
	}

	return responseImage, responseVideo, nil
}

func EventCheckJoinUser(eid, uid string) (bool, error) {
	isJoined, err := repositories.EventCheckJoinUser(eid, uid)
	if err != nil {
		return false, errors.New("ไม่สามารถตรวจสอบการเข้าร่วมได้")
	}

	return isJoined, nil
}

func EventFavorite(req request.FavoriteReq) (string, error) {
	if req.Status {
		_, err := repositories.AddFavorite(req)
		if err != nil {
			return "บันทึกรายการโปรดสำเร็จ", nil
		}
		return "บันทึกรายการโปรดสำเร็จ", err
	} else if !req.Status {
		_, err := repositories.RemoveFavorite(req)
		if err != nil {
			return "ลบสำเร็จ", nil
		}
		return "ลบสำเร็จ", err
	}

	return "", nil
}

// func EventGetFavorite(uid string) ([]response.EventGetFavorite, error) {
// 	events, err := repositories.EventGetFavorite(uid)
// 	if err != nil {
// 		return nil, errors.New("ไม่สามารถดึงข้อมูลได้")
// 	}

// 	for i, event := range events {
// 		tS, err := time.Parse(time.RFC3339, event.EventStart)
// 		tE, err := time.Parse(time.RFC3339, event.EventEnd)
// 		if err != nil {
// 			return nil, fmt.Errorf("ไม่สามารถแปลงวันที่ event_id %d: %v", event.EventId, err)
// 		}

// 		formattedDateStart := tS.Format("2 January 2006")
// 		formattedDateEnd := tE.Format("2 January 2006")
// 		events[i].EventStart = formattedDateStart
// 		events[i].EventEnd = formattedDateEnd
// 	}

// 	return events, nil
// }
