package setplaystate

import "github.com/matthiasharzer/sync-watch-server/server"

type RequestBody struct {
	RoomID string             `json:"roomId"`
	State  server.PlayerState `json:"state"`
}
