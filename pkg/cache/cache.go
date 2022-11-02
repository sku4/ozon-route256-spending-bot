package cache

import (
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"os"
	"time"
)

var (
	cache *redisCache.Cache
	size  = 1000
	ttl   = time.Minute
)

type Item redisCache.Item
type Do func(*Item) (interface{}, error)

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Fatalf("error loading env variables: %s", err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: ":" + os.Getenv("REDIS_PORT"),
	})

	cache = redisCache.New(&redisCache.Options{
		Redis:      rdb,
		LocalCache: redisCache.NewTinyLFU(size, ttl),
	})
}

func Once(item *Item, do Do) (err error) {
	item.Do = func(*redisCache.Item) (interface{}, error) {
		return do(item)
	}
	return cache.Once((*redisCache.Item)(item))
}
