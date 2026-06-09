package http

import (
	"encoding/json"
	"log"
	nethttp "net/http"
)

func WriteJSON(w nethttp.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}

func WriteError(w nethttp.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{
		"error": message,
	})
}
