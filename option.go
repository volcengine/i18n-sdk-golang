package starling_goclient_public

import "time"

// Option is the options builder which can be set by the client.
type Option func(h *option)

type option struct {
	appKey          string
	httpDomain      string
	enableHTTPs     bool
	httpTimeout     int
	enableSimilar   bool
	enableFallBack  bool
	operator        string
	retryPolicy     RetryPolicy
	logger          Logger
	metricer        Metricer
	proxyer         Proxyer
	refreshInterval time.Duration
	cacheDuration   time.Duration
}

// WithAppKey sets app key of the project for authorization.
func WithAppKey(ak string) Option {
	return func(o *option) {
		o.appKey = ak
	}
}

// WithHTTPDomain sets custom http domain if needed.
func WithHTTPDomain(domain string) Option {
	return func(o *option) {
		o.httpDomain = domain
	}
}

// WithEnableHTTPs sets http proxy with SSL or not.
func WithEnableHTTPs(enable bool) Option {
	return func(o *option) {
		o.enableHTTPs = enable
	}
}

// WithHTTPTimeout sets http proxy total request and response timeout in second.
func WithHTTPTimeout(timeout int) Option {
	return func(o *option) {
		o.httpTimeout = timeout
	}
}

// WithEnableSimilar sets the similar text fallback strategy.
func WithEnableSimilar(similar bool) Option {
	return func(o *option) {
		o.enableSimilar = similar
	}
}

// WithEnableFallback sets the fallback storage when getting text failed.
func WithEnableFallback(fallback bool) Option {
	return func(o *option) {
		o.enableFallBack = fallback
	}
}

// WithOperator sets the user which is using the SDK to retrieve data.
func WithOperator(operator string) Option {
	return func(o *option) {
		o.operator = operator
	}
}

// WithRetryPolicy sets the retry policy when sending request failed.
func WithRetryPolicy(policy RetryPolicy) Option {
	return func(o *option) {
		o.retryPolicy = policy
	}
}

// WithLogger sets the logger to output the internal state content.
func WithLogger(logger Logger) Option {
	return func(o *option) {
		o.logger = logger
	}
}

// WithMetricer sets the metricer to monitor the internal state.
func WithMetricer(metricer Metricer) Option {
	return func(o *option) {
		o.metricer = metricer
	}
}

// WithProxyer sets the custom proxy implementation to retrieve data.
func WithProxyer(proxyer Proxyer) Option {
	return func(o *option) {
		o.proxyer = proxyer
	}
}

// WithRefreshInterval sets the interval time in second for background refresh
// which should be range from 1 second to 1 minute and default is 10 seconds.
func WithRefreshInterval(d time.Duration) Option {
	return func(o *option) {
		o.refreshInterval = d
	}
}

// WithCacheDuration sets the duration of the local cache time which should be
// no shorter than 1 minute and default is 6 hours.
func WithCacheDuration(d time.Duration) Option {
	return func(o *option) {
		o.cacheDuration = d
	}
}
