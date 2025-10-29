package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"API.GOLANG.PROJECT_MEMORYBOX/firebase" // ตรวจพาธให้ตรงโปรเจกต์
	"API.GOLANG.PROJECT_MEMORYBOX/internal/dtos/request"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/models"
	"API.GOLANG.PROJECT_MEMORYBOX/internal/repositories"

	"firebase.google.com/go/v4/messaging"
)

// ---------- Notification Types (เรียง 1,2,3) ----------
const (
	NotiTypeSystem  = 1 // จากระบบ
	NotiTypeMessage = 2 // ข้อความใหม่ในแชทอีเวนต์
	NotiTypeInvite  = 3 // การเชิญเข้าร่วมอีเวนต์
	NotiTypeJoin    = 4 // การเชิญเข้าร่วมอีเวนต์
	NotiTypeEvent24 = 5 // อีเวนต์จะเริ่มภายใน 24 ชม.
)

// ---------- Helpers ----------
func dedupeNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func nonEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func buildCommonData(t, sub, eventID string, extra map[string]string) map[string]string {
	d := map[string]string{
		"type":         t,
		"subtype":      sub,
		"event_id":     eventID,
		"click_action": "FLUTTER_NOTIFICATION_CLICK",
	}
	for k, v := range extra {
		d[k] = v
	}
	return d
}

// ---------- Query APIs ----------
func GetUserNotification(uid string) ([]models.Notification, []models.Event, error) {
	notis, err := repositories.GetUserNotification(uid)
	if err != nil {
		return nil, nil, errors.New("ไม่สามารถดึงข้อมูลได้")
	}

	seen := make(map[uint]struct{})
	events := make([]models.Event, 0, 8)

	for i := range notis {
		if notis[i].EventID == nil {
			continue
		}
		eid := uint(*notis[i].EventID)

		if _, ok := seen[eid]; ok {
			continue
		}

		ev, err := repositories.EventFindByID(int(eid))
		if err != nil {
			log.Printf("⚠️ ไม่พบอีเวนต์ id=%d: %v", eid, err)
			continue
		}

		events = append(events, *ev)
		seen[eid] = struct{}{}
	}

	return notis, events, nil
}

func InsertTokenNotification(req request.RequestTokenNotification) (bool, error) {
	_, err := repositories.UserFindByID(req.UserID)
	if err != nil {
		return false, errors.New("User not found")
	}
	status, err := repositories.InsertTokenNotification(req)
	if err != nil {
		return false, errors.New("You can update the information, try again.")
	}
	return status, nil
}

func SendNotificationEvent(req request.RequestSendNotiEvent) error {
	eventID, err := strconv.Atoi(req.EventID)
	if err != nil {
		return fmt.Errorf("invalid event ID: %v", err)
	}

	eventData, err := repositories.EventFindByID(eventID)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลอีเวนต์ไม่สำเร็จ: %v", err)
	}

	statusMap, err := repositories.CheckNotiUser(req.UserIDs, req.EventID)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลผู้ใช้ไม่สำเร็จ: %v", err)
	}

	var notified, missing []string
	for _, uid := range req.UserIDs {
		if statusMap[uid] {
			notified = append(notified, uid)
		} else {
			missing = append(missing, uid)
		}
	}

	var inviterName string
	if req.InviterID != "" {
		if _, err := strconv.Atoi(req.InviterID); err == nil {
			if u, err := repositories.UserFindByID(req.InviterID); err == nil {
				inviterName = u.Name
			}
		}
	}

	title := "การเชิญเข้าร่วมอีเวนต์"
	body := fmt.Sprintf("%s เชิญคุณเข้าร่วมอีเวนต์ %s", nonEmpty(inviterName, "ผู้ใช้"), eventData.EventName)

	// ✅ data-only + type=3
	data := buildCommonData("event", "invited", req.EventID, map[string]string{
		"notification_type": strconv.Itoa(NotiTypeInvite), // 3
		"title":             title,
		"body":              body,
		"inviter_id":        req.InviterID,
	})

	tokensRaw, err := repositories.GetTokensByUserIDs(missing)
	if err != nil {
		return fmt.Errorf("ดึง token ผู้ใช้ไม่สำเร็จ: %v", err)
	}
	tokens := dedupeNonEmpty(tokensRaw)
	if len(tokens) == 0 {
		userNames, _ := repositories.GetUserNamesByIDs(req.UserIDs)
		if len(userNames) > 0 {
			return fmt.Errorf("ไม่พบ token สำหรับผู้ใช้: %v", strings.Join(userNames, ", "))
		}
		return fmt.Errorf("ไม่พบ token สำหรับ user_ids: %v", req.UserIDs)
	}

	now := time.Now()
	notis := make([]models.Notification, 0, len(req.UserIDs))
	for _, uidStr := range req.UserIDs {
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			continue
		}
		notis = append(notis, models.Notification{
			UserID:           uid,
			EventID:          &eventID,
			Title:            title,
			Detail:           body,
			NotificationType: NotiTypeInvite,
			NotificationTime: now,
			IsRead:           false,
		})
	}
	if err := repositories.CreateNotificationsLoopTx(notis); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	const chunk = 500
	ctx := context.Background()
	for i := 0; i < len(tokens); i += chunk {
		end := i + chunk
		if end > len(tokens) {
			end = len(tokens)
		}
		msg := &messaging.MulticastMessage{
			Tokens: tokens[i:end],
			Data:   data, // ✅ data-only
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-push-type": "background",
					"apns-priority":  "5",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{ContentAvailable: true},
				},
			},
		}
		if _, err := firebase.MessagingClient.SendEachForMulticast(ctx, msg); err != nil {
			log.Printf("FCM multicast error (%d-%d): %v", i, end, err)
			continue
		}
	}
	return nil
}

func SendNotificationChat(req request.RequestSendNotiChat) error {
	data := buildCommonData("chat", "message", req.EventID, map[string]string{
		"notification_type": strconv.Itoa(NotiTypeMessage), // 2
	})

	eventIdInt, err := strconv.Atoi(req.EventID)
	if err != nil {
		return fmt.Errorf("invalid event ID format: %v", err)
	}
	userIdInt, err := strconv.Atoi(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id format: %v", err)
	}

	eventData, err := repositories.EventFindByID(eventIdInt)
	if err != nil {
		return fmt.Errorf("ไม่สามารถดึงข้อมูลอีเวนต์ได้: %v", err)
	}

	userData, err := repositories.GetUserIDsInEvent(eventIdInt)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้เข้าร่วมอีเวนต์ล้มเหลว: %v", err)
	}

	seenUID := make(map[int]struct{}, len(userData))
	filteredUIDs := make([]int, 0, len(userData))
	for _, uid := range userData {
		if int(uid.ID) == userIdInt {
			continue
		}
		if _, ok := seenUID[int(uid.ID)]; ok {
			continue
		}
		seenUID[int(uid.ID)] = struct{}{}
		filteredUIDs = append(filteredUIDs, int(uid.ID))
	}
	if len(filteredUIDs) == 0 {
		return nil
	}

	title := "การแจ้งเตือนแชท"
	body := "มีข้อความใหม่ในอีเวนต์ " + eventData.EventName

	now := time.Now()
	notis := make([]models.Notification, 0, len(filteredUIDs))
	for _, uid := range filteredUIDs {
		notis = append(notis, models.Notification{
			UserID:           uid,
			EventID:          &eventIdInt,
			Title:            title,
			Detail:           body,
			NotificationType: NotiTypeMessage,
			NotificationTime: now,
			IsRead:           false,
		})
	}
	if err := repositories.CreateNotificationsBulk(notis, 200); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	userIDsStr := make([]string, 0, len(filteredUIDs))
	for _, uid := range filteredUIDs {
		userIDsStr = append(userIDsStr, strconv.Itoa(uid))
	}
	tokens, err := repositories.GetTokensByUserIDs(userIDsStr)
	if err != nil {
		return fmt.Errorf("ไม่สามารถดึง token ผู้ใช้ได้: %v", err)
	}
	tokens = dedupeNonEmpty(tokens)
	if len(tokens) == 0 {
		return nil
	}

	// ✅ ฝัง title/body ลง data
	data["title"] = title
	data["body"] = body

	const chunk = 500
	ctx := context.Background()
	for i := 0; i < len(tokens); i += chunk {
		end := i + chunk
		if end > len(tokens) {
			end = len(tokens)
		}
		msg := &messaging.MulticastMessage{
			Tokens: tokens[i:end],
			Data:   data, // ✅ data-only
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-push-type": "background",
					"apns-priority":  "5",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{ContentAvailable: true},
				},
			},
		}
		if _, err := firebase.MessagingClient.SendEachForMulticast(ctx, msg); err != nil {
			log.Printf("FCM multicast error (%d-%d): %v", i, end, err)
			continue
		}
	}
	return nil
}

func SendNotificationEventjoin(req request.RequestSendNotiChat) error {
	// ✅ subtype ถูกต้อง
	data := buildCommonData("event", "joined", req.EventID, map[string]string{
		"notification_type": strconv.Itoa(NotiTypeJoin), // 4
	})

	eventIdInt, err := strconv.Atoi(req.EventID)
	if err != nil {
		return fmt.Errorf("invalid event ID format: %v", err)
	}
	userIdInt, err := strconv.Atoi(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id format: %v", err)
	}

	eventData, err := repositories.EventFindByID(eventIdInt)
	if err != nil {
		return fmt.Errorf("ไม่สามารถดึงข้อมูลอีเวนต์ได้: %v", err)
	}

	dataUserJoin, err := repositories.UserFindByID(req.UserID)
	if err != nil {
		return fmt.Errorf("ไม่สามารถดึงข้อมูลผู้ใช้ได้: %v", err)
	}

	userData, err := repositories.GetUserIDsInEvent(eventIdInt)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้เข้าร่วมอีเวนต์ล้มเหลว: %v", err)
	}
	seenUID := make(map[int]struct{}, len(userData))
	filteredUIDs := make([]int, 0, len(userData))
	for _, uid := range userData {
		// ตัดตัวผู้ที่เพิ่ง join เอง
		if int(uid.ID) == userIdInt {
			continue
		}
		if _, ok := seenUID[int(uid.ID)]; ok {
			continue
		}
		seenUID[int(uid.ID)] = struct{}{}
		filteredUIDs = append(filteredUIDs, int(uid.ID))
	}
	if len(filteredUIDs) == 0 {
		return nil
	}

	title := "การแจ้งเตือนการเข้าร่วม"
	body := fmt.Sprintf("ผู้ใช้ %s เข้าร่วมอีเวนต์ %s แล้ว!", dataUserJoin.Name, eventData.EventName)

	now := time.Now()
	notis := make([]models.Notification, 0, len(filteredUIDs))
	for _, uid := range filteredUIDs {
		notis = append(notis, models.Notification{
			UserID:           uid,
			EventID:          &eventIdInt,
			Title:            title,
			Detail:           body,
			NotificationType: NotiTypeJoin,
			NotificationTime: now,
			IsRead:           false,
		})
	}
	if err := repositories.CreateNotificationsBulk(notis, 200); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	userIDsStr := make([]string, 0, len(filteredUIDs))
	for _, uid := range filteredUIDs {
		userIDsStr = append(userIDsStr, strconv.Itoa(uid))
	}

	tokens, err := repositories.GetTokensByUserIDs(userIDsStr)
	if err != nil {
		return fmt.Errorf("ไม่สามารถดึง token ผู้ใช้ได้: %v", err)
	}
	tokens = dedupeNonEmpty(tokens)
	if len(tokens) == 0 {
		return nil
	}

	// ✅ ฝัง title/body ลง data (ให้แอปโชว์ local noti เอง)
	data["title"] = title
	data["body"] = body

	const chunk = 500
	ctx := context.Background()
	success, failure := 0, 0
	for i := 0; i < len(tokens); i += chunk {
		end := i + chunk
		if end > len(tokens) {
			end = len(tokens)
		}

		msg := &messaging.MulticastMessage{
			Tokens: tokens[i:end],
			Data:   data, // ✅ data-only
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-push-type": "background",
					"apns-priority":  "5",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{ContentAvailable: true},
				},
			},
		}
		resp, err := firebase.MessagingClient.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("FCM multicast error (%d-%d): %v", i, end, err)
			continue
		}
		success += resp.SuccessCount
		failure += resp.FailureCount
	}
	return nil
}

func SendNotificationEvent24Hr(eventIDStr string) error {
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fmt.Errorf("invalid event ID: %v", err)
	}

	eventData, err := repositories.EventFindByID(eventID)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลอีเวนต์ไม่สำเร็จ: %v", err)
	}

	title := "การแจ้งเตือนอีเวนต์จะเริ่มภายใน 24 ชั่วโมง"
	body := fmt.Sprintf("อีเวนต์ %s กำลังจะเริ่มภายใน 24 ชั่วโมง", eventData.EventName)

	// ✅ data-only + notification_type=5 + ฝัง title/body
	data := buildCommonData("event", "event_24hr", eventIDStr, map[string]string{
		"notification_type": strconv.Itoa(NotiTypeEvent24), // 5
		"title":             title,
		"body":              body,
	})

	tokenRaw, err := repositories.GetUserTokenNotificationInEvent24Hr(eventIDStr)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้ใช้/โทเคนไม่สำเร็จ: %v", err)
	}
	tokens := dedupeNonEmpty(tokenRaw)
	if len(tokens) == 0 {
		return fmt.Errorf("ไม่พบ token สำหรับการแจ้งเตือนอีเวนต์ %s", eventIDStr)
	}

	userData, err := repositories.GetUserIDsInEvent(eventID)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้เข้าร่วมอีเวนต์ล้มเหลว: %v", err)
	}
	if len(userData) == 0 {
		return fmt.Errorf("ไม่พบผู้ใช้ในอีเวนต์ %s", eventIDStr)
	}

	now := time.Now()
	notis := make([]models.Notification, 0, len(userData))
	for _, uid := range userData {
		notis = append(notis, models.Notification{
			UserID:           int(uid.ID),
			EventID:          &eventID,
			Title:            title,
			Detail:           body,
			NotificationType: NotiTypeEvent24,
			NotificationTime: now,
			IsRead:           false,
		})
	}
	if err := repositories.CreateNotificationsLoopTx(notis); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	const chunk = 500
	ctx := context.Background()
	success, failure := 0, 0

	for i := 0; i < len(tokens); i += chunk {
		end := i + chunk
		if end > len(tokens) {
			end = len(tokens)
		}

		// ✅ data-only (ไม่มี Notification)
		msg := &messaging.MulticastMessage{
			Tokens: tokens[i:end],
			Data:   data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				// ChannelID ใช้ในฝั่ง local-noti; ที่นี่ส่งแค่ data ก็พอ
			},
			APNS: &messaging.APNSConfig{
				// ✅ data-only (silent) บน iOS ให้ปลุกแอปไปทำ local noti เอง
				Headers: map[string]string{
					"apns-push-type": "background",
					"apns-priority":  "5",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{ContentAvailable: true},
				},
			},
		}

		resp, err := firebase.MessagingClient.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("❌ FCM multicast error (%d-%d): %v\n", i, end, err)
			continue
		}
		success += resp.SuccessCount
		failure += resp.FailureCount
	}

	log.Printf("✅ [24HR] อีเวนต์ '%s' ส่งสำเร็จ=%d ล้มเหลว=%d ผู้รับ=%d ชุด=%d\n",
		eventData.EventName, success, failure, len(tokens), (len(tokens)+chunk-1)/chunk)
	return nil
}

func UpdateIsReadNotification(nid string) error {
	if err := repositories.UpdateIsReadNotification(nid); err != nil {
		return errors.New("ไม่สามารถอัปเดตสถานะการอ่านได้")
	}

	return nil
}
