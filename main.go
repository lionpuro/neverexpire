package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/trackcerts/db"
)

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening on %s...\n", srv.httpServer.Addr)
	log.Fatal(srv.httpServer.ListenAndServe())
}

type Server struct {
	DB         *db.Service
	Sessions   *SessionStore
	Auth       *AuthService
	httpServer *http.Server
}

func newServer() (*Server, error) {
	conn := fmt.Sprintf(
		"postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	dbService, err := db.NewService(conn)
	if err != nil {
		return nil, err
	}

	sessions, err := newSessionStore()
	if err != nil {
		return nil, err
	}

	auth, err := newAuthService()
	if err != nil {
		return nil, err
	}

	s := &Server{
		DB:       dbService,
		Sessions: sessions,
		Auth:     auth,
	}

	r := http.NewServeMux()

	register := func(p string, h http.HandlerFunc) {
		r.HandleFunc(p, s.sessionMiddleware(h))
	}

	register("GET /", s.handleHomePage)
	register("GET /domains", s.requireAuth(s.handleDomains))
	register("GET /domains/new", s.requireAuth(s.handleNewDomainPage))
	register("POST /domains", s.requireAuth(s.handleCreateDomain))
	register("GET /domains/{id}", s.requireAuth(s.handleDomain(false)))
	register("GET /partials/domains/{id}", s.requireAuth(s.handleDomain(true)))
	register("DELETE /domains/{id}", s.requireAuth(s.handleDeleteDomain))
	register("GET /account", s.requireAuth(s.handleAccountPage))
	register("GET /login", s.handleLoginPage)
	register("GET /logout", s.handleLogout)

	r.HandleFunc("GET /auth/google/login", s.handleAuth(s.Auth.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", s.handleAuthCallback(s.Auth.GoogleClient))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	s.httpServer = &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	return s, nil
}
