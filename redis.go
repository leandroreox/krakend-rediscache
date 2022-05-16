//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=$GOPACKAGE

package httpcache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gregjones/httpcache"
)

type Client interface {
	redis.Cmdable
}

func NewRedis(cfg RedisConfig) Client {
	return redis.NewClient(&redis.Options{
		Addr:               cfg.Address,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		MaxRetries:         cfg.MaxRetries,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,
		PoolSize:           cfg.PoolSize,
		PoolTimeout:        cfg.PoolTimeout,
	})
}

func NewRedisCluster(cfg RedisConfig) Client {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              []string{cfg.Address},
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		MaxRetries:         cfg.MaxRetries,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,
		PoolSize:           cfg.PoolSize,
		PoolTimeout:        cfg.PoolTimeout,
	})
}

type Cache interface {
	httpcache.Cache
}

type RedisCache struct {
	client Client
	ttl    time.Duration
}

func NewRedisCache(client Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisCache) Get(key string) (responseBytes []byte, ok bool) {
	r := c.client.Get(context.Background(), key)
	rb, err := r.Bytes()
	if err != nil {
		return []byte{}, false
	}
	return rb, true
}

func (c *RedisCache) Set(key string, responseBytes []byte) {
	c.client.Set(context.Background(), key, responseBytes, c.ttl)
}

func (c *RedisCache) Delete(key string) {
	c.client.Del(context.Background(), key)
}

func NewRedisCacheTransport(c Cache) *httpcache.Transport {
	t := httpcache.NewTransport(c)
	return t
}
