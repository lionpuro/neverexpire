package http

import (
	"net/http"

	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/model"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := h.AuthService.Session(r)
	if err != nil {
		h.ErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if err := sess.Delete(w, r); err != nil {
		h.ErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Login(a *auth.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := auth.GenerateRandomState()
		if err != nil {
			h.log.Error("failed to generate random state token", "error", err.Error())
			h.ErrorPage(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}

		sess, err := h.AuthService.Session(r)
		if err != nil {
			h.log.Error("failed to retrieve session", "error", err.Error())
			h.ErrorPage(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}

		sess.SetState(state)
		if err := sess.Save(w, r); err != nil {
			h.ErrorPage(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}

		url := a.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (h *Handler) AuthCallback(a *auth.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := h.AuthService.Session(r)
		if err != nil {
			h.log.Error("failed to retrieve session", "error", err.Error())
			h.ErrorPage(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}

		state, ok := sess.State()
		if !ok || r.FormValue("state") != state {
			h.ErrorPage(w, r, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		tkn, err := a.ExchangeToken(r.Context(), code)
		if err != nil {
			h.log.Error("failed to exchange token", "error", err.Error())
			h.ErrorPage(w, r, "Bad request", http.StatusBadRequest)
			return
		}

		idToken, err := a.VerifyToken(r.Context(), tkn)
		if err != nil {
			h.log.Error("failed to verify token", "error", err.Error())
			h.ErrorPage(w, r, "Bad request", http.StatusBadRequest)
			return
		}

		var user struct {
			ID    string `json:"sub"`
			Email string `json:"email"`
		}
		if err := idToken.Claims(&user); err != nil {
			h.log.Error("failed to unmarshal token claims", "error", err.Error())
			h.ErrorPage(w, r, "Bad request", http.StatusBadRequest)
			return
		}

		if err := h.UserService.Create(user.ID, user.Email); err != nil {
			h.log.Error("failed to create user", "error", err.Error())
			h.ErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
			return
		}
		sess.SetUser(model.User{ID: user.ID, Email: user.Email})
		if err := sess.Save(w, r); err != nil {
			h.log.Error("failed to save session", "error", err.Error())
			h.ErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/domains", http.StatusTemporaryRedirect)
	}
}
