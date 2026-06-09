package main

import (
	"log"
	"net/http"
	"pingoo_calls/internal/auth"

	"pingoo_calls/internal/config"
	pingoohttp "pingoo_calls/internal/http"
	pingoolivekit "pingoo_calls/internal/livekit"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		pingoohttp.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	internalMux := http.NewServeMux()

	tokenService := pingoolivekit.NewTokenService(cfg)
	roomService := pingoolivekit.NewRoomService(cfg)
	liveKitHandlers := pingoohttp.NewLiveKitHandlers(tokenService, roomService)

	internalMux.HandleFunc("POST /livekit/token", liveKitHandlers.Token)
	internalMux.HandleFunc("POST /livekit/rooms/ensure", liveKitHandlers.EnsureRoom)
	internalMux.HandleFunc("POST /livekit/rooms/end", liveKitHandlers.EndRoom)

	internalMux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		pingoohttp.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	mux.Handle("/internal/", auth.RequireInternalSecret(
		cfg.PingooInternalSecret,
		http.StripPrefix("/internal", internalMux),
	))

	addr := ":" + cfg.Port

	log.Printf("pingoo_calls listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
