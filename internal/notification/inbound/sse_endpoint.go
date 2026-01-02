package inbound

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

// StreamNotifications streams notification updates to the client using SSE.
// @Summary Stream notifications
// @Description Streams notification updates using Server-Sent Events (SSE).
// @Tags Notification
// @Security BearerAuth
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "streaming unsupported"
// @Router /api/v1/notification/stream [get]
func (h *HTTPEndpoint) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ctx := r.Context()

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, ": connected\n\n"); err != nil {
		slog.ErrorContext(ctx, "failed to send response connected", "error", err)
		return
	}
	flusher.Flush()

	claims := jwt.GetAuth(ctx)
	if claims == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	stream := h.uc.StreamNotifications(ctx, claims.UserID)

	// heartbeat ping, so proxies won’t drop idle connections.
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		// heartbeat ping, so proxies won’t drop idle connections.
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()

		case evt, ok := <-stream:
			if !ok {
				return
			}
			payload, err := json.Marshal(evt)
			if err != nil {
				slog.ErrorContext(ctx, "failed to marshal data", "error", err)
				continue
			}
			if _, err := fmt.Fprintf(w, "event: notification\ndata: %s\n\n", payload); err != nil {
				slog.ErrorContext(ctx, "failed to send response data", "error", err)
				return
			}
			flusher.Flush()
		}
	}
}
