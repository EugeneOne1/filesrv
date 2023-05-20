package fhttp

import (
	"net/http"
)

// Middleware is a function that wraps an [http.Handler].
type Middleware func(http.Handler) http.Handler

// Wrap wraps h with mws.
func Wrap(h http.Handler, mws ...Middleware) (wrapped http.Handler) {
	wrapped = h
	for _, wrap := range mws {
		wrapped = wrap(wrapped)
	}

	return wrapped
}
