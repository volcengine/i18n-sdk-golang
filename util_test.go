package i18n

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	l := DefaultLogger()
	l.Debug("test %s", "debug")
	l.Info("test %s", "info")
	l.Warn("test %s", "warn")
	l.Error("test %s", "error")
}

func TestMetricer(t *testing.T) {
	m := DefaultMetricer()

	m.EmitCounter(clientRetrieveErrorMetricsKey, 1, nil)
	m.EmitCounter(clientRetrieveErrorMetricsKey, 1, map[string]string{"type": "test"})
}

func TestGroup(t *testing.T) {
	g := Group{}
	key := "group"
	fn := func() (interface{}, error) {
		time.Sleep(time.Millisecond*100)
		return 123, nil
	}
	val, _ := g.Do(key, fn)
	assert.Equal(t, val, 123)
}

func TestComposeClientKey(t *testing.T) {
	for _, item := range []struct{
		project   string
		namespace string
		expect    string
	}{
		{"", "", "[]$#$[]"},
		{"a", "b", "[a]$#$[b]"},
		{"project", "namespace", "[project]$#$[namespace]"},
	} {
		assert.Equal(t, item.expect, composeClientKey(item.project, item.namespace))
	}
}

func TestRetry(t *testing.T) {
	noRetry := NewNoRetryPolicy()
	assert.Equal(t, false, noRetry.ShouldRetry(0, nil))
	assert.Equal(t, time.Duration(0), noRetry.RetryDelay(0))

	backoff := NewBackoffRetryPolicy(3, 3000, 1000)
	assert.Equal(t, true, backoff.ShouldRetry(0, nil))
	assert.Equal(t, true, backoff.ShouldRetry(0, net.InvalidAddrError("invalid addr")))
	assert.Equal(t, true, backoff.ShouldRetry(1, fmt.Errorf("context deadline")))
	assert.Equal(t, false, backoff.ShouldRetry(2, fmt.Errorf("unknown error")))
	assert.Equal(t, false, backoff.ShouldRetry(3, nil))

	assert.Equal(t, time.Second, backoff.RetryDelay(0))
	assert.Equal(t, 2*time.Second, backoff.RetryDelay(1))
	assert.Equal(t, 3*time.Second, backoff.RetryDelay(2))
	assert.Equal(t, time.Duration(0), backoff.RetryDelay(3))
}
