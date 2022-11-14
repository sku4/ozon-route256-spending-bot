package cache

import (
	"context"
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/cache/lru"
	"gitlab.ozon.dev/skubach/workshop-1-bot/pkg/logger"
	"os"
	"time"
)

var (
	cache          *redisCache.Cache
	lruCache       *lru.LRU
	lruChan        chan lru.Item
	onceItemChan   chan OnceItem
	onceResultChan chan OnceResult
	size           = 1000
	limitCacheTime = 500 * time.Millisecond
	ttl            = time.Minute
	cacheTotal     = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_total",
		},
		[]string{"key", "from_cache"},
	)
)

type Item redisCache.Item
type Do func(*Item) (interface{}, error)

type OnceItem struct {
	item       *Item
	do         Do
	metricName string
}

type OnceResult struct {
	fromCache string
	err       error
}

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Fatalf("error loading env variables: %s", err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         ":" + os.Getenv("REDIS_PORT"),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	cache = redisCache.New(&redisCache.Options{
		Redis:      rdb,
		LocalCache: redisCache.NewTinyLFU(size, ttl),
	})

	lruCache = lru.NewLRU(size)
	lruChan = make(chan lru.Item, size)
	onceItemChan = make(chan OnceItem, 1)
	onceResultChan = make(chan OnceResult, 1)
}

func Once(item *Item, do Do, metricName string) (err error) {
	onceItemChan <- OnceItem{
		item:       item,
		do:         do,
		metricName: metricName,
	}

	fromCache := "1"
	fromLruCache := false
	lruValue := lruCache.Get(item.Key)
	if lruValue == nil {
		onceResult := <-onceResultChan
		if onceResult.err != nil {
			return onceResult.err
		}
		fromCache = onceResult.fromCache
	} else {
		select {
		case onceResult := <-onceResultChan:
			if onceResult.err != nil {
				return onceResult.err
			}
			fromCache = onceResult.fromCache
		case <-time.After(limitCacheTime):
			item.Value = lruValue
			fromLruCache = true
		}
	}

	cacheTotal.WithLabelValues(metricName, fromCache).Inc()

	if !fromLruCache {
		lruChan <- lru.Item{
			Key:   item.Key,
			Value: item.Value,
		}
	}

	return
}

func Run(ctx context.Context) {
	go addLruCache(ctx)
	go onceWorker(ctx)
}

func addLruCache(ctx context.Context) {
	for {
		select {
		case lruItem := <-lruChan:
			lruCache.Add(lruItem.Key, lruItem.Value)
		case <-ctx.Done():
			return
		}
	}
}

func onceWorker(ctx context.Context) {
	for {
		select {
		case onceItem := <-onceItemChan:
			fromCache := "1"
			onceItem.item.Do = func(*redisCache.Item) (interface{}, error) {
				fromCache = "0"
				return onceItem.do(onceItem.item)
			}
			err := cache.Once((*redisCache.Item)(onceItem.item))
			onceResultChan <- OnceResult{
				fromCache: fromCache,
				err:       err,
			}
		case <-ctx.Done():
			return
		}
	}
}
