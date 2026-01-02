package inbound

import (
	"strconv"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/notification/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type HTTPEndpoint struct {
	uc uc
}

// DeviceRegister registers a device for push notifications.
// @Summary Register device
// @Description Registers a device token for the authenticated user.
// @Tags Notification
// @Security BearerAuth
// @Accept json
// @Param request body RegisterDeviceRequest true "Device registration payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/device [post]
func (h *HTTPEndpoint) DeviceRegister(r *router.Request) (any, error) {
	var req RegisterDeviceRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.DeviceRegister(r.Context(), usecase.DeviceRegisterInput{
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
	})
}

// DeviceRemove removes a device from push notifications.
// @Summary Remove device
// @Description Removes a device token for the authenticated user.
// @Tags Notification
// @Security BearerAuth
// @Accept json
// @Param request body RemoveDeviceRequest true "Device removal payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/device [delete]
func (h *HTTPEndpoint) DeviceRemove(r *router.Request) (any, error) {
	var req RemoveDeviceRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	return nil, h.uc.DeviceRemove(r.Context(), usecase.DeviceRemoveInput{DeviceToken: req.DeviceToken})
}

// ListCategories returns notification categories.
// @Summary List notification categories
// @Description Returns all notification categories.
// @Tags Notification
// @Security BearerAuth
// @Produce json
// @Success 200 {object} router.successResponse{data=NotificationCategoriesResponse} "Category list"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/categories [get]
func (h *HTTPEndpoint) ListCategories(r *router.Request) (any, error) {
	items, err := h.uc.ListCategories(r.Context())
	if err != nil {
		return nil, err
	}

	resp := make([]NotificationCategoryResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, NotificationCategoryResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			IsMandatory: item.IsMandatory,
		})
	}

	return NotificationCategoriesResponse{Categories: resp}, nil
}

// ListSettings returns user notification settings.
// @Summary List notification settings
// @Description Returns notification settings for the authenticated user.
// @Tags Notification
// @Security BearerAuth
// @Produce json
// @Success 200 {object} router.successResponse{data=NotificationSettingsResponse} "Settings list"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/settings [get]
func (h *HTTPEndpoint) ListSettings(r *router.Request) (any, error) {
	items, err := h.uc.ListSettings(r.Context())
	if err != nil {
		return nil, err
	}

	resp := make([]NotificationSettingResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, NotificationSettingResponse{
			CategoryID: item.CategoryID,
			Channel:    channelToString(item.Channel),
			IsEnabled:  item.IsEnabled,
		})
	}

	return NotificationSettingsResponse{Settings: resp}, nil
}

// UpdateSettings updates user notification settings.
// @Summary Update notification settings
// @Description Updates notification settings for the authenticated user.
// @Tags Notification
// @Security BearerAuth
// @Accept json
// @Param request body NotificationSettingsUpdateRequest true "Settings payload"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid request body"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/settings [put]
func (h *HTTPEndpoint) UpdateSettings(r *router.Request) (any, error) {
	var req NotificationSettingsUpdateRequest
	if err := r.DecodeBody(&req); err != nil {
		return nil, err
	}

	inputs := make([]usecase.UpdateSettingInput, 0, len(req.Settings))
	for _, setting := range req.Settings {
		inputs = append(inputs, usecase.UpdateSettingInput{
			CategoryID: setting.CategoryID,
			Channel:    setting.Channel,
			IsEnabled:  setting.IsEnabled,
		})
	}

	return nil, h.uc.UpdateSettings(r.Context(), usecase.UpdateSettingsInput{Settings: inputs})
}

// ListInbox returns user notifications.
// @Summary List inbox
// @Description Returns inbox notifications for the authenticated user.
// @Tags Inbox
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status (all|read|unread)"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} router.successResponse{data=NotificationsResponse} "Notification list"
// @Failure 400 {object} router.errorResponse "Invalid query parameters"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/inbox [get]
func (h *HTTPEndpoint) ListInbox(r *router.Request) (any, error) {
	query := r.URL.Query()
	limit, err := parseInt32(query.Get("limit"))
	if err != nil {
		return nil, goerror.NewInvalidFormat()
	}
	offset, err := parseInt32(query.Get("offset"))
	if err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	items, err := h.uc.ListInbox(r.Context(), usecase.ListInboxInput{
		Status: query.Get("status"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	resp := make([]NotificationResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, NotificationResponse{
			ID:         item.ID,
			CategoryID: item.CategoryID,
			TriggerKey: item.TriggerKey.String(),
			Data:       item.Data,
			Metadata:   item.Metadata,
			ReadAt:     item.ReadAt,
			CreatedAt:  item.CreatedAt,
		})
	}

	return NotificationsResponse{Notifications: resp}, nil
}

// MarkInboxRead marks a notification as read.
// @Summary Mark inbox read
// @Description Marks an inbox notification as read.
// @Tags Inbox
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid notification id"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 404 {object} router.errorResponse "Notification not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/inbox/{id}/read [patch]
func (h *HTTPEndpoint) MarkInboxRead(r *router.Request) (any, error) {
	id, err := strconv.ParseInt(r.GetParam("id"), 10, 64)
	if err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	return nil, h.uc.MarkInboxRead(r.Context(), usecase.MarkInboxReadInput{ID: id})
}

// MarkAllInboxRead marks all notifications as read.
// @Summary Mark all inbox read
// @Description Marks all inbox notifications as read for the authenticated user.
// @Tags Inbox
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/inbox/read-all [put]
func (h *HTTPEndpoint) MarkAllInboxRead(r *router.Request) (any, error) {
	return nil, h.uc.MarkAllInboxRead(r.Context())
}

// DeleteInbox removes a notification.
// @Summary Delete inbox
// @Description Soft deletes an inbox notification for the authenticated user.
// @Tags Inbox
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} router.errorResponse "Invalid notification id"
// @Failure 401 {object} router.errorResponse "Unauthorized"
// @Failure 404 {object} router.errorResponse "Notification not found"
// @Failure 422 {object} router.errorResponse "Validation error"
// @Failure 500 {object} router.errorResponse "Internal server error"
// @Router /api/v1/notification/inbox/{id} [delete]
func (h *HTTPEndpoint) DeleteInbox(r *router.Request) (any, error) {
	id, err := strconv.ParseInt(r.GetParam("id"), 10, 64)
	if err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	return nil, h.uc.DeleteInbox(r.Context(), usecase.DeleteInboxInput{ID: id})
}

func parseInt32(raw string) (int32, error) {
	if raw == "" {
		return 0, nil
	}

	val, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(val), nil
}

func channelToString(ch entity.Channel) string {
	switch ch {
	case entity.ChannelInApp:
		return "in_app"
	case entity.ChannelEmail:
		return "email"
	case entity.ChannelSMS:
		return "sms"
	case entity.ChannelPush:
		return "push"
	case entity.ChannelUnknown:
		return "in_app"
	default:
		return "unknown"
	}
}
