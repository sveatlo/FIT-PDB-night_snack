package config

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/sveatlo/night_snack/internal/config"
)

type Config struct {
	config.Base `mapstructure:",squash"`

	Database struct {
		Host     string
		Port     int
		Username string
	}
}

func NewConfig(files ...string) (c Config, err error) {
	c = Config{
		Base: config.NewCommonConfig(),
	}

	c.Base.ListenAddressHTTP = ":1757"
	c.Base.ListenAddressPrometheus = ":1757"
	c.Base.ListenAddressGRPC = ":1758"
	c.Base.Loglevel = zerolog.InfoLevel.String()
	// database defaults for docker compose
	c.Database.Host = "cockroach"
	c.Database.Port = 26257
	c.Database.Username = "root"

	m, err := config.NewManager(&c, files...)
	if err != nil {
		err = fmt.Errorf("cannot create new config manager: %w", err)
		return
	}

	err = m.Load()
	if err != nil {
		err = fmt.Errorf("cannot load configuration: %w", err)
		return
	}

	return
}

func (c *Config) PostLoad() (err error) {
	return
}
