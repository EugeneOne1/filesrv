package cmd

import (
	"errors"
	"log"
	"net"
	"net/http"

	"filesrv/internal/dirs"
	"filesrv/internal/fhttp"
)

// dieOnErr logs the error and exits if it is not nil.
func dieOnErr(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func Serve() {
	// Configure.
	fsys := http.Dir(".")
	h := http.Handler(
		&dirs.Dirs{
			Handler:    http.FileServer(fsys),
			FileSystem: fsys,
		},
	)

	// Wrap.
	h = fhttp.Wrap(h, withLog)

	// Listen.
	const port = "6060"

	ln, err := net.Listen("tcp", ":"+port)
	dieOnErr(err)

	err = printListenAddrs(port)
	dieOnErr(err)

	// Serve.
	err = http.Serve(ln, h)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("serving terminated: %s", err)
	}
}
