package handlers

import (
	"net/http"

	"github.com/lionpuro/neverexpire/keys"
)

func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("access_key")
		if len(k) != 128 {
			h.json(w, http.StatusUnauthorized, "invalid or missing access key")
			return
		}
		id := k[:8]
		key, err := h.keyService.ByID(r.Context(), id)
		if err != nil {
			h.json(w, http.StatusUnauthorized, "invalid or missing access key")
			return
		}
		match := keys.CompareAccessKey(k, key.Hash)
		if !match {
			h.json(w, http.StatusUnauthorized, "invalid access key")
			return
		}
		ctx := userIDToContext(r.Context(), key.UserID)
		if _, ok := userIDFromContext(ctx); !ok {
			h.log.Error("failed to retrieve user_id from request context")
			h.json(w, http.StatusInternalServerError, "internal server error")
			return
		}
		next(w, r.WithContext(ctx))
	}
}
