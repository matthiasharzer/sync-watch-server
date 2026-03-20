package subscribe

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/matthiasharzer/sync-watch-server/api"
	"github.com/matthiasharzer/sync-watch-server/logging"
)

const readLimit = 1024 * 1024 // 1 MiB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Handler(q *api.Quartermaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error("failed to upgrade connection", "err", err)
			return
		}
		defer conn.Close()

		conn.SetReadLimit(readLimit)

		roomID := r.URL.Query().Get("roomId")
		if roomID == "" {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("missing roomId query parameter"))
			return
		}

		room, exists := q.GetRoom(roomID)
		if !exists {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("room not found"))
			return
		}

		room.AddClient(conn)
		defer room.RemoveClient(conn)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				logging.Warn("failed to read message from client", "err", err)
				break
			}

			if messageType != websocket.TextMessage {
				logging.Warn("received non-text message from client, ignoring")
				continue
			}

			room.BroadcastMessage(message)
		}
	}
}
