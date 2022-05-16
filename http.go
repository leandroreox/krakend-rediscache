// Package httpcache introduces an in-memory-cached http client into the KrakenD stack
package httpcache

import (
	"context"
	"net/http"

	"github.com/gregjones/httpcache"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/proxy"
	"github.com/luraproject/lura/transport/http/client"
)

var (
	memTransport = httpcache.NewMemoryCacheTransport()
	memClient    = http.Client{Transport: memTransport}
)

// NewHTTPClient creates a HTTPClientFactory using an in-memory-cached http client
func NewHTTPClient(cfg *config.Backend) client.HTTPClientFactory {
	c, err := ConfigGetter(cfg)
	if err == nil {
		cacheCfg := c.(Config)

		switch cacheCfg.Type {
		case BackendMemory:
			return func(_ context.Context) *http.Client {
				return &memClient
			}
		case BackendRedis:
			var r Client
			switch cacheCfg.RedisConfig.Mode {
			case RedisModeRedis:
				r = NewRedis(cacheCfg.RedisConfig)
			case RedisModeCluster:
				r = NewRedisCluster(cacheCfg.RedisConfig)
			}
			return func(_ context.Context) *http.Client {
				return &http.Client{Transport: NewRedisCacheTransport(NewRedisCache(r, cacheCfg.RedisConfig.Ttl))}
			}
		}
	}
	return client.NewHTTPClient
}

// BackendFactory returns a proxy.BackendFactory that creates backend proxies using
// an in-memory-cached http client
func BackendFactory(cfg *config.Backend) proxy.BackendFactory {
	return proxy.CustomHTTPProxyFactory(NewHTTPClient(cfg))
}
