package inbound

import (
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	// r.GET("/notifications", end.ListNotifications)
	// r.GET("/notifications/unread-count", end.CountUnread)

	// r.PATCH("/notifications/:id/read", end.MarkRead)
	// r.PATCH("/notifications/bulk/read", end.MarkAllRead)

	// r.DELETE("/notifications/:id", end.DeleteNotification)

	r.POST("/notifications/devices/register", end.RegisterDevice)
	r.POST("/notifications/devices/remove", end.RemoveDevice)
}
