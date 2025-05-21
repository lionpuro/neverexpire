package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/lionpuro/trackcerts/model"
	"github.com/lionpuro/trackcerts/user"
	"golang.org/x/oauth2"
)

type Handler struct {
	authService *Service
	userService *user.Service
}

func NewHandler(as *Service, us *user.Service) (*Handler, error) {
	h := &Handler{
		authService: as,
		userService: us,
	}
	return h, nil
}

func (h *Handler) Login(a *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := generateRandomState()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sess, err := h.authService.sessions.GetSession(r)
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

		url := a.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (h *Handler) Callback(a *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := h.authService.sessions.GetSession(r)
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
		tkn, err := a.config.Exchange(r.Context(), code)
		if err != nil {
			log.Printf("auth callback: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		idToken, err := a.verifyToken(r.Context(), tkn)
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

		if err := h.userService.Create(user.ID, user.Email); err != nil {
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

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := h.authService.sessions.GetSession(r)
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

func generateRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	return state, nil
}

func (h *Handler) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, err := h.authService.sessions.GetUser(r)
		if err == nil {
			ctx = user.SaveToContext(r.Context(), u)
		}
		next(w, r.WithContext(ctx))
	}
}
