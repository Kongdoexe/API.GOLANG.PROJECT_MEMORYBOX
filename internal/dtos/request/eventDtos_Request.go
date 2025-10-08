package request

import "time"

type FavoriteReq struct {
	EID    string    `json:"eid"`
	UID    string    `json:"uid"`
	Date   time.Time `json:"date"`
	Status bool      `json:"status"` // true Favorite, false UnFavorite
}
