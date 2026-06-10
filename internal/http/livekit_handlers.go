package http

import (
	"encoding/json"
	"log"
	nethttp "net/http"

	pingoolivekit "pingoo_calls/internal/livekit"
)

type LiveKitHandlers struct {
	tokenService   *pingoolivekit.TokenService
	roomService    *pingoolivekit.RoomService
	webhookService *pingoolivekit.WebhookService
}

func NewLiveKitHandlers(
	tokenService *pingoolivekit.TokenService,
	roomService *pingoolivekit.RoomService,
	webhookService *pingoolivekit.WebhookService,
) *LiveKitHandlers {
	return &LiveKitHandlers{
		tokenService:   tokenService,
		roomService:    roomService,
		webhookService: webhookService,
	}
}

func (h *LiveKitHandlers) Token(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req pingoolivekit.TokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, nethttp.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.tokenService.Generate(req)
	if err != nil {
		WriteError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, nethttp.StatusOK, resp)
}

func (h *LiveKitHandlers) EnsureRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req pingoolivekit.EnsureRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, nethttp.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.roomService.Ensure(r.Context(), req)
	if err != nil {
		WriteError(w, nethttp.StatusBadGateway, err.Error())
		return
	}

	WriteJSON(w, nethttp.StatusOK, resp)
}

func (h *LiveKitHandlers) EndRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req pingoolivekit.EndRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, nethttp.StatusBadRequest, "invalid json body")
		return
	}

	resp, err := h.roomService.End(r.Context(), req)
	if err != nil {
		WriteError(w, nethttp.StatusBadGateway, err.Error())
		return
	}

	WriteJSON(w, nethttp.StatusOK, resp)
}

func (h *LiveKitHandlers) PrepareSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req pingoolivekit.PrepareSessionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, nethttp.StatusBadRequest, "invalid json body")
		return
	}

	ensureResp, err := h.roomService.Ensure(r.Context(), pingoolivekit.EnsureRoomRequest{
		CallID: req.CallID,
	})
	if err != nil {
		WriteError(w, nethttp.StatusBadGateway, err.Error())
		return
	}

	tokenResp, err := h.tokenService.Generate(pingoolivekit.TokenRequest{
		CallID:    req.CallID,
		UserID:    req.UserID,
		DeviceID:  req.DeviceID,
		MediaKind: req.MediaKind,
	})
	if err != nil {
		WriteError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, nethttp.StatusOK, pingoolivekit.PrepareSessionResponse{
		LiveKitURL: tokenResp.LiveKitURL,
		RoomName:   ensureResp.RoomName,
		Token:      tokenResp.Token,
	})
}

func (h *LiveKitHandlers) Webhook(w nethttp.ResponseWriter, r *nethttp.Request) {
	event, err := h.webhookService.Receive(r)
	if err != nil {
		WriteError(w, nethttp.StatusUnauthorized, err.Error())
		return
	}

	log.Printf(
		"livekit webhook received: event=%s room=%s participant=%s",
		event.Event,
		event.RoomName,
		event.Participant,
	)

	WriteJSON(w, nethttp.StatusOK, map[string]string{
		"status": "received",
	})
}
