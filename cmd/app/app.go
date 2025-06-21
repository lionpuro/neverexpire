package main

import (
	"fmt"
	"log"

	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/http"
	"github.com/lionpuro/neverexpire/user"
)

func main() {
	conf := config.FromEnv()

	pool, err := db.NewPool(conf.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}

	us := user.NewService(user.NewRepository(pool))
	ds := domain.NewService(domain.NewRepository(pool))
	as, err := auth.NewService(conf)
	if err != nil {
		log.Fatal(err)
	}

	h := http.NewHandler(us, ds, as)
	srv := http.NewServer(3000, h)

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
