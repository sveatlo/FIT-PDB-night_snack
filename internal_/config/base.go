package config

import "time"

type Base struct {
	Loglevel  string `mapstructure:"loglevel"`
	SentryDSN string `mapstructure:"sentry_dsn"`

	ListenAddressGRPC       string `mapstructure:"listen_address_grpc"`
	ListenAddressHTTP       string `mapstructure:"listen_address_http"`
	ListenAddressPrometheus string `mapstructure:"listen_address_prometheus"`
	ListenAddressChannelz   string `mapstructure:"listen_address_channelz"`

	Registry struct {
		Instance string `mapstructure:"instance"`
		Type     string `mapstructure:"type"`
		FilePath string `mapstructure:"file_path"`
		Watch    bool   `mapstructure:"watch"`
	} `mapstructure:"registry"`

	Mongo struct {
		URI string `mapstructure:"uri"`
	} `mapstructure:"mongo"`

	NATS struct {
		Servers       string        `mapstructure:"servers"`
		MaxReconnects int           `mapstructure:"max_reconnects"`
		ReconnectWait time.Duration `mapstructure:"reconnect_wait"`
	} `mapstructure:"nats"`
}

func NewCommonConfig() (c Base) {
	c = Base{}

	c.Registry.Type = "file"
	c.Registry.FilePath = "registry.yml"

	c.Mongo.URI = "mongodb://mongo:27017"

	c.NATS.Servers = "nats://nats:4222"

	return
}

func (c *Base) PostLoad() (err error) {
	return
}
