package cache

import (
	"context"
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sku4/ozon-route256-spending-bot/pkg/cache/lru"
	"github.com/sku4/ozon-route256-spending-bot/pkg/logger"
	"os"
	"time"
)

var (
	cache          *redisCache.Cache
	lruCache       *lru.LRU
	onceItemChan   chan OnceItem
	size           = 1000
	bufferSize     = 10_000
	workerCount    = 5
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
	resultCh   chan OnceResult
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
	onceItemChan = make(chan OnceItem, bufferSize)
}

func Once(item *Item, do Do, metricName string) (err error) {
	onceResultChan := make(chan OnceResult)
	onceItemChan <- OnceItem{
		item:       item,
		do:         do,
		metricName: metricName,
		resultCh:   onceResultChan,
	}
	t := time.NewTimer(limitCacheTime)

	fromCache := "1"
	select {
	case onceResult := <-onceResultChan:
		t.Stop()
		if onceResult.err != nil {
			return onceResult.err
		}
		fromCache = onceResult.fromCache
	case <-t.C:
		if lruValue, ok := lruCache.Get(item.Key); ok {
			e := cache.Unmarshal(lruValue.([]byte), item.Value)
			if e != nil {
				return e
			}
		} else {
			onceResult := <-onceResultChan
			if onceResult.err != nil {
				return onceResult.err
			}
			fromCache = onceResult.fromCache
		}
	}

	cacheTotal.WithLabelValues(metricName, fromCache).Inc()

	go func(item *Item) {
		b, _ := cache.Marshal(item.Value)
		lruCache.Add(item.Key, b)
	}(item)

	return
}

func Run(ctx context.Context) {
	for w := 0; w < workerCount; w++ {
		go onceWorker(ctx)
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
			onceItem.resultCh <- OnceResult{
				fromCache: fromCache,
				err:       err,
			}
		case <-ctx.Done():
			return
		}
	}
}
