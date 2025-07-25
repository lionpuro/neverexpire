package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lionpuro/neverexpire/notifications"
	"github.com/lionpuro/neverexpire/users"
	"github.com/lionpuro/neverexpire/web/views"
)

func (h *Handler) NotificationsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	notifs, err := h.notificationService.AllByUser(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to retrieve notifications", "error", err.Error())
		h.htmxError(w, fmt.Errorf("failed to load notifications"))
		return
	}
	h.render(views.Notifications(w, views.LayoutData{User: &u}, notifs))
}

func (h *Handler) NotificationsCount(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	notifs, err := h.notificationService.AllByUser(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to retrieve notifications", "error", err.Error())
		h.htmxError(w, fmt.Errorf("failed to load notifications"))
		return
	}
	var unread int
	for _, n := range notifs {
		if n.ReadAt == nil {
			unread++
		}
	}
	data := map[string]any{"Count": fmt.Sprintf("%d", unread)}
	h.render(views.Component(w, "notification-badge", data))
}

func (h *Handler) ReadNotifications(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.log.Error("failed to parse form", "error", err.Error())
		h.htmxError(w, fmt.Errorf("failed to update notifications"))
	}
	ids := r.Form["notification_id"]
	now := time.Now().UTC()
	var input []notifications.NotificationUpdate
	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			h.log.Error("failed to convert id to int", "error", err.Error())
			h.htmxError(w, fmt.Errorf("failed to update notifications"))
			return
		}
		input = append(input, notifications.NotificationUpdate{ID: id, ReadAt: &now})
	}
	err := h.notificationService.Update(u.ID, input)
	if err != nil {
		h.log.Error("failed to update notifications", "error", err.Error())
		h.htmxError(w, fmt.Errorf("failed to update notifications"))
		return
	}
	w.Header().Set("HX-Redirect", "/notifications")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	settings, err := h.userService.Settings(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to retrieve settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	if settings == (users.Settings{}) {
		sec := notifications.Threshold2Weeks
		sett, err := h.userService.SaveSettings(u.ID, users.SettingsInput{
			ReminderThreshold: &sec,
		})
		if err != nil {
			h.log.Error("failed to save settings", "error", err.Error())
			h.htmxError(w, fmt.Errorf("something went wrong"))
			return
		}
		settings = sett
	}
	h.render(views.Settings(w, views.LayoutData{User: &u}, settings))
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	if err := h.userService.Delete(u.ID); err != nil {
		h.htmxError(w, fmt.Errorf("error deleting account"))
		return
	}
	sess, err := h.Authenticator.Session(r)
	if err != nil {
		h.htmxError(w, fmt.Errorf("error logging out"))
		return
	}
	if err := sess.Delete(w, r); err != nil {
		h.htmxError(w, fmt.Errorf("error logging out"))
		return
	}
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateReminders(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	seconds, err := strconv.Atoi(r.FormValue("reminder_threshold"))
	if err != nil {
		h.htmxError(w, fmt.Errorf("bad request"))
		return
	}
	if _, err := h.userService.SaveSettings(u.ID, users.SettingsInput{ReminderThreshold: &seconds}); err != nil {
		h.log.Error("failed to update settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	w.Header().Set("HX-Retarget", "#banner-container")
	h.render(views.SuccessBanner(w, "Settings saved"))
}

func (h *Handler) AddWebhook(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	provider, url, err := parseWebhook(r.FormValue("webhook_provider"), r.FormValue("webhook_url"))
	if err != nil {
		h.htmxError(w, fmt.Errorf("invalid webhook"))
		return
	}
	input := users.SettingsInput{WebhookProvider: provider, WebhookURL: &url}
	if _, err := h.userService.SaveSettings(u.ID, input); err != nil {
		h.log.Error("failed to save settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	if err := notifications.SendTestNotification(url); err != nil {
		h.log.Error("failed to test notification webhook", "error", err.Error())
		h.htmxError(w, fmt.Errorf("error sending test notification"))
		return
	}
	w.Header().Set("HX-Location", "/settings")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	var s string
	if _, err := h.userService.SaveSettings(u.ID, users.SettingsInput{WebhookURL: &s}); err != nil {
		h.log.Error("failed to save settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	w.Header().Set("HX-Location", "/settings")
	w.WriteHeader(http.StatusNoContent)
}
