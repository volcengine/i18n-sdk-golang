package starling_goclient_public

import (
	"net"
	"strings"
	"time"
)

// RetryPolicy is the strategy to direct the HTTP request retry when errors
// occurred. The strategy is make based on the retry times and error type.
type RetryPolicy interface {
	// ShouldRetry returns whether to retry based retry times and error type.
	ShouldRetry(retryTimes int, err error) bool

	// RetryDelay gives the delay duration to retry based retry times and error
	// type. It should only be called if ShouldRetry returns true.
	RetryDelay(retryTimes int) time.Duration
}

// NewNoRetryPolicy creates a retry policy which does not do retrying.
func NewNoRetryPolicy() RetryPolicy {
	return &noRetryPolicy{}
}

// noRetryPolicy does not do the retrying.
type noRetryPolicy struct{}

// ShouldRetry implements the `RetryPolicy` interface.
func (rp *noRetryPolicy) ShouldRetry(retryTimes int, err error) bool {
	return false
}

// RetryDelay implements the `RetryPolicy` interface.
func (rp *noRetryPolicy) RetryDelay(retryTimes int) time.Duration {
	return 0
}

// NewBackoffRetryPolicy creates a retry policy which does the backoff retrying.
func NewBackoffRetryPolicy(maxRetry int, maxDelayMs, intervalMs int64) RetryPolicy {
	return &backoffRetryPolicy{maxDelayMs, intervalMs, maxRetry}
}

// backoffRetryPolicy does the retrying with exponential back-off strategy.
// This policy will keep retrying until the maximum number of retries is
// reached.  The delay time will be a fixed interval for the first time, then
// 2*interval for the second time, 4*internal for the third, and so on. The
// delay time will be `2^retryTimes*interval` generally and will never exceed
// the max delay time if specified.
type backoffRetryPolicy struct {
	maxDelayMs    int64
	intervalMs    int64
	maxRetryTimes int
}

// ShouldRetry implements the `RetryPolicy` interface.
func (rp *backoffRetryPolicy) ShouldRetry(retryTimes int, err error) bool {
	// Check retry times.
	if retryTimes >= rp.maxRetryTimes {
		return false
	}

	// Check error type.
	if err == nil {
		return true
	}
	if _, ok := err.(net.Error); ok {
		return true
	}
	if strings.Contains(err.Error(), "context deadline") {
		return true
	}
	return false
}

// RetryDelay implements the `RetryPolicy` interface.
func (rp *backoffRetryPolicy) RetryDelay(retryTimes int) time.Duration {
	if retryTimes < 0 || retryTimes >= rp.maxRetryTimes {
		return 0
	}

	delayMs := (1 << retryTimes) * rp.intervalMs
	if delayMs > rp.maxDelayMs {
		delayMs = rp.maxDelayMs
	}
	return time.Millisecond * time.Duration(delayMs)
}
