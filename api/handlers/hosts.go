package handlers

import (
	"net/http"

	"github.com/lionpuro/neverexpire/domain"
)

func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
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
