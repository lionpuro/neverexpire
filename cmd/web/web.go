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
	"github.com/lionpuro/neverexpire/notifications"
	"github.com/lionpuro/neverexpire/users"
	"github.com/lionpuro/neverexpire/web"
)

func main() {
	conf := config.FromEnv()

	pool, err := db.NewPool(conf.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}

	us := users.NewService(users.NewRepository(pool))
	hs := hosts.NewService(hosts.NewRepository(pool))
	ks := keys.NewService(keys.NewRepository(pool))
	ns := notifications.NewService(notifications.NewRepository(pool))
	auth, err := auth.NewAuthenticator(conf)
	if err != nil {
		log.Fatal(err)
	}

	logger := logging.NewLogger()

	mux := http.NewServeMux()

	webh := web.NewHandler(logger, us, hs, ks, ns, auth)

	mux.Handle("/", web.NewRouter(webh))
	api.New(mux, logger, us, hs, ks).Register()

	srv := newServer(3000, mux)

	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func newServer(port int, mux *http.ServeMux) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return srv
}
