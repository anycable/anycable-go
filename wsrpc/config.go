package wsrpc

import "github.com/anycable/anycable-go/ws"

type Config struct {
	Host   string
	Port   int
	Path   string
	Secret string

	WS *ws.Config
}

func NewConfig() *Config {
	wsconf := ws.NewConfig()

	return &Config{
		WS: &wsconf,
	}
}
