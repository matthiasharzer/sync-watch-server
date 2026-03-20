package run

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matthiasharzer/sync-watch-server/api"
	"github.com/matthiasharzer/sync-watch-server/api/createroom"
	"github.com/matthiasharzer/sync-watch-server/api/subscribe"
	"github.com/matthiasharzer/sync-watch-server/logging"
	"github.com/spf13/cobra"
)

const roomCleanupInterval = 10 * time.Minute
const roomInactivityThreshold = 2 * time.Hour

var port int
var host string

func init() {
	Command.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	Command.Flags().StringVarP(&host, "host", "", "", "Host to listen on (default: all interfaces)")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var Command = &cobra.Command{
	Use: "run",
	RunE: func(_ *cobra.Command, _ []string) error {
		quarterMaster := api.NewQuartermaster()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		mux.HandleFunc("POST /api/v1/create-room", createroom.Handler(quarterMaster))
		mux.HandleFunc("/api/v1/ws", subscribe.Handler(quarterMaster))

		corsMux := corsMiddleware(mux)

		finished := make(chan struct{})

		go func() {
			ticker := time.NewTicker(roomCleanupInterval)
			defer ticker.Stop()

			for {
				select {
				case <-finished:
					return
				case <-ticker.C:
					quarterMaster.CleanupRooms(roomInactivityThreshold)
				}
			}
		}()

		addr := fmt.Sprintf("%s:%d", host, port)

		logging.Info("starting sync-watch-server", "host", host, "port", port)
		err := http.ListenAndServe(
			addr,
			corsMux,
		)

		finished <- struct{}{}

		return err
	},
}
