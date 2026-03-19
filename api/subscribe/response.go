package subscribe

import "github.com/matthiasharzer/sync-watch-server/server"

type ResponseBodyLine[T any] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type HeartbeatResponseLine = ResponseBodyLine[struct{}]

type RoomUpdateData struct {
	Progress float64            `json:"progress"`
	State    server.PlayerState `json:"state"`
}

type RoomUpdateResponseLine = ResponseBodyLine[RoomUpdateData]

func NewHeartbeatResponse() HeartbeatResponseLine {
	return HeartbeatResponseLine{
		Type: "heartbeat",
		Data: struct{}{},
	}
}

func NewRoomUpdateResponse(progress float64, state server.PlayerState) RoomUpdateResponseLine {
	return RoomUpdateResponseLine{
		Type: "roomUpdate",
		Data: RoomUpdateData{
			Progress: progress,
			State:    state,
		},
	}
}
