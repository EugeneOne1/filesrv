package cmd

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"filesrv/internal/dirs"
	"filesrv/internal/dirs/themes"
	"filesrv/internal/fhttp"
)

// dieOnErr logs the error and exits if it is not nil.
func dieOnErr(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func Serve() {
	// Parse.
	envs, err := parseEnvs()
	dieOnErr(err)

	// Load.
	var theme dirs.Theme
	if p := envs.ThemePath; p != "" {
		theme = themes.DefaultDynamic(os.DirFS(p))
	} else {
		theme = themes.DefaultEmbedded()
	}
	log.Printf("using theme: %s", theme)

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
	port := strconv.Itoa(int(envs.ListenPort))
	ln, err := net.Listen("tcp", net.JoinHostPort(envs.ListenHost, port))
	dieOnErr(err)

	err = printListenAddrs(port)
	dieOnErr(err)

	// Serve.
	err = http.Serve(ln, h)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("serving terminated: %s", err)
	}
}
