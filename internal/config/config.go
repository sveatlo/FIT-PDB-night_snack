package config

type Config interface {
	PostLoad() error
}
