package httpcache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	httpcache2 "github.com/gregjones/httpcache"
	"github.com/stretchr/testify/assert"

	httpcache "github.com/devopsfaith/krakend-httpcache"
)

func TestRedisCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := httpcache.NewMockClient(ctrl)
	ttl := 10 * time.Millisecond

	c := httpcache.NewRedisCache(client, ttl)

	t.Run("Calls client to set new value", func(t *testing.T) {
		k := "setkey"
		v := []byte("aresponse")

		client.EXPECT().Set(gomock.Any(), k, v, ttl).Times(1)

		c.Set(k, v)
	})

	t.Run("Calls client to delete existing value", func(t *testing.T) {
		k := "delkey"

		client.EXPECT().Del(gomock.Any(), k)

		c.Delete(k)
	})

	t.Run("Get returns ko when cant find in cache", func(t *testing.T) {
		k := "getko"

		res := redis.NewStringResult("", errors.New(""))
		client.EXPECT().Get(gomock.Any(), k).Times(1).Return(res)

		_, ok := c.Get(k)

		assert.False(t, ok)
	})

	t.Run("Get returns value when found in cache", func(t *testing.T) {
		k := "getok"
		v := "foundincache"
		res := redis.NewStringResult(v, nil)
		client.EXPECT().Get(gomock.Any(), k).Times(1).Return(res)

		r, ok := c.Get(k)

		assert.True(t, ok)
		assert.Equal(t, v, string(r))
	})
}

func TestNewRedis(t *testing.T) {
	rc := httpcache.NewRedis(httpcache.RedisConfig{})

	assert.IsType(t, &redis.Client{}, rc)
}

func TestNewRedisCluster(t *testing.T) {
	rc := httpcache.NewRedisCluster(httpcache.RedisConfig{})

	assert.IsType(t, &redis.ClusterClient{}, rc)
}

func TestRedisCacheTransport_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rc := httpcache.NewMockCache(ctrl)

	rct := httpcache.NewRedisCacheTransport(rc)

	assert.IsType(t, &httpcache2.Transport{}, rct)
}
