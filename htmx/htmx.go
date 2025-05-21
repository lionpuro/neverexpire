package htmx

import (
	"log"
	"net/http"

	"github.com/lionpuro/trackcerts/views"
)

func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("HX-Retarget", "#error-container")
	if err := views.ErrorBanner(w, err); err != nil {
		log.Printf("render error: %v", err)
	}
}

func IsHXrequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
