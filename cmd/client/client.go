package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"wise-tcp/internal/pow/providers/hashcash"
	"wise-tcp/pkg/config"
	"wise-tcp/pkg/log"
	"wise-tcp/pkg/zap"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Client ClientConfig `yaml:"client"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Prod bool   `yaml:"isProd"`
}

type ClientConfig struct {
	ServerAddr string `yaml:"serverAddr" env:"SERVER_ADDR"`
	BeaconAddr string `yaml:"beaconAddr" env:"BEACON_ADDR"`
	Async      bool   `yaml:"async" env:"ASYNC"`
	TryReplay  bool   `yaml:"tryReplay" env:"TRY_REPLAY"`
}

func main() {
	cfg := mustLoadConfig()

	initLogger(cfg.App)

	var fn func(cfg *Config) (string, error)

	if cfg.Client.Async {
		fn = getQuoteAsync
	} else {
		fn = getQuoteSync
	}

	resp, err := fn(cfg)
	if err != nil {
		log.Errorf("Failed to get quote: %v", err)
		return
	}

	fmt.Println(resp)
}

func getQuoteSync(cfg *Config) (string, error) {
	quote, solution, err := getQuote(cfg, "")
	if cfg.Client.TryReplay {
		quote, _, err = getQuote(cfg, solution)
		if err != nil {
			return "", err
		}
	}
	return quote, err
}

func getQuote(cfg *Config, replay string) (string, string, error) {
	conn, err := connect(cfg)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Errorf("Failed to close connection: %v", closeErr)
		}
	}()

	challenge, err := receiveChallenge(conn)
	if err != nil {
		return "", "", fmt.Errorf("failed to receive challenge: %v", err)
	}

	log.Debugf("Received challenge: %s", challenge)

	var solution string
	if replay == "" {
		solver := hashcash.NewSolver()
		solution, err = solver.Solve(challenge)
		if err != nil {
			return "", "", fmt.Errorf("failed to solve challenge: %v", err)
		}
		log.Infof("Solved solution: %s", solution)
	} else {
		solution = replay
		log.Debugf("Replaying solution: %s", solution)
	}

	response := []byte("X-Response: " + solution)
	if err = sendMessage(conn, response); err != nil {
		return "", "", fmt.Errorf("failed to send solution: %v", err)
	}

	quote, err := receiveMessage(conn)
	if err != nil {
		return "", "", fmt.Errorf("failed to receive quote message: %v", err)
	}

	log.Debug("Received random quote: %s", quote)

	return quote, solution, nil
}

func getQuoteAsync(cfg *Config) (string, error) {
	udpAddr := "127.0.0.1:9002"
	udpConn, err := net.Dial("udp", udpAddr)
	if err != nil {
		return "", fmt.Errorf("failed to connect challenge beacon: %w", err)
	}
	defer func(udpConn net.Conn) {
		err := udpConn.Close()
		if err != nil {
			log.Errorf("Failed to close UDP connection: %v", err)
		}
	}(udpConn)

	request := "X-Request: challenge"
	_, err = udpConn.Write([]byte(request))
	if err != nil {
		return "", fmt.Errorf("failed to request challenge: %v", err)
	}

	timeout := 2 * time.Second
	err = udpConn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", fmt.Errorf("failed to set read deadline: %v", err)
	}

	buffer := make([]byte, 1024)
	n, err := udpConn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read challenge: %v", err)
	}

	challenge := strings.TrimSpace(strings.TrimPrefix(string(buffer[:n]), "X-Challenge:"))
	log.Debugf("Received challenge: %s", challenge)

	solver := hashcash.NewSolver()
	solution, err := solver.Solve(challenge)
	if err != nil {
		return "", fmt.Errorf("failed to solve challenge: %v", err)
	}
	log.Infof("Solution generated: %s", solution)

	tcpConn, err := connect(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to connect to resource server: %w", err)
	}
	defer func(tcpConn net.Conn) {
		err := tcpConn.Close()
		if err != nil {
			log.Error(err)
		}
	}(tcpConn)

	_, err = tcpConn.Write([]byte("X-Response: " + solution + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send solution: %v", err)
	}

	quote, err := bufio.NewReader(tcpConn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to receive response: %v", err)
	}

	return strings.TrimSpace(quote), nil
}

func connect(cfg *Config) (net.Conn, error) {
	serverAddr := cfg.Client.ServerAddr
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	return conn, nil
}

func receiveMessage(conn net.Conn) (string, error) {
	err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
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

func initLogger(cfg AppConfig) {
	logger, err := zap.New(
		zap.WithName(cfg.Name),
		zap.WithProd(cfg.Prod),
	)
	if err != nil {
		log.Errorf("Failed to initialize zap logger: %v", err)
		return
	}

	log.SetLogger(logger)
}

func applyConfigMapping(v *viper.Viper) error {
	if err := v.BindEnv("client.serverAddr", "SERVER_ADDR"); err != nil {
		return fmt.Errorf("failed to bind SERVER_ADDR: %w", err)
	}
	if err := v.BindEnv("client.beaconAddr", "BEACON_ADDR"); err != nil {
		return fmt.Errorf("failed to bind BEACON_ADDR: %w", err)
	}
	if err := v.BindEnv("client.async", "ASYNC"); err != nil {
		return fmt.Errorf("failed to bind ASYNC: %w", err)
	}
	if err := v.BindEnv("client.tryReplay", "TRY_REPLAY"); err != nil {
		return fmt.Errorf("failed to bind TRY_REPLAY: %w", err)
	}

	return nil
}
