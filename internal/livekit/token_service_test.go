package livekit

import (
	"testing"

	livekitauth "github.com/livekit/protocol/auth"
	"pingoo_calls/internal/config"
)

func TestTokenServiceGenerateAudioOnlyToken(t *testing.T) {
	cfg := &config.Config{
		LiveKitURL:       "ws://internal-livekit.test",
		LiveKitPublicURL: "ws://public-livekit.test",
		LiveKitAPIKey:    "devkey",
		LiveKitAPISecret: "secret",
	}

	service := NewTokenService(cfg)

	resp, err := service.Generate(TokenRequest{
		CallID:    "call-123",
		UserID:    "42",
		DeviceID:  "ios-device",
		MediaKind: "audio",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if resp.LiveKitURL != cfg.LiveKitPublicURL {
		t.Fatalf("LiveKitURL = %q, want %q", resp.LiveKitURL, cfg.LiveKitPublicURL)
	}
	if resp.RoomName != "call_call-123" {
		t.Fatalf("RoomName = %q", resp.RoomName)
	}

	verifier, err := livekitauth.ParseAPIToken(resp.Token)
	if err != nil {
		t.Fatalf("ParseAPIToken returned error: %v", err)
	}
	if verifier.Identity() != "user:42:device:ios-device" {
		t.Fatalf("identity = %q", verifier.Identity())
	}

	_, grants, err := verifier.Verify(cfg.LiveKitAPISecret)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	video := grants.Video
	if video == nil {
		t.Fatal("video grant is nil")
	}
	if !video.RoomJoin {
		t.Fatal("RoomJoin is false")
	}
	if video.Room != "call_call-123" {
		t.Fatalf("grant room = %q", video.Room)
	}
	if video.CanPublish == nil || !*video.CanPublish {
		t.Fatal("CanPublish is not true")
	}
	if video.CanSubscribe == nil || !*video.CanSubscribe {
		t.Fatal("CanSubscribe is not true")
	}
	if len(video.CanPublishSources) != 1 || video.CanPublishSources[0] != "microphone" {
		t.Fatalf("CanPublishSources = %#v", video.CanPublishSources)
	}
}

func TestTokenServiceRejectsUnsupportedOrUnsafeInput(t *testing.T) {
	service := NewTokenService(&config.Config{
		LiveKitURL:       "ws://internal-livekit.test",
		LiveKitPublicURL: "ws://public-livekit.test",
		LiveKitAPIKey:    "devkey",
		LiveKitAPISecret: "secret",
	})

	cases := []TokenRequest{
		{CallID: "", UserID: "1", DeviceID: "d", MediaKind: "audio"},
		{CallID: "call 1", UserID: "1", DeviceID: "d", MediaKind: "audio"},
		{CallID: "call/1", UserID: "1", DeviceID: "d", MediaKind: "audio"},
		{CallID: "call-1", UserID: "1:2", DeviceID: "d", MediaKind: "audio"},
		{CallID: "call-1", UserID: "1", DeviceID: "d", MediaKind: "video"},
	}

	for _, tc := range cases {
		if _, err := service.Generate(tc); err == nil {
			t.Fatalf("Generate(%#v) returned nil error", tc)
		}
	}
}
