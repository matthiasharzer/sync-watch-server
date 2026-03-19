package server

import (
	"context"
	"slices"
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomID() string {
	var seededRand = time.Now().UnixNano()
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[seededRand%int64(len(charset))]
		seededRand /= int64(len(charset))
	}
	return string(b)
}

type Server struct {
	mutex         *sync.RWMutex
	Rooms         map[string]*Room
	subscriptions map[string][]*Observer[*Room]
}

func New() *Server {
	return &Server{
		mutex:         &sync.RWMutex{},
		Rooms:         make(map[string]*Room),
		subscriptions: make(map[string][]*Observer[*Room]),
	}
}

func (s *Server) CleanupObservers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, subs := range s.subscriptions {
		subs = slices.DeleteFunc(subs, func(sub *Observer[*Room]) bool {
			return sub.IsCanceled()
		})
	}
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
		ID:              id,
		Progress:        0,
		State:           PlayerStatePaused,
		LastTimeUpdated: time.Now().Unix(),
	}
	s.Rooms[id] = room
	return room
}

func (s *Server) updateProgress(id string, progress float64) *Room {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.Rooms[id]
	if !ok {
		return nil
	}
	room.Progress = progress
	room.LastTimeUpdated = time.Now().Unix()
	return room
}

func (s *Server) UpdateProgress(id string, progress float64) {
	room := s.updateProgress(id, progress)
	if room == nil {
		return
	}
	s.notifySubscribers(room)
}

func (s *Server) updateState(id string, state PlayerState) *Room {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.Rooms[id]
	if !ok {
		return nil
	}
	room.State = state
	room.LastTimeUpdated = time.Now().Unix()
	return room
}

func (s *Server) UpdateState(id string, state PlayerState) {
	room := s.updateState(id, state)
	if room == nil {
		return
	}
	s.notifySubscribers(room)
}

func (s *Server) addObserver(id string, observer *Observer[*Room]) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.subscriptions[id] = append(s.subscriptions[id], observer)
}
func (s *Server) SubscribeToRoom(id string, ctx context.Context) <-chan *Room {
	ch := make(chan *Room)

	observer := NewObserver[*Room](ctx)
	s.addObserver(id, observer)

	go func() {
		defer close(ch)
		r, ok := s.GetRoom(id)
		if ok {
			ch <- r
		}

		for room := range observer.Range() {
			ch <- room
		}
	}()

	return ch
}

func (s *Server) notifySubscribers(room *Room) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	subs, ok := s.subscriptions[room.ID]
	if !ok {
		return
	}

	for _, sub := range subs {
		sub.Send(room)
	}
}
