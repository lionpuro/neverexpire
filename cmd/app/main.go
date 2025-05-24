package main

import (
	"fmt"
	"log"
	netHTTP "net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/trackcerts/auth"
	"github.com/lionpuro/trackcerts/db"
	"github.com/lionpuro/trackcerts/domain"
	"github.com/lionpuro/trackcerts/http"
	"github.com/lionpuro/trackcerts/user"
)

func main() {
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
		log.Fatal(err)
	}

	us := user.NewService(user.NewRepository(pool))
	ds := domain.NewService(domain.NewRepository(pool))
	as, err := auth.NewService()
	if err != nil {
		log.Fatal(err)
	}

	r := netHTTP.NewServeMux()

	h := http.NewHandler(us, ds, as)

	handle := func(p string, hf netHTTP.HandlerFunc) {
		r.HandleFunc(p, h.Authenticate(hf))
	}

	handle("GET /", h.HomePage)
	handle("GET /domains", h.RequireAuth(h.DomainsPage))
	handle("GET /domains/new", h.RequireAuth(h.NewDomainPage))
	handle("POST /domains", h.RequireAuth(h.CreateDomain))
	handle("GET /domains/{id}", h.RequireAuth(h.DomainPage(false)))
	handle("GET /partials/domains/{id}", h.RequireAuth(h.DomainPage(true)))
	handle("DELETE /domains/{id}", h.RequireAuth(h.DeleteDomain))
	handle("GET /account", h.RequireAuth(h.AccountPage))
	handle("GET /login", h.LoginPage)
	handle("GET /logout", h.Logout)
	handle("DELETE /account", h.RequireAuth(h.DeleteUser))

	r.HandleFunc("GET /auth/google/login", h.Login(as.GoogleClient))
	r.HandleFunc("GET /auth/google/callback", h.AuthCallback(as.GoogleClient))
	r.Handle("GET /static/", netHTTP.StripPrefix("/static", netHTTP.FileServer(netHTTP.Dir("assets/public"))))

	srv := &netHTTP.Server{
		Addr:    ":3000",
		Handler: r,
	}

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
