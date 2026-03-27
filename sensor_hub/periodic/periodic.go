package periodic

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"time"
)

// TaskConfig configures a supervised periodic task.
type TaskConfig struct {
	Name           string
	Interval       time.Duration
	Logger         *slog.Logger
	RunImmediately bool // if true, run the task once before waiting for the first tick
}

const (
	initialBackoff = 5 * time.Second
	maxBackoff     = 5 * time.Minute
)

// RunTask launches a supervised goroutine that executes task on every tick.
// On panic the goroutine logs the stack trace, backs off exponentially, and restarts.
// The goroutine exits cleanly when ctx is cancelled.
func RunTask(ctx context.Context, cfg TaskConfig, task func(ctx context.Context) error) {
	go func() {
		consecutivePanics := 0
		for {
			stopped := runLoop(ctx, cfg, task, &consecutivePanics)
			if stopped {
				return
			}

			// We only reach here after a panic recovery. Back off before restarting.
			backoff := time.Duration(float64(initialBackoff) * math.Pow(2, float64(consecutivePanics-1)))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			cfg.Logger.Error("restarting periodic task after backoff",
				"task", cfg.Name,
				"backoff", backoff.String(),
				"consecutive_panics", consecutivePanics,
			)

			select {
			case <-ctx.Done():
				cfg.Logger.Info("periodic task stopping during backoff", "task", cfg.Name, "reason", ctx.Err())
				return
			case <-time.After(backoff):
			}
		}
	}()
}

// runLoop runs the tick loop. It returns true if ctx was cancelled (clean
// shutdown) or false if the loop exited due to a recovered panic.
func runLoop(ctx context.Context, cfg TaskConfig, task func(ctx context.Context) error, consecutivePanics *int) (stopped bool) {
	defer func() {
		if r := recover(); r != nil {
			*consecutivePanics++
			cfg.Logger.Error("periodic task panicked",
				"task", cfg.Name,
				"panic", fmt.Sprintf("%v", r),
				"stack", string(debug.Stack()),
				"consecutive_panics", *consecutivePanics,
			)
			stopped = false
		}
	}()

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	cfg.Logger.Info("periodic task started", "task", cfg.Name, "interval", cfg.Interval.String())

	if cfg.RunImmediately {
		executeTask(ctx, cfg, task, consecutivePanics)
	}

	for {
		select {
		case <-ctx.Done():
			cfg.Logger.Info("periodic task stopping", "task", cfg.Name, "reason", ctx.Err())
			return true
		case <-ticker.C:
			executeTask(ctx, cfg, task, consecutivePanics)
		}
	}
}

func executeTask(ctx context.Context, cfg TaskConfig, task func(ctx context.Context) error, consecutivePanics *int) {
	err := task(ctx)
	if err != nil {
		cfg.Logger.Error("periodic task error", "task", cfg.Name, "error", err)
	} else {
		*consecutivePanics = 0
		cfg.Logger.Info("periodic task completed", "task", cfg.Name, "next_in", cfg.Interval.String())
	}
}
