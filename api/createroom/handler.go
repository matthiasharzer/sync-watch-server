package createroom

import (
	"encoding/json"
	"net/http"

	"github.com/matthiasharzer/sync-watch-server/api"
)

func Handler(q *api.Quartermaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := q.CreateRoom()

		response := ResponseBody{
			Room: ResponseRoom{
				ID: room.ID,
			},
		}

		valueBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err = w.Write(valueBytes)
		if err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			return
		}
	}
}
