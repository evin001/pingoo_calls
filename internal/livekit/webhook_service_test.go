package livekit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	lkproto "github.com/livekit/protocol/livekit"
	"pingoo_calls/internal/config"
)

func TestWebhookServiceForwardSendsNormalizedEventToElixir(t *testing.T) {
	var received NormalizedWebhookEvent
	var receivedSecret string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSecret = r.Header.Get(internalSecretHeader)
		if r.URL.Path != "/internal/livekit/events" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewWebhookService(&config.Config{
		LiveKitAPIKey:           "devkey",
		LiveKitAPISecret:        "secret",
		PingooInternalSecret:    "internal-secret",
		PingooServerInternalURL: server.URL,
	})

	event := &NormalizedWebhookEvent{
		Event:       "participant_joined",
		RoomName:    "call_123",
		Participant: "user:1:device:ios",
	}

	if err := service.Forward(context.Background(), event); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}

	if receivedSecret != "internal-secret" {
		t.Fatalf("secret header = %q", receivedSecret)
	}
	if received != *event {
		t.Fatalf("received = %#v, want %#v", received, *event)
	}
}

func TestWebhookServiceForwardFailsOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewWebhookService(&config.Config{
		LiveKitAPIKey:           "devkey",
		LiveKitAPISecret:        "secret",
		PingooInternalSecret:    "internal-secret",
		PingooServerInternalURL: server.URL,
	})

	err := service.Forward(context.Background(), &NormalizedWebhookEvent{
		Event: "room_finished",
	})
	if err == nil {
		t.Fatal("Forward returned nil error")
	}
}

func TestNormalizeWebhookEventMarksUnsupportedEventsIgnored(t *testing.T) {
	event := normalizeWebhookEvent(&lkproto.WebhookEvent{
		Event: "track_published",
		Room: &lkproto.Room{
			Name: "call_123",
		},
		Participant: &lkproto.ParticipantInfo{
			Identity: "user:1:device:ios",
		},
	})

	if event == nil {
		t.Fatal("normalizeWebhookEvent returned nil")
	}
	if !event.Ignored {
		t.Fatal("Ignored is false")
	}
	if event.Event != "track_published" {
		t.Fatalf("Event = %q", event.Event)
	}
	if event.RoomName != "call_123" {
		t.Fatalf("RoomName = %q", event.RoomName)
	}
	if event.Participant != "user:1:device:ios" {
		t.Fatalf("Participant = %q", event.Participant)
	}
}
