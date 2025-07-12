package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lionpuro/neverexpire/api"
	"github.com/lionpuro/neverexpire/auth"
	"github.com/lionpuro/neverexpire/config"
	"github.com/lionpuro/neverexpire/db"
	"github.com/lionpuro/neverexpire/hosts"
	"github.com/lionpuro/neverexpire/keys"
	"github.com/lionpuro/neverexpire/logging"
	"github.com/lionpuro/neverexpire/user"
	"github.com/lionpuro/neverexpire/web"
)

func main() {
	conf := config.FromEnv()

	pool, err := db.NewPool(conf.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}

	us := user.NewService(user.NewRepository(pool))
	hs := hosts.NewService(hosts.NewRepository(pool))
	as, err := auth.NewService(conf)
	ks := keys.NewService(keys.NewRepository(pool))
	if err != nil {
		log.Fatal(err)
	}

	logger := logging.NewLogger()
	webh := web.NewHandler(logger, us, hs, ks, as)
	apih := api.NewHandler(logger, us, hs, ks)
	srv := newServer(3000, web.NewRouter(webh), api.NewRouter(apih))

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func newServer(port int, web *http.ServeMux, api *http.ServeMux) *http.Server {
	r := http.NewServeMux()

	r.Handle("/", web)
	r.Handle("/api/", http.StripPrefix("/api", api))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	return srv
}
