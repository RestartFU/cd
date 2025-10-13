package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"

	"github.com/restartfu/cd/internal"
	"github.com/restartfu/cd/internal/config"
	"github.com/restartfu/gophig"
)

func main() {
	cfg, err := loadConfig("config.toml")
	if err != nil {
		log.Fatalln(err)
	}

	service, err := internal.Assemble(cfg, func(r *http.Request) bool {
		apiKey := http.Header(r.Header).Get("API_KEY")
		if slices.Contains(cfg.APIKeys, apiKey) {
			return true
		}
		return false
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatalln(service.Start(cfg.ListenAddr))
}

func loadConfig(configPath string) (config.Config, error) {
	defaultConfig := config.DefaultConfig()

	g := gophig.NewGophig[config.Config](configPath, gophig.TOMLMarshaler{}, os.ModePerm)
	conf, err := g.LoadConf()
	if err != nil {
		if os.IsNotExist(err) {
			if saveErr := g.SaveConf(defaultConfig); saveErr != nil {
				return defaultConfig, fmt.Errorf("failed to save default config: %w", saveErr)
			}
			return defaultConfig, nil
		}
		return config.Config{}, fmt.Errorf("failed to load config: %w", err)
	}
	return conf, nil
}
