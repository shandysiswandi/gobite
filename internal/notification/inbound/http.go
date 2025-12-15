package inbound

import (
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
)

func RegisterHTTPEndpoint(r *pkgrouter.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	r.GET("/notifications", end.ListNotifications)
	r.GET("/notifications/unread-count", end.CountUnread)
	r.POST("/notifications/:id/read", end.MarkRead)
	r.POST("/notifications/read-all", end.MarkAllRead)
	r.POST("/notifications/:id/delete", end.DeleteNotification)

	r.POST("/notifications/devices/register", end.RegisterDevice)
	r.POST("/notifications/devices/remove", end.RemoveDevice)
}
