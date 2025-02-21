package hashcash

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	redisClient *redis.Client
	context     context.Context
	addr        string
}

func NewRedisCache(redisAddress string) *RedisCache {
	return &RedisCache{
		addr: redisAddress,
	}
}

func (r *RedisCache) Start(ctx context.Context) error {
	r.redisClient = redis.NewClient(&redis.Options{
		Addr: r.addr,
	})
	r.context = ctx
	return nil
}

func (r *RedisCache) Stop(_ context.Context) error {
	return r.redisClient.Close()
}

func (r *RedisCache) Add(fingerprint string, challenge string, expiration time.Duration) error {
	err := r.redisClient.Set(r.context, fingerprint, challenge, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store fingerprint: %w", err)
	}
	return nil
}

func (r *RedisCache) Remove(fingerprint string) error {
	exists, err := r.Exists("pow:challenge:" + fingerprint)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("fingerprint not found in cache")
	}
	_, err = r.redisClient.Del(r.context, "pow:challenge:"+fingerprint).Result()
	if err != nil {
		return fmt.Errorf("failed to remove fingerprint: %w", err)
	}

	return err
}

func (r *RedisCache) Exists(fingerprint string) (bool, error) {
	exists, err := r.redisClient.Exists(r.context, fingerprint).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check fingerprint existence: %w", err)
	}
	return exists > 0, nil
}

func (r *RedisCache) Retrieve(fingerprint string) (string, error) {
	value, err := r.redisClient.Get(r.context, fingerprint).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve challenge: %w", err)
	}
	return value, nil
}
