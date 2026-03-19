package createroom

import "github.com/matthiasharzer/sync-watch-server/server"

type ResponseRoom struct {
	ID       string             `json:"id"`
	State    server.PlayerState `json:"state"`
	Progress float64            `json:"progress"`
}

type ResponseBody struct {
	Room ResponseRoom `json:"room"`
}
