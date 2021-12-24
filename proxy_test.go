package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpFetcher(t *testing.T) {
	p := NewHttpFetcher()
	assert.Equal(t, Domain, p.option.httpDomain)
	assert.Equal(t, 10, p.option.httpTimeout)
	assert.Equal(t, false, p.option.enableHTTPs)

	p = NewHttpFetcher(
		WithHTTPDomain("www.starling.com"),
		WithHTTPTimeout(20),
		WithEnableHTTPs(true))
	assert.Equal(t, "www.starling.com", p.option.httpDomain)
	assert.Equal(t, 20, p.option.httpTimeout)
	assert.Equal(t, true, p.option.enableHTTPs)

	p = NewHttpFetcher(WithLogger(DefaultLogger()), WithMetricer(DefaultMetricer()))

	ver, rel, err := p.FetchVersion(context.TODO(), 0, 0, "de")
	t.Log(ver, rel, err)
	assert.NotEmpty(t, err)

	addr := Domain
	ver, rel, err = p.FetchVersion(context.TODO(), 4568, 39174, "en",
		WithAppKey("704dbe7057f511ec8e4aedf71dc34d4f"), WithHTTPDomain(addr))
	t.Log(ver, rel, err)
	assert.Empty(t, err)

	ver, rel, err = p.FetchVersion(context.TODO(), 4568, 39174, "de",
		WithDisableBackupLang(false),
		WithAppKey("704dbe7057f511ec8e4aedf71dc34d4f"), WithHTTPDomain(addr))
	t.Log(ver, rel, err)
	assert.Empty(t, err)

	ver, rel, err = p.FetchVersion(context.TODO(), 4568, 39174, "de",
		WithDisableBackupLang(true),
		WithAppKey("704dbe7057f511ec8e4aedf71dc34d4f"), WithHTTPDomain(addr))
	t.Log(ver, rel, err)
	assert.Empty(t, err)

	data, err := p.Fetch(context.TODO(), 0, 0, "en")
	t.Log(data, err)
	assert.NotEmpty(t, err)

	data, err = p.Fetch(context.TODO(), 4568, 39174, "en",
		WithAppKey("704dbe7057f511ec8e4aedf71dc34d4f"), WithHTTPDomain(addr))
	t.Log(data, err)
	assert.Empty(t, err)

	data, err = p.Fetch(context.TODO(), 4568, 39174, "de",
		WithDisableBackupLang(false),
		WithAppKey("704dbe7057f511ec8e4aedf71dc34d4f"), WithHTTPDomain(addr))
	t.Log(data, err)
	assert.Empty(t, err)

	rp := NewBackoffRetryPolicy(3, 4000, 1000)
	p = NewHttpFetcher(WithLogger(DefaultLogger()), WithHTTPDomain("xxxxxxx"), WithRetryPolicy(rp),
		WithAppKey("12345678"))
	ver, rel, err = p.FetchVersion(context.TODO(), 1, 2, "")
	t.Log(ver, rel, err)
	assert.NotEmpty(t, err)

	data, err = p.Fetch(context.TODO(), 1, 2, "",
		WithDisableBackupStorage(true),
		WithAppKey("12345678"))
	t.Log(ver, rel, err)
	assert.NotEmpty(t, err)
}
