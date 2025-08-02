package web

import (
	"net/http"
	"strings"
)

// Retrieve session from store and save user data to the request context.
// Use RequireAuth afterwards to actually stop any unauthenticated requests.
func (h *Handler) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sess, err := h.Authenticator.Session(r)
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

func redirectTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		l := len(path) - 1
		if l > 0 && strings.HasSuffix(path, "/") {
			url := path[:l]
			http.Redirect(w, r, url, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
