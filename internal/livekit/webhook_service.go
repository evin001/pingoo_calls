package livekit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	nethttp "net/http"
	"strings"
	"time"

	livekitauth "github.com/livekit/protocol/auth"
	lkproto "github.com/livekit/protocol/livekit"
	lkwebhook "github.com/livekit/protocol/webhook"
	"pingoo_calls/internal/config"
)

const internalSecretHeader = "X-Pingoo-Internal-Secret"

type WebhookService struct {
	keyProvider    livekitauth.KeyProvider
	callbackURL    string
	internalSecret string
	httpClient     *nethttp.Client
}

type NormalizedWebhookEvent struct {
	Event       string `json:"event"`
	RoomName    string `json:"room_name,omitempty"`
	Participant string `json:"participant,omitempty"`
	Ignored     bool   `json:"-"`
}

func NewWebhookService(cfg *config.Config) *WebhookService {
	return &WebhookService{
		keyProvider: livekitauth.NewSimpleKeyProvider(
			cfg.LiveKitAPIKey,
			cfg.LiveKitAPISecret,
		),
		callbackURL:    strings.TrimRight(cfg.PingooServerInternalURL, "/") + "/internal/livekit/events",
		internalSecret: cfg.PingooInternalSecret,
		httpClient:     nethttp.DefaultClient,
	}
}

func (s *WebhookService) Receive(r *nethttp.Request) (*NormalizedWebhookEvent, error) {
	event, err := lkwebhook.ReceiveWebhookEvent(r, s.keyProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid livekit webhook: %w", err)
	}

	return normalizeWebhookEvent(event), nil
}

func (s *WebhookService) Forward(ctx context.Context, event *NormalizedWebhookEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to encode webhook event: %w", err)
	}

	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, s.callbackURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build webhook callback request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalSecretHeader, s.internalSecret)

	startedAt := time.Now()
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf(
			"livekit webhook forward transport failed: event=%s room=%s participant=%s callback_url=%s duration=%s error=%v",
			event.Event,
			event.RoomName,
			event.Participant,
			s.callbackURL,
			time.Since(startedAt),
			err,
		)
		return fmt.Errorf("failed to forward webhook event: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startedAt)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		log.Printf(
			"livekit webhook forward rejected: event=%s room=%s participant=%s callback_url=%s status=%d duration=%s body=%s",
			event.Event,
			event.RoomName,
			event.Participant,
			s.callbackURL,
			resp.StatusCode,
			duration,
			string(respBody),
		)
		return fmt.Errorf("webhook callback failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	log.Printf(
		"livekit webhook forwarded: event=%s room=%s participant=%s callback_url=%s status=%d duration=%s",
		event.Event,
		event.RoomName,
		event.Participant,
		s.callbackURL,
		resp.StatusCode,
		duration,
	)

	return nil
}

func normalizeWebhookEvent(event *lkproto.WebhookEvent) *NormalizedWebhookEvent {
	normalized := &NormalizedWebhookEvent{
		Event: event.Event,
	}

	if event.Room != nil {
		normalized.RoomName = event.Room.Name
	}

	if event.Participant != nil {
		normalized.Participant = event.Participant.Identity
	}

	normalizedEvent, ok := normalizeEventName(event.Event)
	if !ok {
		normalized.Ignored = true
		return normalized
	}

	normalized.Event = normalizedEvent
	return normalized
}

func normalizeEventName(event string) (string, bool) {
	switch event {
	case "participant_joined":
		return "participant_joined", true
	case "participant_left":
		return "participant_left", true
	case "room_finished":
		return "room_finished", true
	default:
		return "", false
	}
}
