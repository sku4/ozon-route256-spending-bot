package cache

import (
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"os"
	"time"
)

var (
	cache      *redisCache.Cache
	size       = 1000
	ttl        = time.Minute
	cacheTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_total",
		},
		[]string{"key", "from_cache"},
	)
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

func Once(item *Item, do Do, metricName string) (err error) {
	fromCache := "1"
	item.Do = func(*redisCache.Item) (interface{}, error) {
		fromCache = "0"
		return do(item)
	}
	err = cache.Once((*redisCache.Item)(item))
	cacheTotal.WithLabelValues(metricName, fromCache).Inc()
	return
}
