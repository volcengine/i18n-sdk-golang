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
	proxyer := NewHttpProxy()
	for _, item := range []struct{
		input  Option
		expect interface{}
	}{
		{WithAppKey("APPKEY"), option{appKey: "APPKEY"}},
		{WithHTTPDomain("www.starling.com"), option{httpDomain: "www.starling.com"}},
		{WithEnableHTTPs(true), option{enableHTTPs: true}},
		{WithHTTPTimeout(10), option{httpTimeout: 10}},
		{WithEnableSimilar(true), option{enableSimilar: true}},
		{WithEnableFallback(true), option{enableFallBack: true}},
		{WithOperator("operator"), option{operator: "operator"}},
		{WithRetryPolicy(retry), option{retryPolicy: retry}},
		{WithLogger(logger), option{logger: logger}},
		{WithMetricer(metricer), option{metricer: metricer}},
		{WithProxyer(proxyer), option{proxyer: proxyer}},
		{WithRefreshInterval(time.Second), option{refreshInterval: time.Second}},
		{WithCacheDuration(time.Hour), option{cacheDuration: time.Hour}},
	} {
		o := option{}
		item.input(&o)
		assert.Equal(t, item.expect, o)
	}
}
