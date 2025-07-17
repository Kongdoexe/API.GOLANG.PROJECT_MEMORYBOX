package response

import "time"

type MediaResponse struct {
	MediaID    uint      `json:"mediaID"`
	ImageURL   string    `json:"imageURL"`
	UploadedAt time.Time `json:"uploadedAt"`
	TakenAt    time.Time `json:"takenAt"`
	UploadedBy struct {
		UserID   uint   `json:"userID"`
		UserName string `json:"userName"`
	} `json:"uploadedBy"`
}
