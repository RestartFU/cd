package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/restartfu/cd/internal/protocol"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorBold   = "\033[1m"
)

func main() {
	var (
		serverAddr  = flag.String("server", "localhost:8080", "Server address")
		apiKey      = flag.String("key", "", "API key for authentication")
		gitURL      = flag.String("git", "", "Git repository URL")
		environment = flag.String("env", "production", "Environment name")
		envFile     = flag.String("env-file", "", "Path to environment variables file (.env format)")
		envVars     = flag.String("env-vars", "", "Comma-separated list of environment variables (KEY=VALUE,KEY2=VALUE2)")
		secretsFile = flag.String("secrets-file", "", "Path to secrets file (.env format)")
		secrets     = flag.String("secrets", "", "Comma-separated list of secrets (KEY=VALUE,KEY2=VALUE2)")
		diagnose    = flag.Bool("diagnose", false, "Run Docker diagnostics")
		help        = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *diagnose {
		runDockerDiagnostics()
		return
	}

	if *gitURL == "" {
		runInteractiveMode(*serverAddr, *apiKey, *environment)
		return
	}

	envVarMap, err := parseEnvironmentVariables(*envFile, *envVars)
	if err != nil {
		fmt.Printf("%sError parsing environment variables:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	secretsMap, err := parseEnvironmentVariables(*secretsFile, *secrets)
	if err != nil {
		fmt.Printf("%sError parsing secrets:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	if *apiKey == "" {
		fmt.Printf("%sError:%s API key is required\n", colorRed, colorReset)
		os.Exit(1)
	}

	if err := runDeployment(*serverAddr, *apiKey, *gitURL, *environment, envVarMap, secretsMap); err != nil {
		fmt.Printf("%sError:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf("%sCD CLI Tool - Continuous Deployment Client%s\n\n", colorBold, colorReset)
	fmt.Println("Usage:")
	fmt.Printf("  %scd-cli%s [options]\n\n", colorCyan, colorReset)
	fmt.Println("Options:")
	fmt.Printf("  %s-server%s string    Server address (default \"localhost:8080\")\n", colorYellow, colorReset)
	fmt.Printf("  %s-key%s string       API key for authentication\n", colorYellow, colorReset)
	fmt.Printf("  %s-git%s string       Git repository URL\n", colorYellow, colorReset)
	fmt.Printf("  %s-env%s string       Environment name (default \"production\")\n", colorYellow, colorReset)
	fmt.Printf("  %s-env-file%s string  Path to environment variables file (.env format)\n", colorYellow, colorReset)
	fmt.Printf("  %s-env-vars%s string  Comma-separated environment variables (KEY=VALUE,KEY2=VALUE2)\n", colorYellow, colorReset)
	fmt.Printf("  %s-secrets-file%s string Path to secrets file (.env format)\n", colorYellow, colorReset)
	fmt.Printf("  %s-secrets%s string   Comma-separated secrets (KEY=VALUE,KEY2=VALUE2)\n", colorYellow, colorReset)
	fmt.Printf("  %s-ssh-key%s string   Path to SSH private key file for git cloning\n", colorYellow, colorReset)
	fmt.Printf("  %s-diagnose%s         Run Docker connection diagnostics\n", colorYellow, colorReset)
	fmt.Printf("  %s-help%s             Show this help message\n", colorYellow, colorReset)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s# Interactive mode%s\n", colorGray, colorReset)
	fmt.Printf("  cd-cli\n\n")
	fmt.Printf("  %s# Deploy specific repository%s\n", colorGray, colorReset)
	fmt.Printf("  cd-cli -key=mykey -git=https://github.com/user/repo.git\n\n")
	fmt.Printf("  %s# Deploy to staging environment%s\n", colorGray, colorReset)
	fmt.Printf("  cd-cli -key=mykey -git=https://github.com/user/repo.git -env=staging\n\n")
	fmt.Printf("  %s# Deploy with environment variables%s\n", colorGray, colorReset)
	fmt.Printf("  cd-cli -key=mykey -git=https://github.com/user/repo.git -env-vars=\"API_KEY=value,DEBUG=true\"\n\n")
	fmt.Printf("  %s# Deploy with environment file%s\n", colorGray, colorReset)
	fmt.Printf("  cd-cli -key=mykey -git=https://github.com/user/repo.git -env-file=.env\n")
}

func runInteractiveMode(serverAddr, defaultAPIKey, defaultEnv string) {
	fmt.Printf("%s=== CD CLI Interactive Mode ===%s\n\n", colorBold+colorCyan, colorReset)

	reader := bufio.NewReader(os.Stdin)

	if serverAddr == "localhost:8080" {
		fmt.Printf("Server address [%s]: ", serverAddr)
		if input := readLine(reader); input != "" {
			serverAddr = input
		}
	}

	apiKey := defaultAPIKey
	if apiKey == "" {
		fmt.Print("API key: ")
		apiKey = readLine(reader)
		if apiKey == "" {
			fmt.Printf("%sError:%s API key is required\n", colorRed, colorReset)
			os.Exit(1)
		}
	}

	for {
		fmt.Printf("\n%s=== New Deployment ===%s\n", colorBold+colorGreen, colorReset)

		fmt.Print("Git repository URL: ")
		gitURL := readLine(reader)
		if gitURL == "" {
			fmt.Printf("%sError:%s Git URL is required\n", colorRed, colorReset)
			continue
		}

		environment := defaultEnv
		fmt.Printf("Environment [%s]: ", environment)
		if input := readLine(reader); input != "" {
			environment = input
		}

		fmt.Printf("\n%sStarting deployment...%s\n", colorYellow, colorReset)
		if err := runDeployment(serverAddr, apiKey, gitURL, environment, nil, nil); err != nil {
			fmt.Printf("%sDeployment failed:%s %v\n", colorRed, colorReset, err)
		}

		fmt.Printf("\n%sDeploy another repository? (y/N): %s", colorCyan, colorReset)
		if response := readLine(reader); !strings.HasPrefix(strings.ToLower(response), "y") {
			break
		}
	}

	fmt.Printf("\n%sGoodbye!%s\n", colorGreen, colorReset)
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func parseEnvironmentVariables(filePath, envVarsString string) (map[string]string, error) {
	envVars := make(map[string]string)

	if filePath != "" {
		if err := parseEnvFile(filePath, envVars); err != nil {
			return nil, fmt.Errorf("failed to parse environment file %s: %w", filePath, err)
		}
	}

	if envVarsString != "" {
		if err := parseEnvString(envVarsString, envVars); err != nil {
			return nil, fmt.Errorf("failed to parse environment variables: %w", err)
		}
	}

	return envVars, nil
}

func parseEnvFile(filePath string, envVars map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		envVars[key] = value
	}

	return scanner.Err()
}

func parseEnvString(envVarsString string, envVars map[string]string) error {
	pairs := strings.Split(envVarsString, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format: %s (expected KEY=VALUE)", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}

	return nil
}

func runDockerDiagnostics() {
	fmt.Printf("%s=== Docker Diagnostics ===%s\n", colorBold+colorCyan, colorReset)
	fmt.Println()

	fmt.Printf("%sSystem Information:%s\n", colorYellow, colorReset)
	fmt.Printf("Operating System: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Println()

	fmt.Printf("%sEnvironment Variables:%s\n", colorYellow, colorReset)
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		fmt.Printf("DOCKER_HOST: %s\n", host)
	} else {
		fmt.Printf("DOCKER_HOST: %snot set%s\n", colorGray, colorReset)
	}
	fmt.Println()

	fmt.Printf("%sDocker Socket Check:%s\n", colorYellow, colorReset)

	var socketPaths []string
	switch runtime.GOOS {
	case "darwin":
		socketPaths = []string{
			os.ExpandEnv("$HOME/.docker/run/docker.sock"),
			"/var/run/docker.sock",
		}
	case "linux":
		socketPaths = []string{"/var/run/docker.sock"}
	case "windows":
		fmt.Printf("Windows: Using named pipe ////./pipe/docker_engine\n")
		socketPaths = []string{}
	}

	for _, path := range socketPaths {
		if stat, err := os.Stat(path); err == nil {
			fmt.Printf("✅ Found: %s%s%s (size: %d bytes)\n", colorGreen, path, colorReset, stat.Size())
		} else {
			fmt.Printf("❌ Missing: %s%s%s (%v)\n", colorRed, path, colorReset, err)
		}
	}

	fmt.Println()
	fmt.Printf("%sRecommendations:%s\n", colorYellow, colorReset)

	switch runtime.GOOS {
	case "darwin":
		fmt.Println("• Ensure Docker Desktop is running")
		fmt.Println("• Check the Docker whale icon in your menu bar")
		fmt.Println("• Try restarting Docker Desktop if socket is missing")
	case "linux":
		fmt.Println("• Ensure Docker daemon is running: sudo systemctl start docker")
		fmt.Println("• Add your user to docker group: sudo usermod -aG docker $USER")
		fmt.Println("• Check Docker service status: sudo systemctl status docker")
	case "windows":
		fmt.Println("• Ensure Docker Desktop is running")
		fmt.Println("• Check Docker Desktop in system tray")
	}

	fmt.Println()
	fmt.Printf("%sNext Steps:%s\n", colorYellow, colorReset)
	fmt.Println("• Test Docker: docker run hello-world")
	fmt.Println("• Start CD server: ./bin/cd-server")
	fmt.Println("• Try deployment again")
}

func runDeployment(serverAddr, apiKey, gitURL, environment string, envVars, secrets map[string]string) error {
	client := protocol.NewClient()

	fmt.Printf("%s[INFO]%s Connecting to server %s...\n", colorBlue, colorReset, serverAddr)
	if err := client.Connect(serverAddr); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	fmt.Printf("%s[INFO]%s Authenticating...\n", colorBlue, colorReset)
	if err := client.Authenticate(apiKey); err != nil {
		return fmt.Errorf("failed to send auth: %w", err)
	}

	if err := client.WaitForAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("%s[SUCCESS]%s Authenticated successfully\n", colorGreen, colorReset)

	updateChan, err := client.Deploy(gitURL, environment, envVars, secrets)
	if err != nil {
		return fmt.Errorf("failed to start deployment: %w", err)
	}

	var finalResult *protocol.ResultMessage
	var finalError *protocol.ErrorMessage

	fmt.Printf("\n%s=== Deployment Log ===%s\n", colorBold+colorPurple, colorReset)

	for update := range updateChan {
		switch update.Type {
		case protocol.MessageTypeLog:
			if update.Log != nil {
				printLogMessage(update.Log)
			}

		case protocol.MessageTypeStatus:
			if update.Status != nil {
				printStatusMessage(update.Status)
			}

		case protocol.MessageTypeResult:
			if update.Result != nil {
				finalResult = update.Result
			}

		case protocol.MessageTypeError:
			if update.Error != nil {
				finalError = update.Error
			}
		}
	}

	fmt.Printf("\n%s=== Deployment Result ===%s\n", colorBold+colorCyan, colorReset)

	if finalError != nil {
		fmt.Printf("%s[FAILED]%s %s (Code: %s)\n", colorRed+colorBold, colorReset, finalError.Message, finalError.Code)
		return fmt.Errorf("deployment failed: %s", finalError.Message)
	}

	if finalResult != nil {
		if finalResult.Success {
			fmt.Printf("%s[SUCCESS]%s %s\n", colorGreen+colorBold, colorReset, finalResult.Message)
			fmt.Printf("%s[INFO]%s Duration: %s\n", colorBlue, colorReset, finalResult.Duration)
			if finalResult.ContainerID != "" {
				fmt.Printf("%s[INFO]%s Container ID: %s\n", colorBlue, colorReset, finalResult.ContainerID)
			}
		} else {
			fmt.Printf("%s[FAILED]%s %s\n", colorRed+colorBold, colorReset, finalResult.Message)
			return fmt.Errorf("deployment failed: %s", finalResult.Message)
		}
	}

	return nil
}

func printLogMessage(log *protocol.LogMessage) {
	var levelColor string
	var levelText string

	switch strings.ToLower(log.Level) {
	case "error":
		levelColor = colorRed
		levelText = "ERROR"
	case "warn", "warning":
		levelColor = colorYellow
		levelText = "WARN"
	case "info":
		levelColor = colorBlue
		levelText = "INFO"
	default:
		levelColor = colorGray
		levelText = strings.ToUpper(log.Level)
	}

	var sourceColor string
	switch strings.ToLower(log.Source) {
	case "git":
		sourceColor = colorPurple
	case "docker":
		sourceColor = colorCyan
	case "system":
		sourceColor = colorGray
	default:
		sourceColor = colorGray
	}

	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s%s%s %s[%s]%s %s[%s]%s %s\n",
		colorGray, timestamp, colorReset,
		levelColor, levelText, colorReset,
		sourceColor, strings.ToUpper(log.Source), colorReset,
		log.Message)
}

func printStatusMessage(status *protocol.StatusMessage) {
	var stageColor string
	switch strings.ToLower(status.Stage) {
	case "complete", "authenticated":
		stageColor = colorGreen
	case "starting", "preparing", "cloning", "building":
		stageColor = colorYellow
	default:
		stageColor = colorBlue
	}

	progressBar := generateProgressBar(status.Progress, 30)

	fmt.Printf("\n%s[STATUS]%s %s%s%s (%d%%) %s\n",
		colorBold+colorBlue, colorReset,
		stageColor, strings.ToUpper(status.Stage), colorReset,
		status.Progress,
		progressBar)

	if status.Message != "" {
		fmt.Printf("%s         %s%s\n", colorGray, status.Message, colorReset)
	}
}

func generateProgressBar(progress, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	filled := (progress * width) / 100
	empty := width - filled

	bar := colorGreen + strings.Repeat("█", filled) + colorGray + strings.Repeat("░", empty) + colorReset
	return fmt.Sprintf("[%s]", bar)
}
