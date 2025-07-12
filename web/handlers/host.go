package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/web/views"
)

func (h *Handler) HostPage(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		h.ErrorPage(w, r, "Bad request", http.StatusBadRequest)
		return
	}

	u, _ := userFromContext(r.Context())
	host, err := h.hostService.ByID(r.Context(), id, u.ID)
	if err != nil {
		errCode := http.StatusNotFound
		errMsg := "Host not found"
		if !db.IsErrNoRows(err) {
			errCode = http.StatusInternalServerError
			errMsg = "Error retrieving host data"
			h.log.Error("failed to retrieve host data", "error", err.Error())
		}
		h.ErrorPage(w, r, errMsg, errCode)
		return
	}
	h.render(views.Host(w, views.LayoutData{User: &u}, host))
}

func (h *Handler) HostsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	hosts, err := h.hostService.AllByUser(r.Context(), u.ID)
	if err != nil {
		h.log.Error("failed to get hosts", "error", err.Error())
		h.ErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}
	h.render(views.Hosts(w, views.LayoutData{User: &u}, hosts))
}

func (h *Handler) NewHostsPage(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	h.render(views.NewHosts(w, views.LayoutData{User: &u}, ""))
}

func (h *Handler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		h.ErrorPage(w, r, "Bad request", http.StatusBadRequest)
		return
	}
	u, _ := userFromContext(r.Context())
	if err := h.hostService.Delete(u.ID, id); err != nil {
		h.log.Error("failed to delete host", "error", err.Error())
		if isHXrequest(r) {
			h.ErrorPage(w, r, "Error deleting host", http.StatusInternalServerError)
			return
		}
		h.htmxError(w, fmt.Errorf("error deleting host"))
		return
	}
	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/hosts")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/hosts", http.StatusOK)
}

func (h *Handler) CreateHosts(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromContext(r.Context())
	input := strings.TrimSpace(r.FormValue("hosts"))
	hs := strings.Split(input, ",")
	if len(input) < 3 {
		h.htmxError(w, fmt.Errorf("please enter at least one valid host"))
		return
	}
	var names []string
	var errs []error
	for _, h := range hs {
		name, err := hosts.ParseHostname(h)
		if err != nil {
			errs = append(errs, err)
		}
		if name != "" {
			names = append(names, name)
		}
	}
	if len(errs) > 0 {
		err := fmt.Errorf("invalid hostname")
		if isHXrequest(r) {
			h.htmxError(w, err)
			return
		}
		h.render(views.NewHosts(w, views.LayoutData{User: &u, Error: err}, ""))
		return
	}

	if err := h.hostService.Create(u, names); err != nil {
		e := fmt.Errorf("error adding host")
		switch {
		case
			strings.Contains(err.Error(), "already tracking"),
			strings.Contains(err.Error(), "can't connect to"):
			e = err
		default:
			h.log.Error("failed to create host", "error", err.Error())
		}
		h.htmxError(w, e)
		return
	}

	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/hosts")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/hosts", http.StatusOK)
}
