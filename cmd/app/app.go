package main

import (
	"fmt"
	"log"

	"github.com/lionpuro/neverexpire/api"
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/domain"
	"github.com/lionpuro/neverexpire/http"
	"github.com/lionpuro/neverexpire/logging"
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
	ks := api.NewKeyService(api.NewKeyRepository(pool))
	if err != nil {
		log.Fatal(err)
	}

	logger := logging.NewLogger()
	webh := http.NewHandler(logger, us, ds, as, ks)
	apih := api.NewHandler(logger, us, ds, ks)
	srv := http.NewServer(3000, webh, apih)

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
