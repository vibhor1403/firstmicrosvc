package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type proxyMW struct {
	next      StringService     // all services other than the one defined below will continue to be served normally.
	uppercase endpoint.Endpoint // uppercase will be served through this
}

func (mw proxyMW) Uppercase(ctx context.Context, s string) (output string, err error) {
	response, err := mw.uppercase(ctx, uppercaseRequest{S: s})
	if err != nil {
		return "", err
	}
	resp := response.(uppercaseResponse)
	if resp.Err != "" {
		return resp.V, errors.New(resp.Err)
	}
	return resp.V, nil
}

func (mw proxyMW) Count(ctx context.Context, s string) (n int) {
	return mw.next.Count(ctx, s)
}

func proxyingMiddleware(instances string, logger log.Logger) ServiceMiddleware {
	fmt.Println("in prox middlwr", instances)
	if instances == "" {
		logger.Log("proxy_to", "none")
		return func(next StringService) StringService { return next }
	}

	var (
		qps         = 100
		maxAttempts = 3
		maxTime     = 250 * time.Millisecond
	)

	// create endpoint for each proxy
	var (
		instanceList = split(instances)
		endpointer   sd.FixedEndpointer
	)
	for _, instance := range instanceList {
		var e endpoint.Endpoint
		e = makeProxyUppercaseEndpoint(instance)
		e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
		e = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(e)
		endpointer = append(endpointer, e)
	}

	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(maxAttempts, maxTime, balancer)

	return func(next StringService) StringService {
		return proxyMW{next, retry}
	}

}
func split(s string) []string {
	a := strings.Split(s, ",")
	for i := range a {
		a[i] = strings.TrimSpace(a[i])
	}
	return a
}
func makeProxyUppercaseEndpoint(proxyURL string) endpoint.Endpoint {
	if !strings.HasPrefix(proxyURL, "http") {
		proxyURL = "http://" + proxyURL
	}
	u, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}
	if u.Path == "" {
		u.Path = "/uppercase"
	}
	return httptransport.NewClient(
		"GET",
		u,
		encodeRequest,
		decodeUppercaseResponse).Endpoint()
}
