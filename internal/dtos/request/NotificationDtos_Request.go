package request

type RequestTokenNotification struct {
	UserID            string `json:"user_id"`
	TokenNotification string `json:"token_notification"`
}

type RequestSendNotificaiton struct {
	EventID string `json:"event_id"`
	Type    uint   `json:"type"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}
