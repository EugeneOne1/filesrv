package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v8"
)

type environments struct {
	// themePath is the path to the theme assets directory.  If empty, the
	// embedded theme is used.
	ThemePath string `env:"THEME_PATH" envDefault:""`

	// listenHost is the host to listen on.
	ListenHost string `env:"HOST" envDefault:""`

	// listenPort is the port to listen on.
	ListenPort uint16 `env:"PORT" envDefault:"6060"`
}

func parseEnvs() (envs environments, err error) {
	for _, e := range os.Environ() {
		switch firstEqual := strings.IndexByte(e, '='); e[:firstEqual] {
		case "THEME_PATH", "HOST", "PORT":
			log.Printf("env %q is set to %q", e[:firstEqual], e[firstEqual+1:])
		default:
			continue
		}
	}

	err = env.Parse(&envs)

	return envs, err
}
