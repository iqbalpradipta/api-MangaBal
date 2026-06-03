package config

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"scrapingmanga/backend/utils"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Enabled        bool
	Addr           string
	Password       string
	DB             int
	PublicCacheTTL time.Duration
}

func LoadRedisConfig() RedisConfig {
	db, _ := strconv.Atoi(utils.GetEnv("REDIS_DB", "0"))
	ttlSeconds, _ := strconv.Atoi(utils.GetEnv("REDIS_PUBLIC_CACHE_TTL_SECONDS", "60"))
	if ttlSeconds < 1 {
		ttlSeconds = 60
	}

	return RedisConfig{
		Enabled:        strings.EqualFold(utils.GetEnv("REDIS_ENABLED", "false"), "true"),
		Addr:           utils.GetEnv("REDIS_ADDR", "127.0.0.1:6379"),
		Password:       utils.GetEnv("REDIS_PASSWORD", ""),
		DB:             db,
		PublicCacheTTL: time.Duration(ttlSeconds) * time.Second,
	}
}

func InitRedis(ctx context.Context, cfg RedisConfig) *redis.Client {
	if !cfg.Enabled {
		log.Println("redis disabled")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Printf("redis unavailable, continuing without cache: %v", err)
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("failed to close redis client: %v", closeErr)
		}
		return nil
	}

	log.Println("redis connected successfully")
	return client
}
