package run

import (
	"fmt"
	"net/http"
	"time"

	"github.com/matthiasharzer/sync-watch-server/api/createroom"
	"github.com/matthiasharzer/sync-watch-server/api/setplaystate"
	"github.com/matthiasharzer/sync-watch-server/api/setprogress"
	"github.com/matthiasharzer/sync-watch-server/api/subscribe"
	"github.com/matthiasharzer/sync-watch-server/logging"
	"github.com/matthiasharzer/sync-watch-server/server"
	"github.com/spf13/cobra"
)

var port int
var host string

func init() {
	Command.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	Command.Flags().StringVarP(&host, "host", "", "", "Host to listen on (default: all interfaces)")
}

var Command = &cobra.Command{
	Use: "run",
	RunE: func(_ *cobra.Command, _ []string) error {
		syncServer := server.New()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		mux.HandleFunc("POST /create-room", createroom.Handler(syncServer))
		mux.HandleFunc("POST /subscribe", subscribe.Handler(syncServer))
		mux.HandleFunc("POST /set-progress", setprogress.Handler(syncServer))
		mux.HandleFunc("POST /set-state", setplaystate.Handler(syncServer))

		finished := make(chan struct{})

		go func() {
			ticker := time.NewTicker(10 * time.Minute)
			for {
				select {
				case <-finished:
					return
				case <-ticker.C:
					syncServer.CleanupObservers()
				}
			}
		}()

		addr := fmt.Sprintf("%s:%d", host, port)

		logging.Info("starting sync-watch-server", "host", host, "port", port)
		err := http.ListenAndServe(
			addr,
			mux,
		)

		finished <- struct{}{}

		return err
	},
}
