package i18n

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOption(t *testing.T) {
	retry := NewNoRetryPolicy()
	logger := DefaultLogger()
	metricer := DefaultMetricer()
	fetcher := NewHttpFetcher()
	for _, item := range []struct {
		input  Option
		expect interface{}
	}{
		{WithAppKey("APPKEY"), option{appKey: "APPKEY"}},
		{WithHTTPDomain("www.starling.com"), option{httpDomain: "www.starling.com"}},
		{WithEnableHTTPs(true), option{enableHTTPs: true}},
		{WithHTTPTimeout(10), option{httpTimeout: 10}},
		{WithProjectID(123), option{projectID: 123}},
		{WithNamespaceID(456), option{namespaceID: 456}},
		{WithEnv("normal"), option{env: "normal"}},
		{WithLanguage("en"), option{language: "en"}},
		{WithVersion("1.2.3"), option{version: "1.2.3"}},
		{WithDisableBackupLang(true), option{disableBackupLang: true}},
		{WithBackupLang([]string{"de", "es"}), option{backupLang: []string{"de", "es"}}},
		{WithDisableBackupStorage(true), option{disableBackupStorage: true}},
		{WithOnlyVersion(true), option{onlyVersion: true}},
		{WithOperator("operator"), option{operator: "operator"}},
		{WithRetryPolicy(retry), option{retryPolicy: retry}},
		{WithLogger(logger), option{logger: logger}},
		{WithMetricer(metricer), option{metricer: metricer}},
		{WithFetcher(fetcher), option{fetcher: fetcher}},
		{WithRefreshInterval(time.Second), option{refreshInterval: time.Second}},
		{WithCacheDuration(time.Hour), option{cacheDuration: time.Hour}},
		{WithPluralCount(10), option{pluralCount: 10}},
		{WithPluralDefaultLang("en"), option{pluralDefaultLang: "en"}},
		{WithArguments(map[string]interface{}{"count": 1}), option{arguments: map[string]interface{}{"count": 1}}},
		{WithLeftDelimiter("["), option{leftDelimiter: "["}},
		{WithRightDelimiter("]"), option{rightDelimiter: "]"}},
	} {
		o := op.get()
		item.input(o)
		assert.Equal(t, item.expect, *o)
		op.put(o)
	}
}
