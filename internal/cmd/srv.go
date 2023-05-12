package cmd

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
)

// dieOnErr logs the error and exits if it is not nil.
func dieOneErr(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func Serve() {
	// Configure.
	h := http.FileServer(http.FS(os.DirFS(".")))

	// Wrap.
	h = withMiddlewares(h, withLog)

	// Listen.
	const port = "6060"

	ln, err := net.Listen("tcp", ":"+port)
	dieOneErr(err)

	err = printListenAddrs(port)
	dieOneErr(err)

	// Serve.
	err = http.Serve(ln, h)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("serving terminated: %s", err)
	}
}
