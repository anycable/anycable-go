package wsrpc

type Config struct {
	Port   int
	Path   string
	Secret string
}

func NewConfig() *Config {
	return &Config{}
}
