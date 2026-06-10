package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

type noOpCacheService struct{}

func NewNoOpCacheService() CacheService {
	return &noOpCacheService{}
}

func (s *noOpCacheService) Configured() bool {
	return false
}

func (s *noOpCacheService) Enabled() bool {
	return false
}

func (s *noOpCacheService) PublicKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return publicCachePrefix + hex.EncodeToString(sum[:])
}

func (s *noOpCacheService) GetBytes(ctx context.Context, key string) ([]byte, bool) {
	return nil, false
}

func (s *noOpCacheService) SetBytes(ctx context.Context, key string, value []byte) {
	// no-op
}

func (s *noOpCacheService) ClearPublic(ctx context.Context) {
	// no-op
}

func (s *noOpCacheService) Ping(ctx context.Context) error {
	return nil
}
