package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/notification"
)

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	settings, err := h.UserService.Settings(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to retrieve settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	if settings == (model.Settings{}) {
		sec := notification.Threshold2Weeks
		sett, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{
			RemindBefore: &sec,
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
	if err := h.UserService.Delete(u.ID); err != nil {
		h.htmxError(w, fmt.Errorf("error deleting account"))
		return
	}
	sess, err := h.AuthService.Session(r)
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
	seconds, err := strconv.Atoi(r.FormValue("remind_before"))
	if err != nil {
		h.htmxError(w, fmt.Errorf("bad request"))
		return
	}
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{RemindBefore: &seconds}); err != nil {
		h.log.Error("failed to update settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	w.Header().Set("HX-Retarget", "#banner-container")
	h.render(views.SuccessBanner(w, "Settings saved"))
}

func (h *Handler) AddWebhook(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	url, err := parseWebhookURL(r.FormValue("webhook_url"))
	if err != nil {
		h.htmxError(w, fmt.Errorf("invalid URL"))
		return
	}
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{WebhookURL: &url}); err != nil {
		h.log.Error("failed to save settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	if err := notification.SendTestNotification(url); err != nil {
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
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{WebhookURL: &s}); err != nil {
		h.log.Error("failed to save settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	w.Header().Set("HX-Location", "/settings")
	w.WriteHeader(http.StatusNoContent)
}
