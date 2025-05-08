package main

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
)

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func newServer() (*http.Server, error) {
	sessions, err := newSessionStore()
	if err != nil {
		return nil, err
	}

	googleAuth, err := newGoogleClient()
	if err != nil {
		return nil, err
	}

	r := http.NewServeMux()
	r.HandleFunc("GET /", handleHomePage(sessions))
	r.HandleFunc("GET /login", handleLoginPage)
	r.HandleFunc("GET /logout", handleLogout(sessions))
	r.HandleFunc("GET /auth/google/login", handleAuth(googleAuth, sessions))
	r.HandleFunc("GET /auth/google/callback", handleAuthCallback(googleAuth, sessions))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	return srv, nil
}
