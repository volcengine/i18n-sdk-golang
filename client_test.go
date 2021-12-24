package i18n

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mockData = map[string]string{
		"key1": "v1",
		"key2": "v2",
		"key3": "He comes from [[country]], whose name is [[name]].",
		"key4": "{num, plural, one {Up to {discount} with {count}+ item} other {Up to {discount} with {count}+ items}}",
		"key5": "{number, plural, one {I have # apple from {farm}.} other {I have # apples from {farm}.}}",
	}
)

type mockFetcher struct{}

func (m *mockFetcher) Fetch(ctx context.Context, pid, nid int64, lang string, opts ...Option) (*Package, error) {
	if lang == "INVALID" {
		return nil, ErrBackToSourceFailed
	}
	if pid < 0 || nid < 0 {
		return nil, ErrInvalidParams
	}
	p, n := strconv.FormatInt(pid, 10), strconv.FormatInt(nid, 10)
	return &Package{
		Version:        p + n,
		ReleaseVersion: p + "." + n,
		Data:           mockData,
		Language:       lang,
	}, nil
}

func (m *mockFetcher) FetchVersion(ctx context.Context, pid, nid int64, lang string, opts ...Option) (int64, string, error) {
	if lang == "INVALID" {
		return 0, "", ErrBackToSourceFailed
	}
	if pid < 0 || nid < 0 {
		return 0, "", ErrInvalidParams
	}
	return pid + nid, strconv.FormatInt(pid, 10) + "." + strconv.FormatInt(nid, 10), nil
}

func TestClientNew(t *testing.T) {
	rp := NewBackoffRetryPolicy(3, 4000, 500)
	lg := DefaultLogger()
	me := DefaultMetricer()
	ft := NewHttpFetcher(WithLogger(lg), WithMetricer(me), WithRetryPolicy(rp))

	for _, item := range []struct {
		pid  int64
		nid  int64
		opts []Option

		opt *option
		err error
	}{
		{
			err: ErrInvalidParams,
		},
		{
			pid: 1,
			nid: 2,
		},
		{
			pid:  1,
			nid:  2,
			opts: []Option{WithAppKey("app12345"), WithRetryPolicy(rp), WithLogger(lg), WithMetricer(me), WithFetcher(ft)},

			opt: &option{
				appKey:          "app12345",
				retryPolicy:     rp,
				logger:          lg,
				metricer:        me,
				fetcher:         ft,
				refreshInterval: defaultRefreshInterval,
				cacheDuration:   defaultCacheDuration,
			},
		},
	} {
		c, err := NewClient(item.pid, item.nid, item.opts...)
		assert.Equal(t, item.err, err)
		if err != nil {
			assert.Nil(t, c)
			continue
		}
		assert.NotNil(t, c)

		if item.opt == nil {
			assert.Equal(t, 6, len(c.options))
			c.Shutdown()
			continue
		}

		o := &option{}
		for _, f := range c.options {
			f(o)
		}
		assert.Equal(t, item.opt, o)
		c.Shutdown()
	}
}

func TestClientGetPackage(t *testing.T) {
	c, err := NewClient(1, 2, WithFetcher(&mockFetcher{}))
	assert.NotNil(t, c)
	assert.Nil(t, err)
	defer c.Shutdown()

	pkg, err := c.GetPackage(context.TODO(), "en", WithProjectID(-1))
	assert.Equal(t, ErrInvalidParams, err)
	assert.Nil(t, nil, pkg)

	// Test get from fetcher.
	pkg, err = c.GetPackage(context.TODO(), "en")
	assert.Nil(t, nil, err)
	t.Log(pkg)
	pkg, err = c.GetPackage(context.TODO(), "en", WithVersion("1.0"))
	assert.Nil(t, nil, err)
	t.Log(pkg)

	// Test get from local cache.
	pkg, err = c.GetPackage(context.TODO(), "en")
	assert.Nil(t, nil, err)
	t.Log(pkg)
	pkg, err = c.GetPackage(context.TODO(), "en", WithVersion("1.0"))
	assert.Nil(t, nil, err)
	t.Log(pkg)

	// Test fetch failed.
	pkg, err = c.GetPackage(context.TODO(), "INVALID")
	assert.Equal(t, ErrBackToSourceFailed, err)
	assert.Nil(t, pkg)

	// Test fetch version only.
	pkg, err = c.GetPackage(context.TODO(), "en", WithOnlyVersion(true))
	assert.Nil(t, err)
	assert.Equal(t, "1.2", pkg.ReleaseVersion)

	// Test fetch version only failed.
	pkg, err = c.GetPackage(context.TODO(), "INVALID", WithOnlyVersion(true))
	assert.Equal(t, ErrBackToSourceFailed, err)
	assert.Nil(t, pkg)
}

func TestClientGetText(t *testing.T) {
	c, err := NewClient(1, 2, WithFetcher(&mockFetcher{}))
	assert.NotNil(t, c)
	assert.Nil(t, err)
	defer c.Shutdown()

	// Test get data failed.
	text, err := c.GetText(context.TODO(), "INVALID", "k1")
	assert.Equal(t, ErrBackToSourceFailed, err)
	assert.Empty(t, text)

	// Test key not exist.
	text, err = c.GetText(context.TODO(), "en", "not-exist-key")
	assert.Equal(t, ErrKeyNotExist, err)
	assert.Empty(t, text)

	// Test plain text.
	text, err = c.GetText(context.TODO(), "en", "key1")
	assert.Nil(t, err)
	assert.Equal(t, "v1", text)

	// Test plural text and default variable.
	text, err = c.GetText(context.TODO(), "en", "key4", WithPluralCount(100))
	assert.Nil(t, err)
	assert.Equal(t, "Up to {discount} with {count}+ items", text)
	t.Log(text)
	text, err = c.GetText(context.TODO(), "en", "key4",
		WithPluralCount(100),
		WithArguments(map[string]interface{}{"discount": "30%", "count": 100}))
	assert.Nil(t, err)
	assert.Equal(t, "Up to 30% with 100+ items", text)
	t.Log(text)
	text, err = c.GetText(context.TODO(), "en", "key5",
		WithPluralCount(10),
		WithArguments(map[string]interface{}{"farm": "ByteDance"}))
	assert.Nil(t, err)
	assert.Equal(t, "I have 10 apples from ByteDance.", text)
	t.Log(text)

	// Test custom variable.
	text, err = c.GetText(context.TODO(), "en", "key3",
		WithLeftDelimiter("[["),
		WithRightDelimiter("]]"),
		WithArguments(map[string]interface{}{"country": "China", "name": "Jack"}))
	assert.Nil(t, err)
	assert.Equal(t, "He comes from China, whose name is Jack.", text)
	t.Log(text)
	text, err = c.GetText(context.TODO(), "en", "key3",
		WithLeftDelimiter("[["),
		WithRightDelimiter("]]"),
		WithArguments(map[string]interface{}{"country": "China"}))
	assert.Nil(t, err)
	assert.Equal(t, "He comes from China, whose name is .", text)
	t.Log(text)
}

func TestClientHandleOption(t *testing.T) {
	c, err := NewClient(1, 2)
	assert.NotNil(t, c)
	assert.Nil(t, err)
	defer c.Shutdown()

	for _, item := range []struct {
		opt  *option
		lang string
		opts []Option

		expected *option
		err      error
	}{
		{
			err: ErrInvalidParams,
		},
		{
			opt:      &option{},
			lang:     "en",
			expected: &option{projectID: 1, namespaceID: 2, env: EnvNormal, language: "en"},
		},
		{
			opt:      &option{},
			opts:     []Option{WithProjectID(-1)},
			expected: nil,
			err:      ErrInvalidParams,
		},
	} {
		optArr, err := c.handleOptions(item.opt, item.lang, item.opts...)
		assert.Equal(t, item.err, err)
		if err != nil {
			continue
		}

		actual := &option{}
		for _, f := range optArr {
			f(actual)
		}
		for _, f := range c.options {
			f(item.expected)
		}
		assert.Equal(t, item.expected, actual)
		t.Log(actual, err)
	}
}

func TestClientProcessPlural(t *testing.T) {
	c, err := NewClient(1, 2, WithFetcher(&mockFetcher{}))
	assert.NotNil(t, c)
	assert.Nil(t, err)
	defer c.Shutdown()

	for _, item := range []struct {
		raw     string
		lang    string
		defLang string
		count   interface{}

		result string
		err    error
	}{
		{
			raw: "",
			err: ErrInvalidICUFormat,
		},
		{
			raw: "I have a apple",
			err: ErrInvalidICUFormat,
		},
		{
			raw:    "I have {num, plural, one {# apple}, other {{num} apples}} to store.",
			lang:   "en",
			count:  123,
			result: "I have 123 apples to store.",
		},
		{
			raw:     "I have {num, plural, one {# apple}, other {# apples}} to store.",
			lang:    "en",
			defLang: "en",
			count:   1,
			result:  "I have 1 apple to store.",
		},
	} {
		res, err := c.processPlural(item.raw, item.lang, item.defLang, item.count)
		t.Log(res, err)
		assert.Equal(t, item.err, err)
		assert.Equal(t, item.result, res)
	}
}

func TestClientProcessVars(t *testing.T) {
	c, err := NewClient(1, 2, WithFetcher(&mockFetcher{}))
	assert.NotNil(t, c)
	assert.Nil(t, err)
	defer c.Shutdown()

	for _, item := range []struct {
		raw   string
		vars  map[string]interface{}
		left  string
		right string

		result string
		err    error
	}{
		{},
		{
			raw:    "a test text",
			result: "a test text",
		},
		{
			raw:    "I have {num} apples",
			vars:   map[string]interface{}{"num": 123},
			result: "I have 123 apples",
		},
		{
			raw:    "I have [count] apples with [attitude]",
			vars:   map[string]interface{}{"count": 10, "attitude": "happiness"},
			left:   "[",
			right:  "]",
			result: "I have 10 apples with happiness",
		},
		{
			raw:    "I have [[count]] apples with [[attitude]]",
			vars:   map[string]interface{}{"count": 10},
			left:   "[[",
			right:  "]]",
			result: "I have 10 apples with ",
		},
		{
			raw:    "I have [[count]] apples with [[attitude]]",
			vars:   map[string]interface{}{"count": 10},
			result: "I have [[count]] apples with [[attitude]]",
		},
	} {
		res, err := c.processVars(item.raw, item.vars, item.left, item.right)
		t.Log(res, err)
		assert.Equal(t, item.err, err)
		assert.Equal(t, item.result, res)
	}
}
