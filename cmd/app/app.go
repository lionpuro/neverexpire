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
	conf, err := config.FromEnvFile(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	conn := db.ConnString(
		conf.PostgresUser,
		conf.PostgresPassword,
		conf.PostgresHost,
		conf.PostgresPort,
		conf.PostgresDB,
	)
	pool, err := db.NewPool(conn)
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
	srv := http.NewServer(h)

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
