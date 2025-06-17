package http

import (
	"net/http"

	"github.com/lionpuro/neverexpire/user"
)

// Retrieve session from store and save user data to the request context.
// Use RequireAuth afterwards to actually stop any unauthenticated requests.
func (h *Handler) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sess, err := h.AuthService.Session(r)
		if err == nil {
			if u := sess.User(); u != nil {
				ctx = user.SaveToContext(r.Context(), *u)
			}
		}
		next(w, r.WithContext(ctx))
	}
}

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
