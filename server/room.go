package server

type PlayerState string

const (
	PlayerStatePaused  PlayerState = "paused"
	PlayerStatePlaying PlayerState = "playing"
)

type Room struct {
	ID              string
	Progress        float64
	State           PlayerState
	LastTimeUpdated int64
}
