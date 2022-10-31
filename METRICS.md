Response time quantiles by telegram commands:
```histogram_quantile(0.99, sum(rate(bot_http_histogram_response_time_seconds_bucket[10s])) by (le, cmd))```

HTTP requests per second:
```rate(bot_http_histogram_response_time_seconds_count[10s])```

Currency rates by abbreviation:
```sum(gauge_currency_rate_value) by (le, abbr)```

Avg price by all events:
```bot_event_histogram_summary_event_price_sum / bot_event_histogram_summary_event_price_count```