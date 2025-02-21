package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-redis/redis/v8"

	"wise-tcp/internal/pow/providers/hashcash"
	"wise-tcp/pkg/log"
)

var redisClient *redis.Client
var ctx = context.Background()

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func main() {
	addr := net.UDPAddr{
		Port: 9002,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error creating UDP server:", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Errorf("Failed to close connection: %v", closeErr)
		}
	}()

	// todo: configurable difficulty
	provider := hashcash.NewProvider(hashcash.WithDifficulty(20))

	fmt.Println("Beacon server is running on port 9002...")

	for {
		buffer := make([]byte, 1)

		_, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Error reading client address: %v\n", err)
			continue
		}

		go handleConnection(conn, clientAddr, provider)
	}
}

func handleConnection(serverConn *net.UDPConn, clientAddr *net.UDPAddr, provider *hashcash.Provider) {
	raw, err := provider.RawChallenge(clientAddr.String(), 0)
	if err != nil {
		log.Errorf("Error generating challenge for client %v: %v\n", clientAddr, err)
		return
	}
	challenge := raw.String()

	log.Debugf("Generated challenge for client %v: %s", clientAddr, challenge)

	fingerprint, err := raw.Fingerprint()
	if err != nil {
		log.Infof("Error generating fingerprint for client %v: %v\n", clientAddr, err)
		return
	}

	err = storeFingerprint(fingerprint, challenge, 60*time.Second)
	if err != nil {
		log.Errorf("Failed to store fingerprint in Redis: %v", err)
		_, _ = serverConn.WriteToUDP([]byte("X-Err: internal\n"), clientAddr)
		return
	}
	log.Debugf("Stored fingerprint in Redis: %s (%s)\n", fingerprint, challenge)

	_, err = serverConn.WriteToUDP([]byte("X-Challenge: "+challenge+"\n"), clientAddr)
	if err != nil {
		log.Errorf("Error sending raw to client %v: %v\n", clientAddr, err)
		return
	}
}

func storeFingerprint(fingerprint string, challenge string, expiration time.Duration) error {
	err := redisClient.Set(ctx, "pow:challenge:"+fingerprint, challenge, expiration).Err()
	return err
}
