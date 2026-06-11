package livekit

import (
	"context"
	"fmt"
	"strings"
	"time"

	lkproto "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"pingoo_calls/internal/config"
)

type RoomService struct {
	cfg    *config.Config
	client *lksdk.RoomServiceClient
}

type EnsureRoomRequest struct {
	CallID string `json:"call_id"`
}

type EnsureRoomResponse struct {
	RoomName string `json:"room_name"`
}

type EndRoomRequest struct {
	CallID string `json:"call_id"`
}

type EndRoomResponse struct {
	RoomName string `json:"room_name"`
	Ended    bool   `json:"ended"`
}

func NewRoomService(cfg *config.Config) *RoomService {
	client := lksdk.NewRoomServiceClient(
		cfg.LiveKitURL,
		cfg.LiveKitAPIKey,
		cfg.LiveKitAPISecret,
	)

	return &RoomService{
		cfg:    cfg,
		client: client,
	}
}

func (s *RoomService) Ensure(ctx context.Context, req EnsureRoomRequest) (*EnsureRoomResponse, error) {
	if !validIdentifier(req.CallID) {
		return nil, fmt.Errorf("call_id is required")
	}

	roomName := RoomName(req.CallID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.client.CreateRoom(ctx, &lkproto.CreateRoomRequest{
		Name: roomName,
	})

	if err != nil {
		if isRoomAlreadyExistsError(err) {
			return &EnsureRoomResponse{
				RoomName: roomName,
			}, nil
		}

		return nil, fmt.Errorf("failed to create livekit room: %w", err)
	}

	return &EnsureRoomResponse{
		RoomName: roomName,
	}, nil
}

func (s *RoomService) End(ctx context.Context, req EndRoomRequest) (*EndRoomResponse, error) {
	if !validIdentifier(req.CallID) {
		return nil, fmt.Errorf("call_id is required")
	}

	roomName := RoomName(req.CallID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.client.DeleteRoom(ctx, &lkproto.DeleteRoomRequest{
		Room: roomName,
	})

	if err != nil {
		if isRoomNotFoundError(err) {
			return &EndRoomResponse{
				RoomName: roomName,
				Ended:    false,
			}, nil
		}

		return nil, fmt.Errorf("failed to delete livekit room: %w", err)
	}

	return &EndRoomResponse{
		RoomName: roomName,
		Ended:    true,
	}, nil
}

func RoomName(callID string) string {
	return "call_" + callID
}

func isRoomAlreadyExistsError(err error) bool {
	message := strings.ToLower(err.Error())

	return strings.Contains(message, "already exists") ||
		strings.Contains(message, "already_exist")
}

func isRoomNotFoundError(err error) bool {
	message := strings.ToLower(err.Error())

	return strings.Contains(message, "not found") ||
		strings.Contains(message, "not_found")
}
