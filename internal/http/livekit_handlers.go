package http

import (
	"encoding/json"
	nethttp "net/http"

	pingoolivekit "pingoo_calls/internal/livekit"
)

type LiveKitHandlers struct {
	tokenService *pingoolivekit.TokenService
	roomService  *pingoolivekit.RoomService
}

func NewLiveKitHandlers(
	tokenService *pingoolivekit.TokenService,
	roomService *pingoolivekit.RoomService,
) *LiveKitHandlers {
	return &LiveKitHandlers{
		tokenService: tokenService,
		roomService:  roomService,
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
