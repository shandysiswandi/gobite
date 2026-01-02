package inbound

import (
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	r.GET("/api/v1/notification/categories", end.ListCategories)
	r.GET("/api/v1/notification/settings", end.ListSettings)
	r.PUT("/api/v1/notification/settings", end.UpdateSettings)
	r.GET("/api/v1/notifications", end.ListNotifications)
	r.PATCH("/api/v1/notifications/:id/read", end.MarkNotificationRead)
	r.POST("/api/v1/notifications/read-all", end.MarkAllNotificationsRead)
	r.DELETE("/api/v1/notifications/:id", end.DeleteNotification)

	r.POST("/api/v1/notification/device", end.DeviceRegister)
	r.DELETE("/api/v1/notification/device", end.DeviceRemove)
}
