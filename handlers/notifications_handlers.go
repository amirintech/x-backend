package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/services/notifications"
)

type NotificationsHandlers struct {
	notificationsService notifications.Notifications
}

func NewNotificationsHandlers(notificationsService notifications.Notifications) *NotificationsHandlers {
	return &NotificationsHandlers{
		notificationsService: notificationsService,
	}
}

func (h *NotificationsHandlers) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	// extract userID
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// subscribe to notifications
	notificationsChan := h.notificationsService.Subscribe(models.NotificationTypeFollow, userID)
	defer h.notificationsService.Unsubscribe(models.NotificationTypeFollow, userID)

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
