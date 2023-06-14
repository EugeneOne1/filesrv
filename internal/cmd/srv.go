package cmd

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/http"

	"filesrv/internal/dirs"
	"filesrv/internal/dirs/themes"
	"filesrv/internal/fhttp"
)

var options = struct {
	themePath *string
}{
	themePath: flag.String("t", "", "path to the theme directory"),
}

// dieOnErr logs the error and exits if it is not nil.
func dieOnErr(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func Serve() {
	// Parse options.
	flag.Parse()

	theme := themes.DefaultEmbedded()
	if p := *options.themePath; p != "" {
		theme = themes.DefaultDynamic(p)
	}

	// Configure.
	fsys := http.Dir(".")
	h, err := dirs.NewHTTPFSDirs(&dirs.HTTPFSConfig{
		FS:    fsys,
		Theme: theme,
	})
	dieOnErr(err)

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
