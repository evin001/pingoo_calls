package livekit

import (
	"fmt"
	nethttp "net/http"

	"pingoo_calls/internal/config"
	livekitauth "github.com/livekit/protocol/auth"
	lkproto "github.com/livekit/protocol/livekit"
	lkwebhook "github.com/livekit/protocol/webhook"
)

type WebhookService struct {
	keyProvider livekitauth.KeyProvider
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
	}
}

func (s *WebhookService) Receive(r *nethttp.Request) (*NormalizedWebhookEvent, error) {
	event, err := lkwebhook.ReceiveWebhookEvent(r, s.keyProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid livekit webhook: %w", err)
	}

	return normalizeWebhookEvent(event), nil
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

	return normalized
}
