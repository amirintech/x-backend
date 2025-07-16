package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/services"
	"github.com/aimrintech/x-backend/stores"
)

type NotificationsHandlers struct {
	notificationsService services.Notifications
	notificationsStore   stores.NotificationsStore
}

func NewNotificationsHandlers(notificationsService services.Notifications, notificationsStore stores.NotificationsStore) *NotificationsHandlers {
	return &NotificationsHandlers{
		notificationsService: notificationsService,
		notificationsStore:   notificationsStore,
	}
}

func (h *NotificationsHandlers) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	notificationType := r.PathValue("type")
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if notificationType == "" || userID == "" {
		writeError(w, r, http.StatusBadRequest, "Type and user ID are required")
		return
	}

	// set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, r, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	// subscribe to notifications
	notificationsChan := h.notificationsService.Subscribe(models.NotificationType(notificationType), userID)
	defer h.notificationsService.Unsubscribe(models.NotificationType(notificationType), userID)

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			// client disconnected
			return
		case notification, ok := <-notificationsChan:
			if !ok {
				return
			}

			// marshal JSON
			data, err := json.Marshal(notification)
			if err != nil {
				continue
			}

			// SSE format: "data: <json>\n\n"
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))

			flusher.Flush()
		}
	}
}
