package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const publicCachePrefix = "manga-api:public:"

type CacheService interface {
	Configured() bool
	Enabled() bool
	PublicKey(value string) string
	GetBytes(ctx context.Context, key string) ([]byte, bool)
	SetBytes(ctx context.Context, key string, value []byte)
	ClearPublic(ctx context.Context)
	Ping(ctx context.Context) error
}

type redisCacheService struct {
	client     *redis.Client
	ttl        time.Duration
	configured bool
}

func NewRedisCacheService(client *redis.Client, ttl time.Duration, configured bool) CacheService {
	return &redisCacheService{
		client:     client,
		ttl:        ttl,
		configured: configured,
	}
}

func (s *redisCacheService) Configured() bool {
	return s != nil && s.configured
}

func (s *redisCacheService) Enabled() bool {
	return s != nil && s.client != nil
}

func (s *redisCacheService) PublicKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return publicCachePrefix + hex.EncodeToString(sum[:])
}

func (s *redisCacheService) GetBytes(ctx context.Context, key string) ([]byte, bool) {
	if !s.Enabled() {
		return nil, false
	}

	value, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		log.Printf("redis cache get failed: %v", err)
		return nil, false
	}
	return value, true
}

func (s *redisCacheService) SetBytes(ctx context.Context, key string, value []byte) {
	if !s.Enabled() || len(value) == 0 {
		return
	}
	if err := s.client.Set(ctx, key, value, s.ttl).Err(); err != nil {
		log.Printf("redis cache set failed: %v", err)
	}
}

func (s *redisCacheService) ClearPublic(ctx context.Context) {
	if !s.Enabled() {
		return
	}

	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, publicCachePrefix+"*", 100).Result()
		if err != nil {
			log.Printf("redis public cache scan failed: %v", err)
			return
		}
		if len(keys) > 0 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				log.Printf("redis public cache delete failed: %v", err)
				return
			}
		}
		if nextCursor == 0 {
			return
		}
		cursor = nextCursor
	}
}

func (s *redisCacheService) Ping(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	return s.client.Ping(ctx).Err()
}
