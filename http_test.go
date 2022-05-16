package httpcache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/go-redis/redis/v8"
	httpcache2 "github.com/gregjones/httpcache"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/encoding"
	"github.com/luraproject/lura/proxy"
	"github.com/stretchr/testify/assert"
)

func TestClient_ok(t *testing.T) {
	testCacheSystem(t, func(t *testing.T, URL string) {
		testClient(t, sampleInMemoryCfg, URL)
	}, 1)
}

func TestClient_ko(t *testing.T) {
	cfg := &config.Backend{
		Decoder:     encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{},
	}
	testCacheSystem(t, func(t *testing.T, URL string) {
		testClient(t, cfg, URL)
	}, 100)
}

func testClient(t *testing.T, cfg *config.Backend, URL string) {
	clientFactory := NewHTTPClient(cfg)
	client := clientFactory(context.Background())

	for i := 0; i < 100; i++ {
		resp, err := client.Get(URL)
		if err != nil {
			log.Println(err)
			t.Error(err)
			return
		}
		response, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Error(err)
			return
		}
		if string(response) != statusOKMsg {
			t.Error("unexpected body:", string(response))
		}
	}
}

func TestCreateRedisCacheClient(t *testing.T) {
	clientFactory := NewHTTPClient(sampleRedisCfg)
	client := clientFactory(context.Background())

	cache := client.Transport.(*httpcache2.Transport).Cache
	assert.IsType(t, &RedisCache{}, cache)
	rc := cache.(*RedisCache)

	assert.IsType(t, &redis.Client{}, rc.client)
}

func TestCreateRedisClusterCacheClient(t *testing.T) {
	clientFactory := NewHTTPClient(sampleRedisClusterCfg)
	client := clientFactory(context.Background())

	cache := client.Transport.(*httpcache2.Transport).Cache
	assert.IsType(t, &RedisCache{}, cache)
	rc := cache.(*RedisCache)

	assert.IsType(t, &redis.ClusterClient{}, rc.client)
}

func TestBackendFactory(t *testing.T) {
	testCacheSystem(t, func(t *testing.T, testURL string) {
		backendFactory := BackendFactory(sampleInMemoryCfg)
		backendProxy := backendFactory(sampleInMemoryCfg)
		ctx := context.Background()
		URL, _ := url.Parse(testURL)

		for i := 0; i < 100; i++ {
			req := &proxy.Request{
				Method: "GET",
				URL:    URL,
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
			}
			resp, err := backendProxy(ctx, req)
			if err != nil {
				t.Error(err)
				return
			}
			if !resp.IsComplete {
				t.Error("incomplete response:", *resp)
			}
		}
	}, 1)
}

var (
	statusOKMsg       = `{"status": "ok"}`
	sampleInMemoryCfg = &config.Backend{
		Decoder: encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{},
		},
	}
	sampleRedisCfg = &config.Backend{
		Decoder: encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"type": "redis",
				"redis": map[string]interface{}{
					"address": "ip:port",
					"mode":    "redis",
				},
			},
		},
	}
	sampleRedisClusterCfg = &config.Backend{
		Decoder: encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"type": "redis",
				"redis": map[string]interface{}{
					"address": "ip2:port2",
					"mode":    "rediscluster",
				},
			},
		},
	}
)

func testCacheSystem(t *testing.T, f func(*testing.T, string), expected uint64) {
	var ops uint64 = 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&ops, 1)
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, statusOKMsg)
	}))
	defer testServer.Close()

	f(t, testServer.URL)

	opsFinal := atomic.LoadUint64(&ops)
	if opsFinal != expected {
		t.Errorf("the server should not being hited just %d time(s). Total requests: %d\n", expected, opsFinal)
	}
}
