package lib

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

type CacheCreds struct {
	Addr     string
	Password string
	Db       int
}

type ICacheManager interface {
	ExecuteWithRedisConnection(requestId string, connectionExecutor func(*redis.Client) error, connectionRetry int) error
}

func NewRedisManager(RedisCreds *CacheCreds) *RedisManager {
	return &RedisManager{
		RedisCreds: RedisCreds,
	}
}

type RedisManager struct {
	RedisCreds *CacheCreds
}

func (r *RedisManager) ExecuteWithRedisConnection(requestId string, connectionExecutor func(*redis.Client) error, connectionRetry int) error {

	redisConnection := redis.NewClient(&redis.Options{
		Addr:     r.RedisCreds.Addr,
		Password: r.RedisCreds.Password,
		DB:       r.RedisCreds.Db,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		},
	})

	err := connectionExecutor(redisConnection)
	defer redisConnection.Close()
	if err != nil {
		if connectionRetry > 5 {
			log.Printf("%s: Error: Connection retries more than %d times", requestId, connectionRetry)
			CaptureSentryException(fmt.Sprintf("Redis tried to reconnect more than %d times", connectionRetry))
		} else {
			log.Printf("%s: Debug: Retry count %d", requestId, connectionRetry)
			r.ExecuteWithRedisConnection(requestId, connectionExecutor, connectionRetry+1)
		}
	}
	return err
}
