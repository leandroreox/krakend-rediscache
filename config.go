package httpcache

import (
	"errors"
	"fmt"
	"time"

	"github.com/luraproject/lura/config"
)

const (
	Namespace = "github.com/devopsfaith/krakend-httpcache"

	BackendMemory = "memory"
	BackendRedis  = "redis"

	RedisModeRedis                 = "redis"
	RedisModeCluster               = "rediscluster"
	RedisDefaultDialTimeout        = 100 * time.Millisecond
	RedisDefaultReadTimeout        = 100 * time.Millisecond
	RedisDefaultWriteTimeout       = 200 * time.Millisecond
	RedisDefaultMaxRetries         = 0
	RedisDefaultIdleTimeout        = 5 * time.Minute
	RedisDefaultIdleCheckFrequency = 1 * time.Minute
	RedisDefaultPoolSize           = 10
	RedisDefaultPoolTimeout        = 10 * time.Millisecond
	RedisDefaultTtl                = 1 * time.Hour
)

var WarnNoConfig = errors.New("no config found for krakend-httpcache")
var ErrMappingConfig = errors.New("could not map krakend-httpcache config")
var ErrMissingRequired = func(field string) error {
	return errors.New(fmt.Sprintf("Missing required krakend-httpcache config field [%s]", field))
}
var ErrInvalidValue = func(field string) error {
	return errors.New(fmt.Sprintf("Invalid value for krakend-httpcache config field [%s]", field))
}

type RedisConfig struct {
	Mode               string
	Address            string
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	MaxRetries         int
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	PoolSize           int
	PoolTimeout        time.Duration
	Ttl                time.Duration
}

type Config struct {
	Type        string
	RedisConfig RedisConfig
}

func ConfigGetter(cfg *config.Backend) (interface{}, error) {
	value, ok := cfg.ExtraConfig[Namespace]
	if !ok {
		return nil, WarnNoConfig
	}

	castedConfig, ok := value.(map[string]interface{})
	if !ok {
		return nil, ErrMappingConfig
	}

	c := Config{}
	if value, ok := castedConfig["type"]; ok {
		c.Type = value.(string)
	} else {
		c.Type = BackendMemory
	}

	redisreq := c.Type == BackendRedis
	if rediscfg, ok := castedConfig[BackendRedis].(map[string]interface{}); ok {
		if mode, ok := rediscfg["mode"]; ok {
			switch mode {
			case RedisModeRedis, RedisModeCluster:
				c.RedisConfig.Mode = mode.(string)
			default:
				return nil, ErrInvalidValue("mode")
			}
		} else {
			return nil, ErrMissingRequired("mode")
		}
		addr, err := getRequiredFromConfig(rediscfg, "address")
		if err != nil {
			return nil, err
		}
		c.RedisConfig.Address = addr.(string)
		c.RedisConfig.DialTimeout = getDurationFromConfigWithDefaultValue(rediscfg, "dialTimeout", RedisDefaultDialTimeout)
		c.RedisConfig.ReadTimeout = getDurationFromConfigWithDefaultValue(rediscfg, "readTimeout", RedisDefaultReadTimeout)
		c.RedisConfig.WriteTimeout = getDurationFromConfigWithDefaultValue(rediscfg, "writeTimeout", RedisDefaultWriteTimeout)
		c.RedisConfig.MaxRetries = getIntFromConfigWithDefaultValue(rediscfg, "maxRetries", RedisDefaultMaxRetries)
		c.RedisConfig.IdleTimeout = getDurationFromConfigWithDefaultValue(rediscfg, "idleTimeout", RedisDefaultIdleTimeout)
		c.RedisConfig.IdleCheckFrequency = getDurationFromConfigWithDefaultValue(rediscfg, "idleCheckFrequency", RedisDefaultIdleCheckFrequency)
		c.RedisConfig.PoolSize = getIntFromConfigWithDefaultValue(rediscfg, "poolSize", RedisDefaultPoolSize)
		c.RedisConfig.PoolTimeout = getDurationFromConfigWithDefaultValue(rediscfg, "poolTimeout", RedisDefaultPoolTimeout)
		c.RedisConfig.Ttl = getDurationFromConfigWithDefaultValue(rediscfg, "ttl", RedisDefaultTtl)
	} else if redisreq {
		return nil, ErrMissingRequired("redis")
	}
	return c, nil
}

func getDurationFromConfigWithDefaultValue(cfg map[string]interface{}, k string, d time.Duration) time.Duration {
	val, def := getFromConfigWithDefaultValue(cfg, k, d)
	if def {
		return val.(time.Duration)
	}
	nval, _ := time.ParseDuration(val.(string))
	return nval
}

func getIntFromConfigWithDefaultValue(cfg map[string]interface{}, k string, d int) int {
	val, _ := getFromConfigWithDefaultValue(cfg, k, d)
	return val.(int)
}

func getFromConfigWithDefaultValue(cfg map[string]interface{}, k string, d interface{}) (val interface{}, defRet bool) {
	v, err := getRequiredFromConfig(cfg, k)
	if err != nil {
		val = d
		defRet = true
	} else {
		val = v
	}
	return
}

func getRequiredFromConfig(cfg map[string]interface{}, k string) (interface{}, error) {
	if val, ok := cfg[k]; ok {
		return val, nil
	}
	return nil, ErrMissingRequired(k)
}
