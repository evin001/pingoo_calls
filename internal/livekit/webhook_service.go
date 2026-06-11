package livekit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"strings"

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

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to forward webhook event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("webhook callback failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

func normalizeWebhookEvent(event *lkproto.WebhookEvent) *NormalizedWebhookEvent {
	normalized := &NormalizedWebhookEvent{
		Event: normalizeEventName(event.Event),
	}

	if event.Room != nil {
		normalized.RoomName = event.Room.Name
	}

	if event.Participant != nil {
		normalized.Participant = event.Participant.Identity
	}

	return normalized
}

func normalizeEventName(event string) string {
	switch event {
	case "participant_joined":
		return "participant_joined"
	case "participant_left":
		return "participant_left"
	case "room_finished":
		return "room_finished"
	default:
		return event
	}
}
