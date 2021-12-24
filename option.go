package i18n

import (
	"sync"
	"time"
)

// Option is the options builder which can be set by the client.
type Option func(h *option)

type option struct {
	appKey               string
	httpDomain           string
	enableHTTPs          bool
	httpTimeout          int
	projectID            int64
	namespaceID          int64
	env                  string
	language             string
	version              string
	disableBackupLang    bool
	backupLang           []string
	disableBackupStorage bool
	onlyVersion          bool
	operator             string
	retryPolicy          RetryPolicy
	logger               Logger
	metricer             Metricer
	fetcher              Fetcher
	refreshInterval      time.Duration
	cacheDuration        time.Duration
	pluralCount          interface{}
	pluralDefaultLang    string
	arguments            map[string]interface{}
	leftDelimiter        string
	rightDelimiter       string
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

// WithProjectID sets the project id to getting text.
func WithProjectID(pid int64) Option {
	return func(o *option) {
		o.projectID = pid
	}
}

// WithNamespaceID sets the namespace id to getting text.
func WithNamespaceID(nid int64) Option {
	return func(o *option) {
		o.namespaceID = nid
	}
}

// WithEnv sets the environment to getting text.
func WithEnv(env string) Option {
	return func(o *option) {
		o.env = env
	}
}

// WithLanguage sets the language code to getting text.
func WithLanguage(lang string) Option {
	return func(o *option) {
		o.language = lang
	}
}

// WithVersion sets the release version to getting text.
func WithVersion(ver string) Option {
	return func(o *option) {
		o.version = ver
	}
}

// WithDisableBackupLang sets whether to disable backup language when getting text failed.
func WithDisableBackupLang(val bool) Option {
	return func(o *option) {
		o.disableBackupLang = val
	}
}

// WithBackupLang sets the backup languages when getting text failed.
func WithBackupLang(val []string) Option {
	return func(o *option) {
		o.backupLang = val
	}
}

// WithDisableBackupStorage sets whether to disable the backup storage when getting data failed.
func WithDisableBackupStorage(val bool) Option {
	return func(o *option) {
		o.disableBackupStorage = val
	}
}

// WithOnlyVersion sets whether to only get the version of a text package.
func WithOnlyVersion(val bool) Option {
	return func(o *option) {
		o.onlyVersion = val
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

// WithFetcher sets the custom proxy implementation to retrieve data.
func WithFetcher(fetcher Fetcher) Option {
	return func(o *option) {
		o.fetcher = fetcher
	}
}

// WithRefreshInterval sets the interval time in second for background refresh
// which should be longer than 1 second and default is 1 minute.
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

// WithPluralCount specifies the plural text count value.
func WithPluralCount(val interface{}) Option {
	return func(o *option) {
		o.pluralCount = val
	}
}

// WithPluralDefaultLang specifies the default language code for plural text.
func WithPluralDefaultLang(val string) Option {
	return func(o *option) {
		o.pluralDefaultLang = val
	}
}

// WithArguments provides the key-value pairs for template variables replacing.
func WithArguments(val map[string]interface{}) Option {
	return func(o *option) {
		o.arguments = val
	}
}

// WithLeftDelimiter defines the left delimiter for custom variable, default is '{'.
func WithLeftDelimiter(val string) Option {
	return func(o *option) {
		o.leftDelimiter = val
	}
}

// WithRightDelimiter defines the right delimiter for custom variable, default is '}'.
func WithRightDelimiter(val string) Option {
	return func(o *option) {
		o.rightDelimiter = val
	}
}

// optionPool manages the option objects based on `sync.Pool` for reuse.
type optionPool struct {
	sync.Pool
}

func (p *optionPool) get() (obj *option) {
	return p.Pool.Get().(*option)
}

func (p *optionPool) put(obj *option) {
	if obj == nil {
		return
	}
	if obj != nil {
		obj.appKey = ""
		obj.httpDomain = ""
		obj.enableHTTPs = false
		obj.httpTimeout = 0
		obj.projectID = 0
		obj.namespaceID = 0
		obj.env = ""
		obj.language = ""
		obj.version = ""
		obj.disableBackupLang = false
		obj.backupLang = nil
		obj.disableBackupStorage = false
		obj.onlyVersion = false
		obj.operator = ""
		obj.retryPolicy = nil
		obj.logger = nil
		obj.metricer = nil
		obj.fetcher = nil
		obj.refreshInterval = 0
		obj.cacheDuration = 0
		obj.pluralCount = nil
		obj.pluralDefaultLang = ""
		obj.arguments = nil
		obj.leftDelimiter = ""
		obj.rightDelimiter = ""
	}
	p.Pool.Put(obj)
}
