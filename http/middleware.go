package http

import (
	"net/http"
)

// Retrieve session from store and save user data to the request context.
// Use RequireAuth afterwards to actually stop any unauthenticated requests.
func (h *Handler) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sess, err := h.AuthService.Session(r)
		if err == nil {
			if u := sess.User(); u != nil {
				ctx = userToContext(r.Context(), *u)
			}
		}
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := userFromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func contentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next(w, r)
	}
}
