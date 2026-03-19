package setplaystate

import (
	"net/http"

	"github.com/matthiasharzer/sync-watch-server/server"
	"github.com/matthiasharzer/sync-watch-server/util/httputil"
)

func validateState(state server.PlayerState) bool {
	switch state {
	case server.PlayerStatePlaying, server.PlayerStatePaused:
		return true
	default:
		return false
	}
}

func Handler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := httputil.ParseRequestBody[RequestBody](w, r)
		if err != nil {
			http.Error(w, "failed to parse request body", http.StatusBadRequest)
			return
		}

		valid := validateState(requestBody.State)
		if !valid {
			http.Error(w, "invalid state value", http.StatusBadRequest)
			return
		}

		err = s.UpdateState(requestBody.RoomID, requestBody.State)
		if err != nil {
			http.Error(w, "failed to update state", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
