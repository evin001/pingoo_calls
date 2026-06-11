package livekit

import (
	"fmt"
	"strings"
	"time"

	livekitauth "github.com/livekit/protocol/auth"
	"pingoo_calls/internal/config"
)

const tokenTTL = 10 * time.Minute

type TokenService struct {
	cfg *config.Config
}

type TokenRequest struct {
	CallID    string `json:"call_id"`
	UserID    string `json:"user_id"`
	DeviceID  string `json:"device_id"`
	MediaKind string `json:"media_kind"`
}

type TokenResponse struct {
	LiveKitURL string `json:"livekit_url"`
	RoomName   string `json:"room_name"`
	Token      string `json:"token"`
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{
		cfg: cfg,
	}
}

func (s *TokenService) Generate(req TokenRequest) (*TokenResponse, error) {
	if !validIdentifier(req.CallID) {
		return nil, fmt.Errorf("call_id is required")
	}

	if !validIdentifier(req.UserID) {
		return nil, fmt.Errorf("user_id is required")
	}

	if !validIdentifier(req.DeviceID) {
		return nil, fmt.Errorf("device_id is required")
	}

	if req.MediaKind != "audio" {
		return nil, fmt.Errorf("unsupported media_kind: %s", req.MediaKind)
	}

	roomName := RoomName(req.CallID)
	identity := "user:" + req.UserID + ":device:" + req.DeviceID

	accessToken := livekitauth.NewAccessToken(
		s.cfg.LiveKitAPIKey,
		s.cfg.LiveKitAPISecret,
	)

	canPublish := true
	canSubscribe := true
	canPublishData := true

	grant := &livekitauth.VideoGrant{
		RoomJoin:          true,
		Room:              roomName,
		CanPublish:        &canPublish,
		CanSubscribe:      &canSubscribe,
		CanPublishData:    &canPublishData,
		CanPublishSources: []string{"microphone"},
	}

	token, err := accessToken.
		SetIdentity(identity).
		SetVideoGrant(grant).
		SetValidFor(tokenTTL).
		ToJWT()

	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		LiveKitURL: s.cfg.LiveKitPublicURL,
		RoomName:   roomName,
		Token:      token,
	}, nil
}

func validIdentifier(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}

	return !strings.ContainsFunc(value, func(r rune) bool {
		return r <= 0x20 || r == ':' || r == '/'
	})
}
