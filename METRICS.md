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