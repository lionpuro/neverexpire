package http

import (
	"log"
	"net/http"

	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/model"
	"github.com/lionpuro/neverexpire/views"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := h.AuthService.Session(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := sess.Delete(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Login(a *auth.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := auth.GenerateRandomState()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sess, err := h.AuthService.Session(r)
		if err != nil {
			log.Printf("get session: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		sess.SetState(state)
		if err := sess.Save(w, r); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			log.Printf("get session: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		state, ok := sess.State()
		if !ok || r.FormValue("state") != state {
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
		sess.SetUser(model.User{ID: user.ID, Email: user.Email})
		if err := sess.Save(w, r); err != nil {
			log.Printf("save session: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/domains", http.StatusTemporaryRedirect)
	}
}
