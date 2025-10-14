package config

type Config struct {
	ListenAddr string   `toml:"listen_addr"`
	APIKeys    []string `toml:"api_keys"`
}

func DefaultConfig() Config {
	return Config{
		ListenAddr: ":8080",
		APIKeys:    []string{"default"},
	}
}
