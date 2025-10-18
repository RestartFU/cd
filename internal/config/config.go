package config

type Config struct {
	ListenAddr string `toml:"listen_addr"`
	APIKey     string `toml:"api_key"`
	SSHKeyPath string `toml:"ssh_key_path"`
}

func DefaultConfig() Config {
	return Config{
		ListenAddr: ":8080",
		APIKey:     "default",
		SSHKeyPath: "",
	}
}
