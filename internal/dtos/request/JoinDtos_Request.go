package request

type JoinRequest struct {
	EID string `json:"eid"`
	UID string `json:"uid"`
}

type BlockedJoin struct {
	Eid     string `json:"eid"`
	Uid     string `json:"uid"`
	Current string `json:"current_uid"`
}
