package livekit

import (
	"context"
	"testing"

	"pingoo_calls/internal/config"
)

func TestRoomServiceRejectsUnsafeCallIDBeforeLiveKitRequest(t *testing.T) {
	service := NewRoomService(&config.Config{
		LiveKitURL:       "ws://localhost:7880",
		LiveKitAPIKey:    "devkey",
		LiveKitAPISecret: "secret",
	})

	for _, callID := range []string{"", "call 1", "call/1", "call:1"} {
		t.Run(callID, func(t *testing.T) {
			if _, err := service.Ensure(context.Background(), EnsureRoomRequest{CallID: callID}); err == nil {
				t.Fatal("Ensure returned nil error")
			}
			if _, err := service.End(context.Background(), EndRoomRequest{CallID: callID}); err == nil {
				t.Fatal("End returned nil error")
			}
		})
	}
}
