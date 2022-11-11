Response time quantiles by telegram commands:
```histogram_quantile(0.5, sum(rate(bot_http_histogram_response_time_milliseconds_bucket[10s])) by (le, cmd))```
```histogram_quantile(0.9, sum(rate(bot_http_histogram_response_time_milliseconds_bucket[10s])) by (le, cmd))```
```histogram_quantile(0.99, sum(rate(bot_http_histogram_response_time_milliseconds_bucket[10s])) by (le, cmd))```

HTTP requests per second:
```rate(bot_http_histogram_response_time_seconds_count[10s])```

Currency rates by abbreviation:
```sum(gauge_currency_rate_value) by (le, abbr)```

Avg price by category "Auto" events:
```bot_event_histogram_summary_event_category_price_sum / bot_event_histogram_summary_event_category_price_count{category="Auto"}```

Effective use cache for reports:
```sum(cache_total{key="report",from_cache="1"}) * 100 / sum(cache_total{key="report"})```

Statistic requests report by period:
```sum(report_total{days="365"}) * 100 / sum(report_total)```
```sum(report_total{days="30"}) * 100 / sum(report_total)```
```sum(report_total{days="7"}) * 100 / sum(report_total)```

Response time quantiles by report worker:
```histogram_quantile(0.5, sum(rate(report_http_histogram_response_time_milliseconds_bucket[10s])) by (le, cmd))```
```histogram_quantile(0.9, sum(rate(report_http_histogram_response_time_milliseconds_bucket[10s])) by (le, cmd))```
```histogram_quantile(0.99, sum(rate(report_http_histogram_response_time_milliseconds_bucket[10s])) by (le, cmd))```

Statistic grpc requests by method name:
```sum(bot_grpc_total{method="/api.Spending/SendReport"}) * 100 / sum(bot_grpc_total)```
