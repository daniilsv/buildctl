package auth

import (
	"time"

	"github.com/build-assistant/back/pkg/cache"
)

type TokenInfo struct {
	Subject   string
	Email     string
	ExpiresAt time.Time
}

type TokenCache struct {
	cache cache.Cache[TokenInfo]
}

func NewTokenCache(ttl time.Duration) *TokenCache {
	return &TokenCache{
		cache: cache.NewInMemoryCache[TokenInfo](ttl),
	}
}

func (tc *TokenCache) Get(token string) (*TokenInfo, bool) {
	return tc.cache.Get(token)
}

func (tc *TokenCache) Set(token string, info *TokenInfo) {
	tc.cache.Set(token, info)
}

func (tc *TokenCache) Delete(token string) {
	tc.cache.Delete(token)
}

