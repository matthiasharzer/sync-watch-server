package subscribe

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/matthiasharzer/sync-watch-server/server"
	"github.com/matthiasharzer/sync-watch-server/util/httputil"
)

func sendNdJsonLine(w http.ResponseWriter, line any) error {
	lineBytes, err := json.Marshal(line)
	if err != nil {
		return errors.New("failed to marshal line")
	}
	ndjsonLine := append(lineBytes, '\n')

	_, err = w.Write(ndjsonLine)
	if err != nil {
		return errors.New("failed to write ndjson line")
	}
	return nil
}

type connection struct {
	w          http.ResponseWriter
	flusher    http.Flusher
	controller *http.ResponseController
}

func newConnection(w http.ResponseWriter) (*connection, error) {
	responseController := http.NewResponseController(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("flusher not supported")
	}

	return &connection{
		w:          w,
		flusher:    flusher,
		controller: responseController,
	}, nil

}

func (c connection) flush() {
	c.flusher.Flush()
	_ = c.controller.SetWriteDeadline(time.Now().Add(10 * time.Second))
}

func Handler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := newConnection(w)
		if err != nil {
			http.Error(w, "failed to create connection", http.StatusInternalServerError)
			return
		}

		requestBody, err := httputil.ParseRequestBody[RequestBody](w, r)
		if err != nil {
			http.Error(w, "failed to parse request body", http.StatusBadRequest)
			return
		}

		_, exists := s.GetRoom(requestBody.RoomID)
		if !exists {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}

		heartbeatTicker := time.NewTicker(5 * time.Second)
		defer heartbeatTicker.Stop()

		w.Header().Add("Content-Type", "application/x-ndjson")

		updates := s.SubscribeToRoom(requestBody.RoomID, r.Context())

		conn.flush()

	out:
		for {
			select {
			case <-r.Context().Done():
				break out
			case <-heartbeatTicker.C:
				line := NewHeartbeatResponse()
				err := sendNdJsonLine(w, line)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					break out
				}
				conn.flush()
			case room, ok := <-updates:
				if !ok {
					break out
				}

				line := NewRoomUpdateResponse(room.Progress, room.State)
				err := sendNdJsonLine(w, line)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					break out
				}
				conn.flush()
			}
		}
	}
}
