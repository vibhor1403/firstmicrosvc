package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// StringService provides operations on strings.
type StringService interface {
	Uppercase(context.Context, string) (string, error)
	Count(context.Context, string) int
}

type myService struct{}

func (myService) Uppercase(ctx context.Context, input string) (string, error) {
	fmt.Println("In service.Uppercase")
	if input == "" {
		return "", errors.New("Empty string")
	}
	return strings.ToUpper(input), nil
}

func (myService) Count(ctx context.Context, input string) int {
	fmt.Println("In service.Count")
	return len(input)
}

// ServiceMiddleware is a chainable behavior modifier for StringService.
type ServiceMiddleware func(StringService) StringService
