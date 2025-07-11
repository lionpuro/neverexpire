package api

import (
	"encoding/json"
	"net/http"

	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
)

type Handler struct {
	userService   *user.Service
	domainService *domain.Service
	keyService    *KeyService
	log           logging.Logger
}

func NewHandler(logger logging.Logger, us *user.Service, ds *domain.Service, ks *KeyService) *Handler {
	return &Handler{
		userService:   us,
		domainService: ds,
		keyService:    ks,
		log:           logger,
	}
}

func (h *Handler) json(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("failed to write json response: %v", err)
	}
}

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
		match := CompareKey(k, key.Hash)
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

func (h *Handler) Hosts(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	domains, err := h.domainService.AllByUser(r.Context(), uid)
	if err != nil {
		h.log.Error("failed to get domains", "error", err.Error())
		h.json(w, http.StatusInternalServerError, "internal server error")
		return
	}
	var hosts []domain.APIModel
	for _, d := range domains {
		host := domain.ToAPIModel(d)
		hosts = append(hosts, host)
	}
	data := map[string]interface{}{
		"data": hosts,
	}
	h.json(w, http.StatusOK, data)
}
