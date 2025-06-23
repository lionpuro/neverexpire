package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/notification"
	"github.com/lionpuro/neverexpire/user"
)

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	settings, err := h.UserService.Settings(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to render template", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	if settings == (model.Settings{}) {
		sec := 14 * 24 * 60 * 60
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
	if err := views.Settings(w, &u, settings); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
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
	u, _ := user.FromContext(r.Context())
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
	if err := views.SuccessBanner(w, "Settings saved"); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) AddWebhook(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
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
	u, _ := user.FromContext(r.Context())
	var s string
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{WebhookURL: &s}); err != nil {
		h.log.Error("failed to save settings", "error", err.Error())
		h.htmxError(w, fmt.Errorf("something went wrong"))
		return
	}
	w.Header().Set("HX-Location", "/settings")
	w.WriteHeader(http.StatusNoContent)
}
