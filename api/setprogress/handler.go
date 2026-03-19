package setprogress

import (
	"net/http"

	"github.com/matthiasharzer/sync-watch-server/server"
	"github.com/matthiasharzer/sync-watch-server/util/httputil"
)

func Handler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := httputil.ParseRequestBody[RequestBody](w, r)
		if err != nil {
			http.Error(w, "failed to parse request body", http.StatusBadRequest)
			return
		}

		s.UpdateProgress(requestBody.RoomID, requestBody.Progress)
		w.WriteHeader(http.StatusOK)
	}
}
