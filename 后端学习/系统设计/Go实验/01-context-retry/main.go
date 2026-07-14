package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var errTransient = errors.New("transient dependency error")

func retry(
	ctx context.Context,
	maxAttempts int,
	baseDelay time.Duration,
	maxDelay time.Duration,
	rng *rand.Rand,
	fn func(context.Context) error,
) error {
	if maxAttempts <= 0 {
		return errors.New("maxAttempts must be positive")
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("before attempt %d: %w", attempt, err)
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if !errors.Is(err, errTransient) || attempt == maxAttempts {
			break
		}

		delay := baseDelay << (attempt - 1)
		if delay > maxDelay {
			delay = maxDelay
		}
		jitterLimit := delay / 2
		jitter := time.Duration(0)
		if jitterLimit > 0 {
			jitter = time.Duration(rng.Int63n(int64(jitterLimit) + 1))
		}

		timer := time.NewTimer(delay + jitter)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return fmt.Errorf("waiting before attempt %d: %w", attempt+1, ctx.Err())
		}
	}

	return fmt.Errorf("retry exhausted: %w", lastErr)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	attempts := 0
	rng := rand.New(rand.NewSource(7))
	err := retry(ctx, 5, 10*time.Millisecond, 80*time.Millisecond, rng, func(context.Context) error {
		attempts++
		fmt.Printf("attempt=%d\n", attempts)
		if attempts < 3 {
			return errTransient
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	if attempts != 3 {
		panic(fmt.Sprintf("want 3 attempts, got %d", attempts))
	}
	fmt.Println("success with bounded retry")
}
