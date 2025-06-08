package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/http"
	"github.com/lionpuro/neverexpire/user"
)

func main() {
	conn := db.ConnString(
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

	h := http.NewHandler(us, ds, as)
	srv := http.NewServer(h)

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
