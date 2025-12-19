package inbound

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/notification/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type HTTPEndpoint struct {
	uc uc
}

func (h *HTTPEndpoint) ListNotifications(ctx context.Context, r *http.Request) (any, error) {
	limit := int32(20)
	offset := int32(0)

	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, goerror.NewBusiness("invalid limit", goerror.CodeInvalidInput)
		}
		limit = int32(n)
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, goerror.NewBusiness("invalid offset", goerror.CodeInvalidInput)
		}
		offset = int32(n)
	}

	resp, err := h.uc.ListNotifications(ctx, usecase.ListNotificationsInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]NotificationItemResponse, 0, len(resp.Notifications))
	for _, n := range resp.Notifications {
		items = append(items, NotificationItemResponse{
			ID:                  n.ID,
			TriggerKey:          n.TriggerKey,
			Data:                n.Data,
			Metadata:            n.Metadata,
			ReadAt:              n.ReadAt,
			CreatedAt:           n.CreatedAt,
			CategoryName:        n.CategoryName,
			CategoryDescription: n.CategoryDescription,
		})
	}

	return ListNotificationsResponse{
		Notifications: items,
		Limit:         resp.Limit,
		Offset:        resp.Offset,
	}, nil
}

func (h *HTTPEndpoint) CountUnread(ctx context.Context, r *http.Request) (any, error) {
	resp, err := h.uc.CountUnread(ctx, usecase.CountUnreadInput{})
	if err != nil {
		return nil, err
	}

	return CountUnreadResponse{Count: resp.Count}, nil
}

func (h *HTTPEndpoint) MarkRead(ctx context.Context, r *http.Request) (any, error) {
	notificationID, err := parseNotificationID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.uc.MarkRead(ctx, usecase.MarkReadInput{NotificationID: notificationID}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) MarkAllRead(ctx context.Context, r *http.Request) (any, error) {
	if err := h.uc.MarkAllRead(ctx, usecase.MarkAllReadInput{}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) DeleteNotification(ctx context.Context, r *http.Request) (any, error) {
	notificationID, err := parseNotificationID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.uc.DeleteNotification(ctx, usecase.DeleteNotificationInput{NotificationID: notificationID}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) RegisterDevice(ctx context.Context, r *http.Request) (any, error) {
	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.RegisterDevice(ctx, usecase.RegisterDeviceInput{
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *HTTPEndpoint) RemoveDevice(ctx context.Context, r *http.Request) (any, error) {
	var req RemoveDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	if err := h.uc.RemoveDevice(ctx, usecase.RemoveDeviceInput{DeviceToken: req.DeviceToken}); err != nil {
		return nil, err
	}

	return nil, nil
}

func parseNotificationID(ctx context.Context) (int64, error) {
	raw := router.GetParam(ctx, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, goerror.NewBusiness("invalid notification id", goerror.CodeInvalidInput)
	}
	return id, nil
}
