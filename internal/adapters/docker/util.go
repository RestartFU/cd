package docker

import (
	"log"
	"os"
	"runtime"
)

func getDockerHost() string {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		log.Printf("Using DOCKER_HOST environment variable: %s", host)
		return host
	}

	log.Printf("Detecting Docker socket for OS: %s", runtime.GOOS)

	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			os.ExpandEnv("$HOME/.docker/run/docker.sock"),
			"/var/run/docker.sock",
		}

		for _, path := range paths {
			log.Printf("Checking Docker socket: %s", path)
			if _, err := os.Stat(path); err == nil {
				log.Printf("Found Docker socket: %s", path)
				return "unix://" + path
			}
		}

		log.Printf("No Docker socket found, letting Docker client auto-detect")
		return ""

	case "linux":
		path := "/var/run/docker.sock"
		log.Printf("Checking Docker socket: %s", path)
		if _, err := os.Stat(path); err == nil {
			log.Printf("Found Docker socket: %s", path)
			return "unix://" + path
		}
		log.Printf("Docker socket not found, using auto-detection")
		return ""

	case "windows":
		log.Printf("Using Windows named pipe for Docker")
		return "npipe:////./pipe/docker_engine"
	default:
		log.Printf("Unknown OS, using Docker client auto-detection")
		return ""
	}
}
