package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Middleware takes endpoint and returns Endpoint (based on decorator pattern)
type Middleware func(endpoint.Endpoint) endpoint.Endpoint

func endPointloggingMiddleware(logger log.Logger) Middleware {
	fmt.Println("In endPointLoggingMiddleware")
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			logger.Log("msg", "calling endpoint")
			defer logger.Log("msg", "called endpoint")
			return next(ctx, request)
		}
	}
}

func loggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next StringService) StringService {
		return appLoggingMiddleware{logger, next}
	}
}

type appLoggingMiddleware struct {
	logger log.Logger
	next   StringService
}

func (mw appLoggingMiddleware) Uppercase(ctx context.Context, s string) (output string, err error) {
	fmt.Println("In loggingMiddleware.Uppercase")
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "uppercase",
			"input", s,
			"output", output,
			"error", err,
			"duration", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.next.Uppercase(ctx, s)
	return
}

func (mw appLoggingMiddleware) Count(ctx context.Context, s string) (n int) {
	fmt.Println("In loggingMiddleware.Count")
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "uppercase",
			"input", s,
			"length", n,
			"duration", time.Since(begin),
		)
	}(time.Now())
	n = mw.next.Count(ctx, s)
	return
}
