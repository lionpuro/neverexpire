package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/trackcert/db"
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

	r.HandleFunc("GET /", s.sessionMiddleware(s.handleHomePage))
	r.HandleFunc("GET /domains", s.sessionMiddleware(s.handleDomains))
	r.HandleFunc("GET /domains/new", s.sessionMiddleware(s.handleNewDomainPage))
	r.HandleFunc("POST /domains", s.sessionMiddleware(s.handleCreateDomain))
	r.HandleFunc("GET /domains/{id}", s.sessionMiddleware(s.handleDomain))
	r.HandleFunc("DELETE /domains/{id}", s.sessionMiddleware(s.handleDeleteDomain))

	r.HandleFunc("GET /account", s.sessionMiddleware(s.handleAccountPage))
	r.HandleFunc("GET /login", s.sessionMiddleware(s.handleLoginPage))
	r.HandleFunc("GET /logout", s.handleLogout)
	r.HandleFunc("GET /auth/google/login", s.handleAuth(s.Auth.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", s.handleAuthCallback(s.Auth.GoogleClient))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	s.httpServer = &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	return s, nil
}
