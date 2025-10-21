package request

type RequestTokenNotification struct {
	UserID            string `json:"user_id"`
	TokenNotification string `json:"token_notification"`
}

type RequestSendNotiChat struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}

type RequestSendNotiEvent struct {
	EventID   string   `json:"event_id"`
	UserIDs   []string `json:"user_ids"`
	InviterID string   `json:"inviter_id"`
}

type RequestSendNotificaiton struct {
	EventID string `json:"event_id"`
	Type    uint   `json:"type"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}
