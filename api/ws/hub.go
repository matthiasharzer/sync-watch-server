package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/matthiasharzer/sync-watch-server/logging"
	"github.com/matthiasharzer/sync-watch-server/rooms"
)

type HubRoom struct {
	room    *rooms.Room
	clients map[*websocket.Conn]bool
	mutex   *sync.RWMutex
}

func (h *HubRoom) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logging.Warn("failed to send message to client: %v", err)
		}
	}
}

func (h *HubRoom) addClient(conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[conn] = true
}

func (h *HubRoom) removeClient(conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.clients, conn)
}

type Hub struct {
	rooms map[string]*HubRoom
	mutex *sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		rooms: make(map[string]*HubRoom),
		mutex: &sync.RWMutex{},
	}
}

func (h *Hub) getOrCreateHubRoom(room *rooms.Room) (*HubRoom, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	hubRoom, exists := h.rooms[room.ID]
	if !exists {
		hubRoom = &HubRoom{
			room:    room,
			clients: make(map[*websocket.Conn]bool),
			mutex:   &sync.RWMutex{},
		}
		h.rooms[room.ID] = hubRoom
	}

	return hubRoom, nil
}
