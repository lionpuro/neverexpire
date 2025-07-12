package api

import (
	"net/http"

	"github.com/lionpuro/neverexpire/api/handlers"
)

func NewRouter(h *handlers.Handler) *http.ServeMux {
	r := http.NewServeMux()
	handle := func(p string, hf http.HandlerFunc) {
		r.Handle(p, h.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				hf(w, r)
			}),
		))
	}

	handle("GET /hosts", h.ListHosts)
	handle("GET /hosts/{hostname}", h.FindHost)

	return r
}
