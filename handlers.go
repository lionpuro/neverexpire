package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/trackcerts/certs"
	"github.com/lionpuro/trackcerts/model"
	"github.com/lionpuro/trackcerts/user"
	"github.com/lionpuro/trackcerts/views"
)

func (s *Server) handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		handleErrorPage(w, r, "Page not found", http.StatusNotFound)
		return
	}
	var usr *model.User
	if u, ok := user.FromContext(r.Context()); ok {
		usr = &u
	}
	if err := views.Home(w, usr, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleAccountPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := views.Account(w, &u); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleDomain(partial bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.PathValue("id")
		id, err := strconv.Atoi(p)
		if err != nil {
			handleErrorPage(w, r, "Bad request", http.StatusBadRequest)
			return
		}
		u, _ := user.FromContext(r.Context())

		domain, err := s.Domains.ByID(r.Context(), id, u.ID)
		if err != nil {
			errCode := http.StatusInternalServerError
			errMsg := "Error retrieving domain data"
			if err == pgx.ErrNoRows {
				errCode = http.StatusNotFound
				errMsg = "Domain not found"
			}
			log.Printf("get domain: %v", err)
			handleErrorPage(w, r, errMsg, errCode)
			return
		}

		refreshData := domain.Certificate.CheckedAt.Before(time.Now().UTC().Add(-time.Minute))
		if partial && refreshData {
			info, err := certs.FetchCert(r.Context(), domain.DomainName)
			if err != nil {
				log.Printf("get domain: %v", err)
				if isHXrequest(r) {
					hxError(w, fmt.Errorf("Error fetching certificate"))
					return
				}
				handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
				return
			}
			domain.Certificate = *info
			d, err := s.Domains.Update(domain)
			if err != nil {
				log.Printf("update domain: %v", err)
				handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
				return
			}
			domain = d
			if err := views.DomainPartial(w, domain); err != nil {
				log.Printf("render template: %v", err)
			}
			return
		}
		if err := views.Domain(w, &u, domain, nil, refreshData); err != nil {
			log.Printf("render template: %v", err)
		}
	}
}

func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	domains, err := s.Domains.All(r.Context(), u.ID)
	if err != nil {
		log.Printf("get domains: %v", err)
		handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if err := views.Domains(w, &u, domains, nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		handleErrorPage(w, r, "Bad request", http.StatusBadRequest)
		return
	}
	u, _ := user.FromContext(r.Context())
	if err := s.Domains.Delete(u.ID, id); err != nil {
		log.Printf("delete domain: %v", err)
		if !isHXrequest(r) {
			handleErrorPage(w, r, "Error deleting domain", http.StatusInternalServerError)
			return
		}
		hxError(w, fmt.Errorf("Error deleting domain"))
		return
	}
	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}

func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	val := r.FormValue("domain")
	input := strings.ReplaceAll(strings.TrimSpace(val), "https://", "")
	if len(input) == 0 {
		e := fmt.Errorf("Please enter a valid domain name")
		if isHXrequest(r) {
			hxError(w, e)
			return
		}
		if err := views.NewDomain(w, &u, "", e); err != nil {
			log.Printf("render template: %v", err)
		}
		return
	}

	if err := s.Domains.Create(u, input); err != nil {
		e := fmt.Errorf("Error adding domain")
		str := `duplicate key value violates unique constraint "unique_domain_per_user"`
		if strings.Contains(err.Error(), str) {
			e = fmt.Errorf("Already tracking %s", input)
		} else {
			log.Printf("create domain: %v", err)
		}
		hxError(w, e)
		return
	}

	if isHXrequest(r) {
		w.Header().Set("HX-Location", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}

func (s *Server) handleNewDomainPage(w http.ResponseWriter, r *http.Request) {
	u, _ := user.FromContext(r.Context())
	if err := views.NewDomain(w, &u, "", nil); err != nil {
		log.Printf("render template: %v", err)
	}
}

func handleErrorPage(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if err := views.Error(w, code, msg); err != nil {
		log.Printf("render template: %v", err)
	}
}

func hxError(w http.ResponseWriter, err error) {
	w.Header().Set("HX-Retarget", "#error-container")
	if err := views.ErrorBanner(w, err); err != nil {
		log.Printf("render error: %v", err)
	}
}

func isHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
