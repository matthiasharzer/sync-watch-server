package rooms

import (
	"sync"
	"time"

	"github.com/matthiasharzer/sync-watch-server/util/randomutil"
)

const roomIdLength = 8

type Quartermaster struct {
	rooms map[string]*Room
	mutex *sync.RWMutex
}

func NewQuartermaster() *Quartermaster {
	return &Quartermaster{
		rooms: make(map[string]*Room),
		mutex: &sync.RWMutex{},
	}
}

func (q *Quartermaster) getNextID() string {
	for {
		id := randomutil.RandomString(roomIdLength)
		if _, exists := q.rooms[id]; !exists {
			return id
		}
	}
}

func (q *Quartermaster) CreateRoom() *Room {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	room := &Room{
		ID:              q.getNextID(),
		LastInteraction: time.Now().UTC().Unix(),
	}

	q.rooms[room.ID] = room

	return room
}

func (q *Quartermaster) GetRoom(id string) (*Room, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	room, ok := q.rooms[id]
	return room, ok
}

func (q *Quartermaster) InteractWithRoom(id string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if room, ok := q.rooms[id]; ok {
		room.LastInteraction = time.Now().UTC().Unix()
	}
}

func (q *Quartermaster) CleanupRooms(timeout time.Duration) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	now := time.Now().UTC().Unix()

	for id, room := range q.rooms {
		if now-room.LastInteraction > int64(timeout.Seconds()) {
			delete(q.rooms, id)
		}
	}
}
