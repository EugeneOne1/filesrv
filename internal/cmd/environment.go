package cmd

import (
	"github.com/c2h5oh/datasize"
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

	// MaxUploadSize is the maximum size of a file that can be uploaded.  It's
	// 4GB by default.
	MaxUploadSize datasize.ByteSize `env:"MAX_UPLOAD_SIZE" envDefault:"4GB"`
}

func parseEnvs() (envs environments, err error) {
	return envs, env.Parse(&envs)
}
