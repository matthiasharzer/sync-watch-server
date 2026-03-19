package subscribe

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/matthiasharzer/sync-watch-server/server"
)

func sendNdJsonLine(w http.ResponseWriter, line any) {
	lineBytes, err := json.Marshal(line)
	if err != nil {
		http.Error(w, "failed to marshal line", http.StatusInternalServerError)
		return
	}
	ndjsonLine := append(lineBytes, '\n')

	_, err = w.Write(ndjsonLine)
	if err != nil {
		http.Error(w, "failed to write ndjson line", http.StatusInternalServerError)
	}
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
			http.Error(w, "failed to created connection", http.StatusInternalServerError)
			return
		}

		limitedBody := http.MaxBytesReader(w, r.Body, int64(1024*10)) // Limit to 10KB

		body, err := io.ReadAll(limitedBody)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			return
		}

		var requestBody RequestBody
		err = json.Unmarshal(body, &requestBody)
		if err != nil {
			http.Error(w, "failed to parse request body", http.StatusBadRequest)
			return
		}

		go func() {
			time.Sleep(8 * time.Second)
			s.UpdateProgress(requestBody.RoomID, 10.5)
		}()

		heartbeatTicker := time.NewTicker(5 * time.Second)
		defer heartbeatTicker.Stop()

		w.Header().Add("Content-Type", "application/x-ndjson")

		updates := s.SubscribeToRoom(requestBody.RoomID, r.Context())

		conn.flush()
	out:
		for {
			select {

			case <-heartbeatTicker.C:
				line := NewHeartbeatResponse()
				sendNdJsonLine(w, line)
				conn.flush()
			case room, ok := <-updates:
				if !ok {
					break out
				}

				line := NewRoomUpdateResponse(room.Progress, room.State)
				sendNdJsonLine(w, line)
				conn.flush()
			}
		}

	}
}
