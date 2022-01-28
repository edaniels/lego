package wait

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/edaniels/golog"
	"go.uber.org/multierr"
)

// For polls the given function 'f', once every 'interval', up to 'timeout'.
func For(ctx context.Context, msg string, timeout, interval time.Duration, f func() (bool, error), logger golog.Logger) error {
	logger.Infof("Wait for %s [timeout: %s, interval: %s]", msg, timeout, interval)

	var lastErr error
	timeUp := time.After(timeout)
	for {
		if ctx.Err() != nil {
			return multierr.Combine(lastErr, ctx.Err())
		}
		select {
		case <-timeUp:
			if lastErr == nil {
				return errors.New("time limit exceeded")
			}
			return fmt.Errorf("time limit exceeded: last error: %w", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			lastErr = err
		}

		time.Sleep(interval)
	}
}
