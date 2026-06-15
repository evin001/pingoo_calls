package http

import (
	"context"
	"encoding/json"
	"log"
	nethttp "net/http"
	"time"

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
		log.Printf(
			"livekit token failed: call_id=%s user_id=%s device_id=%s media_kind=%s error=%v",
			req.CallID,
			req.UserID,
			req.DeviceID,
			req.MediaKind,
			err,
		)
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
		log.Printf("livekit ensure room failed: call_id=%s error=%v", req.CallID, err)
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
		log.Printf("livekit end room failed: call_id=%s error=%v", req.CallID, err)
		WriteError(w, nethttp.StatusBadGateway, err.Error())
		return
	}

	WriteJSON(w, nethttp.StatusOK, resp)
}

func (h *LiveKitHandlers) Webhook(w nethttp.ResponseWriter, r *nethttp.Request) {
	event, err := h.webhookService.Receive(r)
	if err != nil {
		log.Printf("livekit webhook rejected: remote_addr=%s error=%v", r.RemoteAddr, err)
		WriteError(w, nethttp.StatusUnauthorized, err.Error())
		return
	}

	log.Printf(
		"livekit webhook received: event=%s room=%s participant=%s",
		event.Event,
		event.RoomName,
		event.Participant,
	)

	if event.Ignored {
		log.Printf(
			"livekit webhook ignored: event=%s room=%s participant=%s",
			event.Event,
			event.RoomName,
			event.Participant,
		)
		WriteJSON(w, nethttp.StatusOK, map[string]string{
			"status": "ignored",
		})
		return
	}

	forwardCtx, forwardCancel := contextWithTimeout(r, 5*time.Second)
	defer forwardCancel()

	if err := h.webhookService.Forward(forwardCtx, event); err != nil {
		log.Printf(
			"failed to forward livekit webhook: event=%s room=%s participant=%s error=%v",
			event.Event,
			event.RoomName,
			event.Participant,
			err,
		)
		WriteError(w, nethttp.StatusBadGateway, err.Error())
		return
	}

	log.Printf(
		"livekit webhook handled: event=%s room=%s participant=%s",
		event.Event,
		event.RoomName,
		event.Participant,
	)

	WriteJSON(w, nethttp.StatusOK, map[string]string{
		"status": "received",
	})
}

func contextWithTimeout(r *nethttp.Request, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), timeout)
}
