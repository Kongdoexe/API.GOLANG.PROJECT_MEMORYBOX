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

// ======================================================
// 3) Event will start within 24 hours
// ======================================================
func SendNotificationEvent24Hr(eventIDStr string) error {
	// --- parse event id ---
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fmt.Errorf("invalid event ID: %v", err)
	}

	// --- ดึงข้อมูลอีเวนต์ ---
	eventData, err := repositories.EventFindByID(eventID)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลอีเวนต์ไม่สำเร็จ: %v", err)
	}

	title := "การแจ้งเตือนอีเวนต์จะเริ่มภายใน 24 ชั่วโมง"
	body := fmt.Sprintf("อีเวนต์ %s กำลังจะเริ่มภายใน 24 ชั่วโมง", eventData.EventName)

	data := map[string]string{
		"type":         "event",
		"subtype":      "event_24hr",
		"event_id":     eventIDStr,
		"click_action": "FLUTTER_NOTIFICATION_CLICK",
	}

	// --- ดึง tokens สำหรับกลุ่มเป้าหมายในอีเวนต์ (ภายใน 24 ชม.) ---
	// ฟังก์ชันนี้สมมติว่าคืนมาเป็นรายการ token ตรง ๆ
	tokenRaw, err := repositories.GetUserTokenNotificationInEvent24Hr(eventIDStr)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้ใช้/โทเคนไม่สำเร็จ: %v", err)
	}
	tokens := dedupeNonEmpty(tokenRaw)
	if len(tokens) == 0 {
		return fmt.Errorf("ไม่พบ token สำหรับการแจ้งเตือนอีเวนต์ %s", eventIDStr)
	}

	// --- ดึง user_id ทั้งหมดในอีเวนต์ เพื่อบันทึกลงตาราง notification ---
	userData, err := repositories.GetUserIDsInEvent(eventID)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้เข้าร่วมอีเวนต์ล้มเหลว: %v", err)
	}
	if len(userData) == 0 {
		return fmt.Errorf("ไม่พบผู้ใช้ในอีเวนต์ %s", eventIDStr)
	}

	// --- เตรียม notification บันทึกลงฐานข้อมูล ---
	now := time.Now()
	notis := make([]models.Notification, 0, len(userData))
	for _, uid := range userData {
		if len(userData) <= 0 {
			continue
		}
		notis = append(notis, models.Notification{
			UserID:           int(uid.ID),
			EventID:          &eventID,
			Title:            title,
			Detail:           body,
			NotificationType: NotiTypeEvent24, // = 5
			NotificationTime: now,
			IsRead:           false,
		})
	}
	if len(notis) == 0 {
		return fmt.Errorf("ไม่มี notification ที่ valid ให้บันทึก")
	}
	if err := repositories.CreateNotificationsLoopTx(notis); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	// --- ส่ง FCM แบบ chunk ---
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
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "event_channel",
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{"apns-priority": "10"},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{Category: "chat"},
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

// ======================================================
// 4) User Join to Event (ส่งแจ้งเตือนเมื่อมีคนเข้าร่วม)
// ======================================================
func SendNotificationEventUserJoin(eventIDStr, userIDJoin string) error {
	// --- parse event id ---
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return fmt.Errorf("invalid event ID: %v", err)
	}

	// --- ดึงข้อมูลอีเวนต์ ---
	eventData, err := repositories.EventFindByID(eventID)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลอีเวนต์ไม่สำเร็จ: %v", err)
	}

	// --- ดึงข้อมูลผู้ใช้ที่เพิ่ง join (เพื่อใส่ชื่อในข้อความ) ---
	userData, err := repositories.UserFindByID(userIDJoin)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลผู้ใช้ไม่สำเร็จ: %v", err)
	}

	// --- จำนวนผู้เข้าร่วมทั้งหมด (ไว้ใส่ในข้อความ) ---
	lengthJoin, err := repositories.GetCountUserJoin(eventIDStr)
	if err != nil {
		return fmt.Errorf("ดึงข้อความผู้ใช้ Join ไม่สำเร็จ: %v", err)
	}

	title := "การแจ้งเตือนการเข้าร่วม"
	body := fmt.Sprintf("ผู้ใช้ %s เข้าร่วมอีเวนต์ %s และอีก %d คนเข้าร่วมอีเวนต์แล้ว", userData.Name, eventData.EventName, lengthJoin)

	data := map[string]string{
		"type":         "event",
		"subtype":      "joined",
		"event_id":     eventIDStr,
		"click_action": "FLUTTER_NOTIFICATION_CLICK",
	}

	// --- ดึงรายชื่อผู้ใช้ในอีเวนต์ (มีทั้ง user_id และ token_notification) ---
	userDataL, err := repositories.GetUserIDsInEvent(eventID)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้เข้าร่วมอีเวนต์ล้มเหลว: %v", err)
	}
	if len(userDataL) == 0 {
		return fmt.Errorf("ไม่พบผู้ใช้ในอีเวนต์ %s", eventIDStr)
	}

	// --- เตรียม notifications สำหรับบันทึกลงตาราง ---
	now := time.Now()
	notis := make([]models.Notification, 0, len(userDataL))
	for _, u := range userDataL {
		notis = append(notis, models.Notification{
			UserID:           int(u.ID), // <-- ปรับ field ให้ตรงกับ struct ที่คืนมา
			EventID:          &eventID,
			Title:            title,
			Detail:           body,
			NotificationType: NotiTypeJoin, // = 5
			NotificationTime: now,
			IsRead:           false,
		})
	}
	if err := repositories.CreateNotificationsLoopTx(notis); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	// --- รวบรวม token ที่ไม่ว่าง + ตัดซ้ำ ---
	tokenSet := make(map[string]struct{}, len(userDataL))
	tokens := make([]string, 0, len(userDataL))
	for _, u := range userDataL {
		t := strings.TrimSpace(u.TokenNotification) // <-- ปรับ field ให้ตรงกับของคุณ
		if t == "" {
			continue
		}
		if _, ok := tokenSet[t]; !ok {
			tokenSet[t] = struct{}{}
			tokens = append(tokens, t)
		}
	}

	// ถ้าไม่มี token ก็จบที่บันทึกใน DB อย่างเดียว
	if len(tokens) == 0 {
		log.Printf("ℹ️ ไม่มี FCM token สำหรับอีเวนต์ '%s' (event_id=%d) บันทึก DB แล้ว", eventData.EventName, eventID)
		return nil
	}

	// --- ส่ง FCM แบบ chunk ---
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
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "event_channel",
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{"apns-priority": "10"},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{Category: "chat"},
				},
			},
		}

		resp, err := firebase.MessagingClient.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("❌ FCM multicast error (%d-%d): %v\n", i, end, err)
			// นับเป็นล้มเหลวทั้งหมดของชุดนี้ เพื่อให้สถิติตรง
			failure += (end - i)
			continue
		}
		success += resp.SuccessCount
		failure += resp.FailureCount
	}

	log.Printf("✅ [JOIN] อีเวนต์ '%s' ส่งสำเร็จ=%d ล้มเหลว=%d ผู้รับ(ไม่ซ้ำ)=%d ชุด=%d\n",
		eventData.EventName, success, failure, len(tokens), (len(tokens)+chunk-1)/chunk)

	return nil
}

// ======================================================
// 1) Invite to event
// ======================================================
func SendNotificationEvent(req request.RequestSendNotiEvent) error {
	eventID, err := strconv.Atoi(req.EventID)
	if err != nil {
		return fmt.Errorf("invalid event ID: %v", err)
	}

	// ดึงข้อมูลอีเวนต์
	eventData, err := repositories.EventFindByID(eventID)
	if err != nil {
		return fmt.Errorf("ดึงข้อมูลอีเวนต์ไม่สำเร็จ: %v", err)
	}

	// ดึงชื่อผู้เชิญ (ถ้ามี)
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

	data := map[string]string{
		"type":         "event",
		"subtype":      "invited",
		"event_id":     req.EventID,
		"inviter_id":   req.InviterID,
		"click_action": "FLUTTER_NOTIFICATION_CLICK",
	}

	// Tokens
	tokensRaw, err := repositories.GetTokensByUserIDs(req.UserIDs)
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

	// บันทึกลง DB
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
	if len(notis) == 0 {
		return fmt.Errorf("ไม่มี notification ที่ valid ให้บันทึก")
	}
	if err := repositories.CreateNotificationsLoopTx(notis); err != nil {
		return fmt.Errorf("create notifications bulk: %v", err)
	}

	// ส่ง FCM
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
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "event_channel",
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{"apns-priority": "10"},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{Category: "chat"},
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

// ======================================================
// 2) New chat message
// ======================================================
func SendNotificationChat(req request.RequestSendNotiChat) error {
	data := map[string]string{
		"type":         "chat",
		"subtype":      "message",
		"event_id":     req.EventID,
		"click_action": "FLUTTER_NOTIFICATION_CLICK",
	}

	// parse ids
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

	// ดึงผู้เข้าร่วมทั้งหมด (ยกเว้นผู้ส่งเอง)
	userData, err := repositories.GetUserIDsInEvent(eventIdInt)
	if err != nil {
		return fmt.Errorf("ดึงรายชื่อผู้เข้าร่วมอีเวนต์ล้มเหลว: %v", err)
	}
	seenUID := make(map[int]struct{}, len(userData))
	filteredUIDs := make([]int, 0, len(userData))
	for _, uid := range userData {
		if len(userData) == 0 || int(uid.ID) == userIdInt {
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

	// บันทึกลง DB
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

	// ดึง tokens
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

	// ส่ง FCM
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
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "chat_channel",
				},
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{"apns-priority": "10"},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{Category: "chat"},
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

func UpdateIsReadNotification(nid string) error {
	if err := repositories.UpdateIsReadNotification(nid); err != nil {
		return errors.New("ไม่สามารถอัปเดตสถานะการอ่านได้")
	}

	return nil
}
