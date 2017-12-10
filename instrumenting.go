package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
)

func instrumentingMiddleware(
	requestCount metrics.Counter,
	requestLatency metrics.Histogram,
	countResult metrics.Histogram,
) ServiceMiddleware {
	return func(next StringService) StringService {
		return appinstrumentingMiddleware{requestCount, requestLatency, countResult, next}
	}
}

type appinstrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	countResult    metrics.Histogram
	next           StringService
}

func (mw appinstrumentingMiddleware) Uppercase(ctx context.Context, s string) (output string, err error) {
	defer func(begin time.Time) {
		params := []string{"method", "uppercase", "error", fmt.Sprint(err != nil)}

		mw.requestCount.With(params...).Add(1)
		mw.requestLatency.With(params...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	output, err = mw.next.Uppercase(ctx, s)
	return
}
func (mw appinstrumentingMiddleware) Count(ctx context.Context, s string) (n int) {
	defer func(begin time.Time) {
		params := []string{"method", "count", "error", "false"}

		mw.requestCount.With(params...).Add(1)
		mw.requestLatency.With(params...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	n = mw.next.Count(ctx, s)
	return
}
