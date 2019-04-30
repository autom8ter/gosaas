package cache

import (
	"github.com/go-redis/redis"
	"time"
)

var db *redis.Client

func Init(network, addr, password string) {
	if db == nil {
		db = redis.NewClient(&redis.Options{
			Network:  network,
			Addr:     addr,
			Password: password,
		})
	}
}

func Get(key string) (string, error) {
	return db.Get(key).Result()
}

func Do(args ...string) (interface{}, error) {
	return db.Do(args).Result()
}

func Set(key string, val interface{}, expiration time.Duration) error {
	_, err := db.Set(key, val, expiration).Result()
	return err
}

func SetAll(values map[string]interface{}, expiration time.Duration) error {
	for k, v := range values {
		_, err := db.Set(k, v, expiration).Result()
		if err != nil {
			return err
		}
	}
	return nil
}
