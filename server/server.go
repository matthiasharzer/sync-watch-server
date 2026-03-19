package server

import (
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Server struct {
	mutex *sync.RWMutex
	Rooms map[string]*Room
}

func New() *Server {
	return &Server{
		mutex: &sync.RWMutex{},
		Rooms: make(map[string]*Room),
	}
}

func randomID() string {
	var seededRand = time.Now().UnixNano()
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[seededRand%int64(len(charset))]
		seededRand /= int64(len(charset))
	}
	return string(b)
}

func (s *Server) getNextID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for {
		id := randomID()
		if _, exists := s.Rooms[id]; !exists {
			return id
		}
	}
}

func (s *Server) GetRoom(id string) (*Room, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	room, ok := s.Rooms[id]
	return room, ok
}

func (s *Server) CreateRoom() *Room {
	id := s.getNextID()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	room := &Room{
		ID:       id,
		Progress: 0,
		State:    PlayerStatePaused,
	}
	s.Rooms[id] = room
	return room
}

func (s *Server) UpdateProgress(id string, progress float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.Rooms[id]
	if !ok {
		return
	}
	room.Progress = progress
}

func (s *Server) UpdateState(id string, state PlayerState) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.Rooms[id]
	if !ok {
		return
	}
	room.State = state
}
