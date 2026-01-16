package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// simpleBlocker tests
// ============================================================================

func TestSimpleBlocker_NewSimpleBlocker(t *testing.T) {
	blocker := newSimpleBlocker()

	assert.NotNil(t, blocker)
	assert.NotNil(t, blocker.m)
	assert.NotNil(t, blocker.allowOnce)
	assert.NotNil(t, blocker.allowCount)
}

// ============================================================================
// getRemainingSeconds tests
// ============================================================================

func TestSimpleBlocker_GetRemainingSeconds_NoBlock(t *testing.T) {
	blocker := newSimpleBlocker()

	remaining := blocker.getRemainingSeconds("testKey")

	assert.Equal(t, 0, remaining)
}

func TestSimpleBlocker_GetRemainingSeconds_Blocked(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 10)

	remaining := blocker.getRemainingSeconds("testKey")

	// Should be approximately 10 seconds (allowing for test execution time)
	assert.True(t, remaining >= 9 && remaining <= 10, "remaining should be 9-10, got %d", remaining)
}

func TestSimpleBlocker_GetRemainingSeconds_Expired(t *testing.T) {
	blocker := newSimpleBlocker()

	// Directly set an expired block
	blocker.mu.Lock()
	blocker.m["testKey"] = time.Now().Add(-1 * time.Second)
	blocker.mu.Unlock()

	remaining := blocker.getRemainingSeconds("testKey")

	assert.Equal(t, 0, remaining)
	// Entry should be deleted
	blocker.mu.Lock()
	_, exists := blocker.m["testKey"]
	blocker.mu.Unlock()
	assert.False(t, exists)
}

// ============================================================================
// blockFor tests
// ============================================================================

func TestSimpleBlocker_BlockFor_Success(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 5)

	blocker.mu.Lock()
	_, exists := blocker.m["testKey"]
	_, allowOnceExists := blocker.allowOnce["testKey"]
	count := blocker.allowCount["testKey"]
	blocker.mu.Unlock()

	assert.True(t, exists)
	assert.True(t, allowOnceExists)
	assert.Equal(t, 3, count)
}

func TestSimpleBlocker_BlockFor_ZeroSeconds(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 0)

	blocker.mu.Lock()
	_, exists := blocker.m["testKey"]
	blocker.mu.Unlock()

	assert.False(t, exists)
}

func TestSimpleBlocker_BlockFor_NegativeSeconds(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", -5)

	blocker.mu.Lock()
	_, exists := blocker.m["testKey"]
	blocker.mu.Unlock()

	assert.False(t, exists)
}

func TestSimpleBlocker_BlockFor_UpdatesExisting(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 5)
	firstRemaining := blocker.getRemainingSeconds("testKey")

	blocker.blockFor("testKey", 15)
	secondRemaining := blocker.getRemainingSeconds("testKey")

	assert.True(t, secondRemaining > firstRemaining)
}

// ============================================================================
// consumeAllowOnceIfReady tests
// ============================================================================

func TestSimpleBlocker_ConsumeAllowOnceIfReady_NoEntry(t *testing.T) {
	blocker := newSimpleBlocker()

	result := blocker.consumeAllowOnceIfReady("nonexistent")

	assert.False(t, result)
}

func TestSimpleBlocker_ConsumeAllowOnceIfReady_NotYetReady(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 5)

	result := blocker.consumeAllowOnceIfReady("testKey")

	// Not ready yet because the block hasn't expired
	assert.False(t, result)
}

func TestSimpleBlocker_ConsumeAllowOnceIfReady_Ready(t *testing.T) {
	blocker := newSimpleBlocker()

	// Set allowOnce to a time in the past
	blocker.mu.Lock()
	blocker.allowOnce["testKey"] = time.Now().Add(-1 * time.Second)
	blocker.allowCount["testKey"] = 3
	blocker.mu.Unlock()

	result := blocker.consumeAllowOnceIfReady("testKey")

	assert.True(t, result)

	// Count should be decremented
	blocker.mu.Lock()
	count := blocker.allowCount["testKey"]
	blocker.mu.Unlock()
	assert.Equal(t, 2, count)
}

func TestSimpleBlocker_ConsumeAllowOnceIfReady_ExhaustsAllowances(t *testing.T) {
	blocker := newSimpleBlocker()

	// Set allowOnce with count 1
	blocker.mu.Lock()
	blocker.allowOnce["testKey"] = time.Now().Add(-1 * time.Second)
	blocker.allowCount["testKey"] = 1
	blocker.mu.Unlock()

	result := blocker.consumeAllowOnceIfReady("testKey")

	assert.True(t, result)

	// Entry should be deleted after exhausting allowances
	blocker.mu.Lock()
	_, allowOnceExists := blocker.allowOnce["testKey"]
	_, countExists := blocker.allowCount["testKey"]
	blocker.mu.Unlock()
	assert.False(t, allowOnceExists)
	assert.False(t, countExists)
}

func TestSimpleBlocker_ConsumeAllowOnceIfReady_ZeroCount(t *testing.T) {
	blocker := newSimpleBlocker()

	// Set allowOnce with count 0
	blocker.mu.Lock()
	blocker.allowOnce["testKey"] = time.Now().Add(-1 * time.Second)
	blocker.allowCount["testKey"] = 0
	blocker.mu.Unlock()

	result := blocker.consumeAllowOnceIfReady("testKey")

	assert.False(t, result)
}

// ============================================================================
// forceClearAllowOnce tests
// ============================================================================

func TestSimpleBlocker_ForceClearAllowOnce_Clears(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 5)

	blocker.forceClearAllowOnce("testKey")

	blocker.mu.Lock()
	_, allowOnceExists := blocker.allowOnce["testKey"]
	_, countExists := blocker.allowCount["testKey"]
	blocker.mu.Unlock()

	assert.False(t, allowOnceExists)
	assert.False(t, countExists)
}

func TestSimpleBlocker_ForceClearAllowOnce_NoEntry(t *testing.T) {
	blocker := newSimpleBlocker()

	// Should not panic when clearing non-existent entry
	blocker.forceClearAllowOnce("nonexistent")
}

func TestSimpleBlocker_ForceClearAllowOnce_PreservesBlock(t *testing.T) {
	blocker := newSimpleBlocker()

	blocker.blockFor("testKey", 5)

	blocker.forceClearAllowOnce("testKey")

	// Block should still exist
	blocker.mu.Lock()
	_, exists := blocker.m["testKey"]
	blocker.mu.Unlock()
	assert.True(t, exists)
}

// ============================================================================
// Concurrency tests
// ============================================================================

func TestSimpleBlocker_ConcurrentAccess(t *testing.T) {
	blocker := newSimpleBlocker()
	done := make(chan bool)

	// Multiple goroutines accessing the same blocker
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "key"
			blocker.blockFor(key, 1)
			blocker.getRemainingSeconds(key)
			blocker.consumeAllowOnceIfReady(key)
			blocker.forceClearAllowOnce(key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================================================
// Package-level blockers tests
// ============================================================================

func TestPackageLevelBlockers_Exist(t *testing.T) {
	// Reset package-level blockers for clean state
	ipBlocker = newSimpleBlocker()
	userBlocker = newSimpleBlocker()

	assert.NotNil(t, ipBlocker)
	assert.NotNil(t, userBlocker)
}

func TestPackageLevelBlockers_Independent(t *testing.T) {
	// Reset package-level blockers for clean state
	ipBlocker = newSimpleBlocker()
	userBlocker = newSimpleBlocker()

	ipBlocker.blockFor("192.168.1.1", 10)
	userBlocker.blockFor("testUser", 20)

	ipRemaining := ipBlocker.getRemainingSeconds("192.168.1.1")
	userRemaining := userBlocker.getRemainingSeconds("testUser")

	assert.True(t, ipRemaining >= 9 && ipRemaining <= 10)
	assert.True(t, userRemaining >= 19 && userRemaining <= 20)

	// Cross-check - they should not affect each other
	assert.Equal(t, 0, ipBlocker.getRemainingSeconds("testUser"))
	assert.Equal(t, 0, userBlocker.getRemainingSeconds("192.168.1.1"))
}
