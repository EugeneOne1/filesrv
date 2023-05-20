package cmd

import (
	"log"
	"net/http"

	"filesrv/internal/fhttp"
)

// type check
var _ fhttp.Middleware = withLog

// withLog logs the request path and remote address.
func withLog(h http.Handler) (wrapped http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\t%s\tfrom\t%s", r.URL.Path, r.RemoteAddr)
		h.ServeHTTP(w, r)
		log.Printf("\t%s\tfrom\t%s: finished", r.URL.Path, r.RemoteAddr)
	})
}
