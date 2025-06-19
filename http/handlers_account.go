package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/notification"
	"github.com/lionpuro/neverexpire/user"
	"github.com/lionpuro/neverexpire/views"
)

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	settings, err := h.UserService.Settings(r.Context(), u.ID)
	if err != nil {
		log.Printf("get user settings: %v", err)
		htmxError(w, fmt.Errorf("Something went wrong"))
		return
	}
	if settings == (model.Settings{}) {
		sec := 14 * 24 * 60 * 60
		sett, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{
			RemindBefore: &sec,
		})
		if err != nil {
			log.Printf("save user settings: %v", err)
			htmxError(w, fmt.Errorf("Something went wrong"))
			return
		}
		settings = sett
	}
	if err := views.Settings(w, &u, settings); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := h.UserService.Delete(u.ID); err != nil {
		htmxError(w, fmt.Errorf("Error deleting account"))
		return
	}
	sess, err := h.AuthService.Session(r)
	if err != nil {
		htmxError(w, fmt.Errorf("Error logging out"))
		return
	}
	if err := sess.Delete(w, r); err != nil {
		htmxError(w, fmt.Errorf("Error logging out"))
		return
	}
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateReminders(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	seconds, err := strconv.Atoi(r.FormValue("remind_before"))
	if err != nil {
		htmxError(w, fmt.Errorf("Bad request"))
		return
	}
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{RemindBefore: &seconds}); err != nil {
		log.Printf("save user settings: %v", err)
		htmxError(w, fmt.Errorf("Something went wrong"))
		return
	}
	w.Header().Set("HX-Retarget", "#banner-container")
	if err := views.SuccessBanner(w, "Settings saved"); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) AddWebhook(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	url, err := parseWebhookURL(r.FormValue("webhook_url"))
	if err != nil {
		htmxError(w, fmt.Errorf("Invalid URL"))
		return
	}
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{WebhookURL: &url}); err != nil {
		log.Printf("save user settings: %v", err)
		htmxError(w, fmt.Errorf("Something went wrong"))
		return
	}
	if err := notification.SendTestNotification(url); err != nil {
		log.Printf("send message: %v", err)
		htmxError(w, fmt.Errorf("Error sending test notification"))
		return
	}
	w.Header().Set("HX-Location", "/settings")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	var s string
	if _, err := h.UserService.SaveSettings(u.ID, model.SettingsInput{WebhookURL: &s}); err != nil {
		log.Printf("save user settings: %v", err)
		htmxError(w, fmt.Errorf("Something went wrong"))
		return
	}
	w.Header().Set("HX-Location", "/settings")
	w.WriteHeader(http.StatusNoContent)
}
