package api

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/matthiasharzer/sync-watch-server/logging"
	"github.com/matthiasharzer/sync-watch-server/util/randomutil"
)

const roomIDLength = 8

type Client struct {
	conn  *websocket.Conn
	mutex *sync.Mutex
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn:  conn,
		mutex: &sync.Mutex{},
	}
}

func (c *Client) SendMessage(message []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.conn.WriteMessage(websocket.TextMessage, message)
}

type Room struct {
	ID              string
	LastInteraction int64

	clients map[*websocket.Conn]*Client
	mutex   *sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:              id,
		LastInteraction: time.Now().UTC().Unix(),
		clients:         make(map[*websocket.Conn]*Client),
		mutex:           &sync.RWMutex{},
	}
}

func (r *Room) AddClient(conn *websocket.Conn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.clients[conn] = NewClient(conn)
}

func (r *Room) RemoveClient(conn *websocket.Conn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.clients, conn)
}

func (r *Room) broadcastClients() []*Client {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	clients := make([]*Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

func (r *Room) BroadcastMessage(message []byte) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	clients := r.broadcastClients()

	for _, client := range clients {
		err := client.SendMessage(message)
		if err != nil {
			logging.Warn("failed to send message to client: %v", err)
		}
	}

	r.LastInteraction = time.Now().UTC().Unix()
}

func (r *Room) IsExpired(timeout time.Duration) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	now := time.Now().UTC().Unix()
	return now-r.LastInteraction > int64(timeout.Seconds())
}

func (r *Room) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for client := range r.clients {
		err := client.Close()
		if err != nil {
			logging.Warn("failed to close client connection: %v", err)
		}
	}

	r.clients = make(map[*websocket.Conn]*Client)
}

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
		id := randomutil.RandomString(roomIDLength)
		if _, exists := q.rooms[id]; !exists {
			return id
		}
	}
}

func (q *Quartermaster) CreateRoom() *Room {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	room := NewRoom(q.getNextID())

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

	for id, room := range q.rooms {
		if room.IsExpired(timeout) {
			room.Close()
			delete(q.rooms, id)
		}
	}
}
