package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lionpuro/trackcert/certs"
	"github.com/lionpuro/trackcert/model"
	"github.com/lionpuro/trackcert/views"
)

func (s *Server) handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		handleErrorPage(w, r, "Page not found", http.StatusNotFound)
		return
	}
	var user *model.User
	if u, ok := getUserCtx(r.Context()); ok {
		user = &u
	}
	if err := views.Home(w, user); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleAccountPage(w http.ResponseWriter, r *http.Request) {
	user, _ := getUserCtx(r.Context())
	if err := views.Account(w, &user); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := views.Login(w); err != nil {
		log.Printf(err.Error())
	}
}

func (s *Server) handleDomain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, _ := getUserCtx(r.Context())

	domain, err := s.DB.DomainByID(r.Context(), id, user.ID)
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

	if domain.CheckedAt.Before(time.Now().UTC().Add(-time.Minute)) {
		info, err := certs.FetchCert(r.Context(), domain.DomainName)
		if err != nil {
			log.Printf("get domain: %v", err)
			handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
			return
		}
		domain.CertificateInfo = *info
		if err := s.DB.UpdateDomainInfo(domain.ID, user.ID, *info); err != nil {
			log.Printf("update domain: %v", err)
			handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
			return
		}
		d, err := s.DB.DomainByID(r.Context(), id, user.ID)
		if err != nil {
			errCode := http.StatusInternalServerError
			errMsg := "Something went wrong"
			if err == pgx.ErrNoRows {
				errCode = http.StatusNotFound
				errMsg = "Domain not found"
			}
			log.Printf("get domains: %v", err)
			handleErrorPage(w, r, errMsg, errCode)
			return
		}
		domain = d
	}
	if err := views.Domain(w, &user, domain); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	user, _ := getUserCtx(r.Context())
	domains, err := s.DB.Domains(r.Context(), user.ID)
	if err != nil {
		log.Printf("get domains: %v", err)
		handleErrorPage(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if err := views.Domains(w, &user, domains); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *Server) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, _ := getUserCtx(r.Context())
	if err := s.DB.DeleteDomain(id, user.ID); err != nil {
		log.Printf("delete domain: %v", err)
		http.Error(w, "Error deleting domain", http.StatusInternalServerError)
		return
	}
	if isHXrequest(r) {
		w.Header().Set("HX-Redirect", "/domains")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/domains", http.StatusOK)
}

func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := getUserCtx(r.Context())
	val := r.FormValue("domain")
	input := strings.ReplaceAll(strings.TrimSpace(val), "https://", "")
	if len(input) == 0 {
		http.Redirect(w, r, "/domains/new", http.StatusBadRequest)
		return
	}
	cert, err := certs.FetchCert(r.Context(), input)
	if err != nil {
		log.Printf("query certificates: %v", err)
		http.Error(w, "Host not found", http.StatusBadRequest)
		return
	}
	domain := model.Domain{
		UserID:          user.ID,
		DomainName:      input,
		CertificateInfo: *cert,
	}

	if err := s.DB.CreateDomain(domain); err != nil {
		log.Printf("create domain: %v", err)
		http.Error(w, "Error adding domain", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/domains")
	http.Redirect(w, r, "/domains", http.StatusOK)
}

func (s *Server) handleNewDomainPage(w http.ResponseWriter, r *http.Request) {
	user, _ := getUserCtx(r.Context())
	if err := views.NewDomain(w, &user); err != nil {
		log.Printf("render template: %v", err)
	}
}

func handleErrorPage(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if err := views.Error(w, code, msg); err != nil {
		log.Printf("render template: %v", err)
	}
}

func isHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
