# The client Golang SDK to retrieve text from starling

## Quick Start

1. Create a client

```go
client := NewClient(
    context.Background(),
    "projectName",
    "namespace",
    WithAppKey("AppKey"))
```

The project name and namespace are created on the starling website. The `AppKey` must be provided for authorization.

2. Get a text or package

There are two types API: get a text of a given key or get a whole package.

- Get a text

```go
GetText(key string, lang string, mode ...ModeType) (val string, lang string)
GetTextWithFallback(key string, langs []string, mode ...ModeType) (val string, lang string)
GetTextWithFallbackVersion(key, lang string, fb FallbackType, ver int, mode ...ModeType) (val string, lang string, version int64)
```

- Get a package

```go
GetPackage(lang string, mode ...ModeType) (map[string]string, string, int64)
GetPackageWithFallback(langs []string, mode ...ModeType) (map[string]string, string, int64)
GetPackageWithFallbackVersion(lang string, fb FallbackType, ver int, mode ...ModeType) (map[string]string, string, int64)
```

## Options

Except the `WithAppKey` option which is required, there are a lot of options which are not required can be set:

- `WithHTTPDomain(domain string)`: set custom http domain to retrieve data if needed.
- `WithEnableHTTPs(enable bool)`: set http proxy with SSL or not.
- `WithHTTPTimeout(timeout int)`: set http proxy total request and response timeout in second.
- `WithEnableSimilar(similar bool)`: set the similar text fallback strategy or not.
- `WithEnableFallback(fallback bool)`: set the fallback storage when getting text failed.
- `WithOperator(operator string)`: set the operator user which is using the SDK to retrieve data.
- `WithRetryPolicy(policy RetryPolicy)`: set the retry policy when sending request failed.
   - `NewNoRetryPolicy()`: create a no retry policy
   - `NewBackoffRetryPolicy(maxRetry int, maxDelayMs, intervalMs int64)`: create a backoff retry policy
- `WithLogger(logger Logger)`: set the logger to output the internal state content.
- `WithMetricer(metricer Metricer)`: set the metricer to monitor the internal state.
- `WithProxyer(proxyer Proxyer)`: set the custom proxy implementation to retrieve data, default use the http proxy in this SDK.
- `WithRefreshInterval(d time.Duration)`: sets the interval time in second for background refresh which should be range from 1 second to 1 minute and default is 10 seconds.
- `WithCacheDuration(d time.Duration)`: set the duration of the local cache time which should be no shorter than 1 minute and default is 6 hours.

## Contact

Please contact starling-manager@mail.bytedance.net if encounter any problem.