package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	r := http.NewServeMux()
	r.HandleFunc("GET /", homeHandler)
	r.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	fmt.Printf("Listening on %s...\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
