package request

type InsertMessage struct {
	Eid string `json:"eid"`
	Uid string `json:"uid"`
	Msg string `json:"msg"`
}

type GetMessage struct {
	Eid string `json:"eid"`
	Uid string `json:"uid"`
}
