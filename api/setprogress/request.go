package setprogress

type RequestBody struct {
	RoomID   string  `json:"roomId"`
	Progress float64 `json:"progress"`
}
