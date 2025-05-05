package main

import (
	"log"
	"net/http"

	"github.com/lionpuro/trackcert/views"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err := views.Home(w); err != nil {
		log.Printf(err.Error())
	}
}
