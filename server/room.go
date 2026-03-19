package server

type PlayerState string

const (
	PlayerStatePaused  PlayerState = "paused"
	PlayerStatePlaying PlayerState = "playing"
)

type Room struct {
	ID       string      `json:"id"`
	Progress float64     `json:"progress"`
	State    PlayerState `json:"state"`
}
