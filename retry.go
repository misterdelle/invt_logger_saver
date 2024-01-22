package main

import (
	"fmt"
	"log"
	"math"
	"time"
)

// RetryFunc represents a function to be retried.
type RetryFunc func(args ...interface{}) (interface{}, error)

// RetryWithBackoff performs a function with retries, baseDelay, and a max number of attempts.
func RetryWithBackoff(fn RetryFunc, maxRetries int, baseDelay time.Duration, args ...interface{}) (interface{}, error) {
	var x interface{}
	var err error

	for i := 0; i < maxRetries; i++ {
		x, err = fn(args...)
		if err == nil {
			return x, nil
		}

		secRetry := math.Pow(2, float64(i))
		delay := time.Duration(secRetry) * baseDelay
		log.Printf("Attempt %d failed, retrying in %v: %v", i+1, delay, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("after %d attempts, last error: %s", maxRetries, err)
}
