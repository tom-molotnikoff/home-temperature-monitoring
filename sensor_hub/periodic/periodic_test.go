package periodic

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testLogger() *slog.Logger {
	return slog.Default()
}

func TestRunTask_NormalExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32

	RunTask(ctx, TaskConfig{
		Name:           "test_normal",
		Interval:       50 * time.Millisecond,
		Logger:         testLogger(),
		RunImmediately: true,
	}, func(ctx context.Context) error {
		count.Add(1)
		return nil
	})

	time.Sleep(180 * time.Millisecond)
	cancel()
	time.Sleep(20 * time.Millisecond)

	got := count.Load()
	assert.GreaterOrEqual(t, got, int32(3), "expected at least 3 executions (1 immediate + 2 ticks)")
}

func TestRunTask_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var count atomic.Int32

	RunTask(ctx, TaskConfig{
		Name:           "test_cancel",
		Interval:       50 * time.Millisecond,
		Logger:         testLogger(),
		RunImmediately: true,
	}, func(ctx context.Context) error {
		count.Add(1)
		return nil
	})

	time.Sleep(30 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)

	got := count.Load()
	// Should have run once immediately then stopped
	assert.LessOrEqual(t, got, int32(2), "expected at most 2 executions after cancel")
}

func TestRunTask_PanicRecoveryAndRestart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32

	RunTask(ctx, TaskConfig{
		Name:           "test_panic",
		Interval:       50 * time.Millisecond,
		Logger:         testLogger(),
		RunImmediately: true,
	}, func(ctx context.Context) error {
		n := count.Add(1)
		if n == 1 {
			panic("simulated panic")
		}
		return nil
	})

	// First call panics immediately. After initial backoff (5s) it restarts.
	// We wait long enough for the backoff + one more execution.
	time.Sleep(6 * time.Second)
	cancel()
	time.Sleep(20 * time.Millisecond)

	got := count.Load()
	assert.GreaterOrEqual(t, got, int32(2), "expected goroutine to restart after panic")
}

func TestRunTask_ErrorDoesNotKillGoroutine(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32

	RunTask(ctx, TaskConfig{
		Name:           "test_error",
		Interval:       50 * time.Millisecond,
		Logger:         testLogger(),
		RunImmediately: true,
	}, func(ctx context.Context) error {
		count.Add(1)
		return errors.New("transient error")
	})

	time.Sleep(180 * time.Millisecond)
	cancel()

	got := count.Load()
	assert.GreaterOrEqual(t, got, int32(3), "errors should not stop the periodic loop")
}

func TestRunTask_RunImmediatelyFalse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32

	RunTask(ctx, TaskConfig{
		Name:           "test_no_immediate",
		Interval:       50 * time.Millisecond,
		Logger:         testLogger(),
		RunImmediately: false,
	}, func(ctx context.Context) error {
		count.Add(1)
		return nil
	})

	// Before first tick fires, count should still be 0
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), count.Load(), "should not have run before first tick")

	time.Sleep(80 * time.Millisecond)
	assert.GreaterOrEqual(t, count.Load(), int32(1), "should have run after first tick")
}

func TestRunTask_BackoffResetsOnSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32

	RunTask(ctx, TaskConfig{
		Name:           "test_backoff_reset",
		Interval:       50 * time.Millisecond,
		Logger:         testLogger(),
		RunImmediately: true,
	}, func(ctx context.Context) error {
		n := count.Add(1)
		if n == 1 {
			panic("first panic")
		}
		// After restart, succeed — this should reset the backoff counter
		return nil
	})

	// Wait for panic + backoff (5s) + successful restart + a few ticks
	time.Sleep(6 * time.Second)

	got := count.Load()
	assert.GreaterOrEqual(t, got, int32(3), "should have multiple successful runs after recovery")
}
