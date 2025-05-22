package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/trackcerts/auth"
	"github.com/lionpuro/trackcerts/db"
	"github.com/lionpuro/trackcerts/domain"
	"github.com/lionpuro/trackcerts/user"
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
	httpServer *http.Server
}

func newServer() (*Server, error) {
	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_HOST_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	pool, err := db.NewPool(conn)
	if err != nil {
		return nil, err
	}

	us := user.NewService(&user.UserRepository{DB: pool})
	ds := domain.NewService(&domain.DomainRepository{DB: pool})
	as, err := auth.NewService()
	if err != nil {
		return nil, err
	}

	dh := domain.NewHandler(ds)
	ah, err := auth.NewHandler(as, us)
	if err != nil {
		return nil, err
	}

	r := http.NewServeMux()

	register := func(p string, h http.HandlerFunc) {
		r.HandleFunc(p, ah.Authenticate(h))
	}

	register("GET /", handleHomePage)
	register("GET /domains", requireAuth(dh.Domains))
	register("GET /domains/new", requireAuth(dh.NewDomainPage))
	register("POST /domains", requireAuth(dh.CreateDomain))
	register("GET /domains/{id}", requireAuth(dh.Domain(false)))
	register("GET /partials/domains/{id}", requireAuth(dh.Domain(true)))
	register("DELETE /domains/{id}", requireAuth(dh.DeleteDomain))
	register("GET /account", requireAuth(handleAccountPage))
	register("GET /login", handleLoginPage)
	register("GET /logout", ah.Logout)

	r.HandleFunc("GET /auth/google/login", ah.Login(as.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", ah.Callback(as.GoogleClient))
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("assets/public"))))

	s := &Server{
		httpServer: &http.Server{
			Addr:    ":3000",
			Handler: r,
		},
	}
	return s, nil
}
