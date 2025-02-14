package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"wise-tcp/internal/pow/hashcash"
	"wise-tcp/pkg/config"
	"wise-tcp/pkg/log"
	"wise-tcp/pkg/log/zap"
)

type Config struct {
	Logger log.Config   `yaml:"logger"`
	Client ClientConfig `yaml:"client"`
}

type ClientConfig struct {
	ServerAddr string `yaml:"serverAddr" env:"SERVER_ADDR"`
	TryReplay  bool   `yaml:"tryReplay" env:"TRY_REPLAY"`
}

func main() {
	cfg := mustLoadConfig()

	logger := initLogger(cfg.Logger)
	logger.Debug("Client application starting...")

	quote, solution, err := getQuote(cfg, logger, "")
	if err != nil {
		logger.Error("Failed to get quote: %v", err)
		return
	}

	fmt.Println(quote)

	if cfg.Client.TryReplay {
		quote, _, err = getQuote(cfg, logger, solution)
		if err != nil {
			logger.Error("Failed to get quote with replayed solution: %v", err)
			return
		}

		fmt.Println(quote)
	}
}

func getQuote(cfg *Config, logger log.Logger, replay string) (string, string, error) {
	conn, err := connect(cfg, logger)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Error("Failed to close connection: %v", closeErr)
		}
	}()

	challenge, err := receiveChallenge(conn)
	if err != nil {
		return "", "", fmt.Errorf("failed to receive challenge: %v", err)
	}

	logger.Debug("Received challenge: %s", challenge)

	var solution string
	if replay == "" {
		solver := hashcash.NewSolver()
		solution, err = solver.Solve(challenge)
		if err != nil {
			return "", "", fmt.Errorf("failed to solve challenge: %v", err)
		}
		logger.Info("Solved solution: %s", solution)
	} else {
		solution = replay
		logger.Debug("Replaying solution: %s", solution)
	}

	response := []byte("X-Response: " + solution)
	if err = sendMessage(conn, response); err != nil {
		return "", "", fmt.Errorf("failed to send solution: %v", err)
	}

	quote, err := receiveMessage(conn)
	if err != nil {
		return "", "", fmt.Errorf("failed to receive quote message: %v", err)
	}

	logger.Debug("Received random quote: %s", quote)

	return quote, solution, nil
}

func connect(cfg *Config, logger log.Logger) (net.Conn, error) {
	serverAddr := cfg.Client.ServerAddr
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	logger.Debug("Connecting to server at %s...", serverAddr)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	logger.Debug("Connected to server: %s", conn.RemoteAddr())
	return conn, nil
}

func receiveMessage(conn net.Conn) (string, error) {
	err := conn.SetReadDeadline(time.Now().Add(50 * time.Second))
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(conn)

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read message: %w", err)
	}

	return strings.TrimSpace(response), nil
}

func sendMessage(conn net.Conn, solution []byte) error {
	err := conn.SetWriteDeadline(time.Now().Add(50 * time.Second))
	if err != nil {
		return err
	}
	if _, err = conn.Write(solution); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func receiveChallenge(conn net.Conn) (string, error) {
	msg, err := receiveMessage(conn)
	if err != nil {
		return "", err
	}

	challenge := strings.TrimSpace(strings.TrimPrefix(msg, "X-Challenge:"))

	return challenge, nil
}

func mustLoadConfig() *Config {
	return config.MustLoad[Config]("cfg/client.yml",
		config.WithEnvMapper[Config](applyConfigMapping))
}

func initLogger(cfg log.Config) log.Logger {
	return zap.Init(cfg)
}

func applyConfigMapping(v *viper.Viper) error {
	if err := v.BindEnv("client.serverAddr", "SERVER_ADDR"); err != nil {
		return fmt.Errorf("failed to bind SERVER_ADDR: %w", err)
	}
	if err := v.BindEnv("client.tryReplay", "TRY_REPLAY"); err != nil {
		return fmt.Errorf("failed to bind TRY_REPLAY: %w", err)
	}

	return nil
}
