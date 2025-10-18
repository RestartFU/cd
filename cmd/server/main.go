package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/restartfu/cd/internal"
	"github.com/restartfu/cd/internal/config"
	"github.com/restartfu/gophig"
)

func main() {
	cfg, err := loadConfig("config.toml")
	if err != nil {
		log.Fatalln(err)
	}

	server, err := internal.Assemble(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		if err := server.Stop(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	log.Printf("Starting TCP server on %s", cfg.ListenAddr)
	if err := server.Start(cfg.ListenAddr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
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
