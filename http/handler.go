package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/trackcerts/auth"
	"github.com/lionpuro/trackcerts/certs"
	"github.com/lionpuro/trackcerts/domain"
	"github.com/lionpuro/trackcerts/model"
	"github.com/lionpuro/trackcerts/notification"
	"github.com/lionpuro/trackcerts/user"
	"github.com/lionpuro/trackcerts/views"
)

type Handler struct {
	UserService   *user.Service
	DomainService *domain.Service
	AuthService   *auth.Service
}

func NewHandler(us *user.Service, ds *domain.Service, as *auth.Service) *Handler {
	return &Handler{
		UserService:   us,
		DomainService: ds,
		AuthService:   as,
	}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		if err := views.Error(w, http.StatusNotFound, "Page not found"); err != nil {
			log.Printf("render template: %v", err)
		}
		return
	}
	var usr *model.User
	if u, ok := user.FromContext(r.Context()); ok {
		usr = &u
	}
	if err := views.Home(w, usr, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

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

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf("render template: %v", err)
	}
}

// Middleware

func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := user.FromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (h *Handler) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, err := h.AuthService.Sessions.GetUser(r)
		if err == nil {
			ctx = user.SaveToContext(r.Context(), u)
		}
		next(w, r.WithContext(ctx))
	}
}

// Auth

func (h *Handler) Login(a *auth.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := h.AuthService.GenerateRandomState()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sess, err := h.AuthService.Sessions.GetSession(r)
		if err != nil {
			log.Printf("get session: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		sess.Values["state"] = state
		if err := sess.Save(r, w); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		url := a.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (h *Handler) AuthCallback(a *auth.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := h.AuthService.Sessions.GetSession(r)
		if err != nil {
			log.Printf("get session: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if r.FormValue("state") != sess.Values["state"] {
			http.Error(w, "Invalid state parameter.", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		tkn, err := a.ExchangeToken(r.Context(), code)
		if err != nil {
			log.Printf("auth callback: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		idToken, err := a.VerifyToken(r.Context(), tkn)
		if err != nil {
			log.Printf("verify token: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var user struct {
			ID    string `json:"sub"`
			Email string `json:"email"`
		}
		if err := idToken.Claims(&user); err != nil {
			log.Printf("unmarshal claims: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := h.UserService.Create(user.ID, user.Email); err != nil {
			log.Printf("%v", err)
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}
		sess.Values["user"] = model.User{ID: user.ID, Email: user.Email}
		if err := sess.Save(r, w); err != nil {
			log.Printf("save session: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/domains", http.StatusTemporaryRedirect)
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := h.AuthService.Sessions.GetSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := h.UserService.Delete(u.ID); err != nil {
		htmxError(w, fmt.Errorf("Error deleting account"))
		return
	}
	sess, err := h.AuthService.Sessions.GetSession(r)
	if err != nil {
		htmxError(w, fmt.Errorf("Error logging out"))
		return
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
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

// Domain

func (h *Handler) DomainPage(partial bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.PathValue("id")
		id, err := strconv.Atoi(p)
		if err != nil {
			handleErrorPage(w, r, "Bad request", http.StatusBadRequest)
			return
		}
		u, _ := user.FromContext(r.Context())

		domain, err := h.DomainService.ByID(r.Context(), id, u.ID)
		if err != nil {
			errCode := http.StatusInternalServerError
			errMsg := "Error retrieving domain data"
			if err == pgx.ErrNoRows {
				errCode = http.StatusNotFound
				errMsg = "Domain not found"
			}
			log.Printf("get domain: %v", err)
			handleErrorPage(w, r, errMsg, errCode)
			return
		}

		refreshData := domain.Certificate.CheckedAt.Before(time.Now().UTC().Add(-time.Minute))
		if partial && refreshData {
			info, err := certs.FetchCert(r.Context(), domain.DomainName)
			if err != nil {
				log.Printf("get domain: %v", err)
				if isHXrequest(r) {
					htmxError(w, fmt.Errorf("Error fetching certificate"))
					return
				}
				handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
				return
			}
			domain.Certificate = *info
			d, err := h.DomainService.Update(domain)
			if err != nil {
				log.Printf("update domain: %v", err)
				handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
				return
			}
			domain = d
			if err := views.DomainPartial(w, domain); err != nil {
				log.Printf("render template: %v", err)
			}
			return
		}
		if err := views.Domain(w, &u, domain, nil, refreshData); err != nil {
			log.Printf("render template: %v", err)
		}
	}
}

func (h *Handler) DomainsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	domains, err := h.DomainService.All(r.Context(), u.ID)
	if err != nil {
		log.Printf("get domains: %v", err)
		handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if err := views.Domains(w, &u, domains, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) NewDomainPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := views.NewDomain(w, &u, "", nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		handleErrorPage(w, r, "Bad request", http.StatusBadRequest)
		return
	}
	u, _ := user.FromContext(r.Context())
	if err := h.DomainService.Delete(u.ID, id); err != nil {
		log.Printf("delete domain: %v", err)
		if isHXrequest(r) {
			handleErrorPage(w, r, "Error deleting domain", http.StatusInternalServerError)
			return
		}
		htmxError(w, fmt.Errorf("Error deleting domain"))
		return
	}
	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}

func (h *Handler) CreateDomains(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	input := strings.TrimSpace(r.FormValue("domains"))
	ds := strings.Split(input, ",")
	if len(input) < 3 {
		htmxError(w, fmt.Errorf("Please enter at least one valid domain"))
		return
	}
	var names []string
	var errs []error
	for _, d := range ds {
		name, err := parseDomain(d)
		if err != nil {
			errs = append(errs, err)
		}
		if name != "" {
			names = append(names, name)
		}
	}
	if len(errs) > 0 {
		fmt.Printf("%v", errs)
		err := fmt.Errorf("Invalid domain name")
		if isHXrequest(r) {
			htmxError(w, err)
			return
		}
		if err := views.NewDomain(w, &u, "", err); err != nil {
			log.Printf("render template: %v", err)
		}
		return
	}

	if err := h.DomainService.CreateMultiple(u, names); err != nil {
		e := fmt.Errorf("Error adding domain")
		switch {
		case strings.Contains(err.Error(), "already tracking"):
			e = err
		case strings.Contains(err.Error(), "can't connect to"):
			e = err
		default:
			log.Printf("create domain: %v", err)
		}
		htmxError(w, e)
		return
	}

	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}

// Helper

func isHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func htmxError(w http.ResponseWriter, err error) {
	w.Header().Set("HX-Retarget", "#banner-container")
	if err := views.ErrorBanner(w, err); err != nil {
		log.Printf("render error: %v", err)
	}
}

func handleErrorPage(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if err := views.Error(w, code, msg); err != nil {
		log.Printf("render template: %v", err)
	}
}
