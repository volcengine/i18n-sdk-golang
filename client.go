package i18n

import (
	"context"
	"time"
)

// Client provides the essential APIs of the i18n client SDK.
type Client interface {
	// GetPackage returns a single language i18n text package data as a whole.
	GetPackage(ctx context.Context, lang string, opts ...Option) (*Package, error)
	// GetText returns the text of the given key in a single language i18n
	// text package data as a string.
	GetText(ctx context.Context, lang, key string, opts ...Option) (string, error)
	// AddOption allows users to add some global options in order to not set them
	// in each request if they will not change frequently. It must not be called
	// concurrently and the same option set later will overwrite the former one.
	AddOption(opts ...Option)
	// Shutdown cleans the resources and exit gracefully, which should be called
	// in a deferred function in the main routine.
	Shutdown()
}

// NewClient creates an instance of client which can be used by callers as a
// handle to get i18n text package data. It implements the Client interface.
func NewClient(pid, nid int64, opts ...Option) (*client, error) {
	if pid <= 0 || nid <= 0 {
		return nil, ErrInvalidParams
	}
	c := &client{
		projectID:   pid,
		namespaceID: nid,
		options:     opts,
		shutdownCh:  make(chan struct{}),
	}
	o := option{}
	for _, f := range opts {
		f(&o)
	}
	if o.retryPolicy == nil {
		o.retryPolicy = NewBackoffRetryPolicy(3, 4000, 500)
		c.options = append(c.options, WithRetryPolicy(o.retryPolicy))
	}
	if o.logger == nil {
		o.logger = DefaultLogger()
		c.options = append(c.options, WithLogger(o.logger))
	}
	if o.metricer == nil {
		o.metricer = DefaultMetricer()
		c.options = append(c.options, WithMetricer(o.metricer))
	}
	if o.fetcher == nil {
		o.fetcher = NewHttpFetcher(append([]Option{
			WithLogger(o.logger),
			WithMetricer(o.metricer),
			WithRetryPolicy(o.retryPolicy),
		}, opts...)...)
		c.options = append(c.options, WithFetcher(o.fetcher))
	}
	if int64(o.refreshInterval) < int64(time.Second) {
		o.refreshInterval = defaultRefreshInterval
		c.options = append(c.options, WithRefreshInterval(defaultRefreshInterval))
	}
	if int64(o.cacheDuration) < int64(time.Minute) {
		o.cacheDuration = defaultCacheDuration
		c.options = append(c.options, WithCacheDuration(defaultCacheDuration))
	}

	go c.refresher(context.Background())
	return c, nil
}
