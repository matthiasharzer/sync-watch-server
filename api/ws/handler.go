package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/matthiasharzer/sync-watch-server/logging"
	"github.com/matthiasharzer/sync-watch-server/rooms"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Handler(q *rooms.Quartermaster) http.HandlerFunc {
	hub := newHub()
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error("failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		roomId := r.URL.Query().Get("roomId")
		if roomId == "" {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("missing roomId query parameter"))
			return
		}

		room, exists := q.GetRoom(roomId)
		if !exists {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("room not found"))
			return
		}

		hubRoom, err := hub.getOrCreateHubRoom(room)
		if err != nil {
			logging.Error("failed to get or create hub room: %v", err)
			return
		}

		hubRoom.addClient(conn)
		defer hubRoom.removeClient(conn)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				logging.Warn("failed to read message from client: %v", err)
				break
			}

			if messageType != websocket.TextMessage {
				logging.Warn("received non-text message from client, ignoring")
				continue
			}

			hubRoom.broadcastMessage(message)
		}
	}
}
