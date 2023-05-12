package cmd

import (
	"log"
	"net/http"
)

// middleware is a function that wraps an [http.Handler].
type middleware func(http.Handler) http.Handler

// withMiddlewares wraps h with mws.
func withMiddlewares(h http.Handler, mws ...middleware) (wrapped http.Handler) {
	wrapped = h
	for _, middleware := range mws {
		wrapped = middleware(wrapped)
	}

	return wrapped
}

// type check
var _ middleware = withLog

// withLog logs the request path and remote address.
func withLog(h http.Handler) (wrapped http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\t%s\tfrom\t%s", r.URL.Path, r.RemoteAddr)
		h.ServeHTTP(w, r)
	})
}
