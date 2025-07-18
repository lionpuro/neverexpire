package api

import (
	"embed"
	"net/http"

	"github.com/lionpuro/neverexpire/logging"
)

//go:embed docs.html
var static embed.FS

func docsHandler(log logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		b, err := static.ReadFile("docs.html")
		if err != nil {
			log.Error("failed to read docs html", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(b); err != nil {
			log.Error("failed to write response", "error", err.Error())
		}
	}
}
