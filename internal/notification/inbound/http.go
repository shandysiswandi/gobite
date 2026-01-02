package inbound

import (
	"net/http"

	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	r.POST("/api/v1/notification/device", end.DeviceRegister)
	r.DELETE("/api/v1/notification/device", end.DeviceRemove)

	r.GET("/api/v1/notification/categories", end.ListCategories)
	r.GET("/api/v1/notification/settings", end.ListSettings)
	r.PUT("/api/v1/notification/settings", end.UpdateSettings)

	r.GET("/api/v1/notification/inbox", end.ListInbox)
	r.PATCH("/api/v1/notification/inbox/:id/read", end.MarkInboxRead)
	r.PUT("/api/v1/notification/inbox/read-all", end.MarkAllInboxRead)
	r.DELETE("/api/v1/notification/inbox/:id", end.DeleteInbox)

	r.GETRaw("/api/v1/notification/stream", http.HandlerFunc(end.StreamNotifications))
}
