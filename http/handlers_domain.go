package http

import (
	"fmt"
	"log"
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
		handleErrorPage(w, "Bad request", http.StatusBadRequest)
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
			log.Printf("get domain: %v", err)
		}
		handleErrorPage(w, errMsg, errCode)
		return
	}
	if err := views.Domain(w, &u, domain, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) DomainsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	domains, err := h.DomainService.AllByUser(r.Context(), u.ID)
	if err != nil {
		log.Printf("get domains: %v", err)
		handleErrorPage(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if err := views.Domains(w, &u, domains, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) NewDomainPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := views.NewDomain(w, &u, "", nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (h *Handler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		handleErrorPage(w, "Bad request", http.StatusBadRequest)
		return
	}
	u, _ := user.FromContext(r.Context())
	if err := h.DomainService.Delete(u.ID, id); err != nil {
		log.Printf("delete domain: %v", err)
		if isHXrequest(r) {
			handleErrorPage(w, "Error deleting domain", http.StatusInternalServerError)
			return
		}
		htmxError(w, fmt.Errorf("Error deleting domain"))
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
		htmxError(w, fmt.Errorf("Please enter at least one valid domain"))
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
			htmxError(w, err)
			return
		}
		if err := views.NewDomain(w, &u, "", err); err != nil {
			log.Printf("render template: %v", err)
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
			log.Printf("create domain: %v", err)
		}
		htmxError(w, e)
		return
	}

	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}
