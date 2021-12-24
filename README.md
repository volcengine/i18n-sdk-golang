# The Golang SDK to retrieve i18n text or package

## Quick Start

` Note: The i18n text package data can only be retrieved after released in Starling platform`

1. Create a client

A `Client` instance is the handle to do everything. You should create and release it as follows:

```go
package main

import (
    i18n "github.com/volcengine/i18n-sdk-golang"
)

var ProjectID, NamespaceID int64 = 1024, 65536
client, err := i18n.NewClient(ProjectID, NamespaceID,
    WithAppKey("AppKey"),
    WithOperator("VolcEngineAdmin"))
if err != nil {
    panic(err)
}
defer client.Shutdown()
```

The project and namespace ID are created on the platform. The `AppKey` must be provided for authorization.

The client instance should be a global variable to reuse in your service, and should 
shutdown with deferred in main routine to release resource gracefully.

Different options can be set when creating the client instance of use `AddOption`as 
the global backup options which will not used if the request-level options are also set.

```go
// Set when creating client instance.
client, err := i18n.NewClient(ProjectID, NamespaceID, WithAppKey("AppKey"))

// Set when calling AddOption method.
client.AddOption(WithLogger(customLoggerImpl), WithDisableBackupLang(true))
```

2. Get a package

Each i18n text data can be treated as a file related to a specific language. The 
SDK can retrieve a i18n text package data with the given language code as a whole:

```go
pkg, err := client.GetPackage(context.Background(), "en")
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println("timestamp number version:", pkg.Version)
    fmt.Println("string release version:", pkg.ReleaseVersion)
    fmt.Println("string language code:", pkg.Language)
    fmt.Println("key-value string pair data:", pkg.Data)
}
```

The above calling will retrieve the English i18n text package data in the normal online
environment. Different options can be passed to control the result:

- Different project or namespace data
```go
pkg, err := client.GetPackage(ctx, "en", WithProjectID(1000), WithNamespaceID(32768))
```
Request-level project and namespace ID has high priority than the global ones 
when passed in creating client instance.

- Get package data in test environment
```go
pkg, err := StarlingClient.GetPackage(ctx, "en", WithEnv("test"))
```
Default is the normal environment, and all environment can be found in `Options` chapter.

- Version control
```go
// Get the specific version data.
pkg, err := client.GetPackage(ctx, "en", WithVersion("1.2.3"))

// Only get version of the language.
pkg, err := client.GetPackage(ctx, "en", WithOnlyVersion(true))
fmt.Println(pkg.Version)
fmt.Println(pkg.ReleaseVersion)
```
Getting latest version each time is the default action, but a specific version can 
also be fetched or only the latest version. It will be not cached locally, so it 
should not be called frequently.

3. Get Single Text

When getting the i18n package data, you can also get a single text of a given
key. The regular code to get a single text show as follows:

```go
val, err := client.GetText(ctx, "en", "key1")
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println("the text value of key1 is:", val)
}
```

The options which can be passed to control the result in the `GetPackage` method
can also be passed here. There are other more options:

- Parse plural value
```go
val, err := client.GetText(ctx, "ja-JP", "key2", 
    WithPluralCount(123), 
    WithPluralDefaultLang("en"), 
)
```

- Replace variables
```go
val, err := client.GetText(ctx, "ja-JP", "key3",
    WithArgument(map[string]interface{}{"name": "Jack", "discount": 0.3},
    WithLeftDelimiter("["),
    WithRightDelimiter("]"),
)
```
The default delimiters are `{` and `}` for each variable in the text string, and
you can set other delimiters as the above code.

If a text is `His name is [name]. The pen got a discount of [discount]`, it will return `His name is Jack. The pen got a discount of 0.3` after the above processing.

`Note：DO NOT use {{ and }} as the delimiters, which are reserved by the ICU format.`


## Advanced options

There are a lot of options, which are not required, can be set for advanced usage cases.

1. Set http request params

- `WithHTTPDomain(domain string)`: set custom http domain to retrieve data if needed.
- `WithEnableHTTPs(enable bool)`: set http proxy with SSL or not.
- `WithHTTPTimeout(timeout int)`: set http proxy total request and response timeout in second.

2. Set backup settings

- `WithDisableBackupLang(val bool)`: set whether to disable backup language when getting text failed.
- `WithBackupLang(langs []string)`: sets the backup languages when getting text failed.
- `WithDisableBackupStorage(val bool)`: sets whether to disable the backup storage when getting data failed from primary storage.

3. Set internal facilities

- `WithRetryPolicy(policy RetryPolicy)`: set the retry policy when sending request failed.
   - `NewNoRetryPolicy()`: create a no retry policy
   - `NewBackoffRetryPolicy(maxRetry int, maxDelayMs, intervalMs int64)`: create a backoff retry policy
- `WithLogger(logger Logger)`: set the logger to output the internal state content.
- `WithMetricer(metricer Metricer)`: set the metricer to monitor the internal state.
- `WithFetcher(f Fetcher)`: sets the custom proxy fetcher implementation to retrieve data, default use the http proxy in this SDK

4. Set local cache setting

- `WithRefreshInterval(d time.Duration)`: sets the interval time in second for background refresh which should be longer than 1 second and default is 1 minute.
- `WithCacheDuration(d time.Duration)`: set the duration of the local cache time which should be no shorter than 1 minute and default is 6 hours.

5. All options

|option setter | meaning | required | default value |
|--------------|---------------|-----|----------|
|WithAppKey(ak string)| sets app key of the project for authorization| true | "" |
|WithEnableHTTPs(enable bool)|sets http proxy with SSL or not| false | false |
|WithHTTPTimeout(timeout int)|sets http proxy total request and response timeout in second | false | 10s |
|WithProjectID(pid int64)| set the custom project ID | false | 0 |
|WithNamespaceID(nid int64)| sets the namespace id to getting text | false | 0 |
|WithEnv(env string) | sets the environment to getting text | false | EnvNormal |
|WithLanguage(lang string) | set the custom language code to getting text | false | "" ｜
|WithVersion(ver string) | sets the release version to getting text, empty means latest | false | "" |
|WithDisableBackupLang(val bool)| sets whether to disable backup language when getting text failed | false | false |
|WithBackupLang(val []string)| sets the backup languages when getting text failed | false | nil |
|WithDisableBackupStorage(val bool) | sets whether to disable the backup storage when getting data failed | false | false |
|WithOnlyVersion(val bool)|sets whether to only get the version of a text package | false | false |
|WithOperator(operator string)| sets the user identifier which is using the SDK to retrieve data | true | "" |
|WithRetryPolicy(policy RetryPolicy) | sets the retry policy when sending request failed | false | `NewBackoffRetryPolicy(3, 4000, 500)` |
|WithLogger(logger Logger)|sets the logger to output the internal state content | false | `DefaultLogger()` |
|WithMetricer(metricer Metricer)|sets the metricer to monitor the internal state | false | `DefaultMetricer()` | 
|WithFetcher(fetcher Fetcher)| sets the custom proxy implementation to retrieve data | false | `HTTPFetcher` |
|WithRefreshInterval(d time.Duration)| sets the interval time for background local cache refresh | false | 1minute |
|WithCacheDuration(d time.Duration) | sets the duration of the local cache time | false | 6 hours |
|WithPluralCount(val interface{})| specifies the plural text count value | false | nil |
|WithPluralDefaultLang(val string)| specifies the default language code for plural text | false | "" |
|WithArguments(val map[string]interface{}) | provides the key-value pairs for template variables replacing | false | nil |
|WithLeftDelimiter(val string) | defines the left delimiter for custom variable | false | "{" |
|WithRightDelimiter(val string)| defines the right delimiter for custom variable | false | "}" |

## Contact

Please contact starling-manager@mail.bytedance.net if encounter any problem.
