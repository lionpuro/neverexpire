package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/http/views"
	"github.com/lionpuro/neverexpire/user"
)

func (h *Handler) DomainPage(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		h.ErrorPage(w, "Bad request", http.StatusBadRequest)
		return
	}

	u, _ := user.FromContext(r.Context())
	domain, err := h.DomainService.ByID(r.Context(), id, u.ID)
	if err != nil {
		errCode := http.StatusNotFound
		errMsg := "Domain not found"
		if !db.IsErrNoRows(err) {
			errCode = http.StatusInternalServerError
			errMsg = "Error retrieving domain data"
			h.log.Error("failed to retrieve domain data", "error", err.Error())
		}
		h.ErrorPage(w, errMsg, errCode)
		return
	}
	if err := views.Domain(w, &u, domain, nil); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) DomainsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	domains, err := h.DomainService.AllByUser(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to get domains", "error", err.Error())
		h.ErrorPage(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if err := views.Domains(w, &u, domains, nil); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) NewDomainPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := views.NewDomain(w, &u, "", nil); err != nil {
		h.log.Error("failed to render template", "error", err.Error())
	}
}

func (h *Handler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		h.ErrorPage(w, "Bad request", http.StatusBadRequest)
		return
	}
	u, _ := user.FromContext(r.Context())
	if err := h.DomainService.Delete(u.ID, id); err != nil {
		h.log.Error("failed to delete domain", "error", err.Error())
		if isHXrequest(r) {
			h.ErrorPage(w, "Error deleting domain", http.StatusInternalServerError)
			return
		}
		h.htmxError(w, fmt.Errorf("Error deleting domain"))
		return
	}
	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}

func (h *Handler) CreateDomains(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	input := strings.TrimSpace(r.FormValue("domains"))
	ds := strings.Split(input, ",")
	if len(input) < 3 {
		h.htmxError(w, fmt.Errorf("Please enter at least one valid domain"))
		return
	}
	var names []string
	var errs []error
	for _, d := range ds {
		name, err := parseDomain(d)
		if err != nil {
			errs = append(errs, err)
		}
		if name != "" {
			names = append(names, name)
		}
	}
	if len(errs) > 0 {
		err := fmt.Errorf("Invalid domain name")
		if isHXrequest(r) {
			h.htmxError(w, err)
			return
		}
		if err := views.NewDomain(w, &u, "", err); err != nil {
			h.log.Error("failed to render template", "error", err.Error())
		}
		return
	}

	if err := h.DomainService.CreateMultiple(u, names); err != nil {
		e := fmt.Errorf("Error adding domain")
		switch {
		case
			strings.Contains(err.Error(), "already tracking"),
			strings.Contains(err.Error(), "can't connect to"):
			e = err
		default:
			h.log.Error("failed to create domain", "error", err.Error())
		}
		h.htmxError(w, e)
		return
	}

	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}
