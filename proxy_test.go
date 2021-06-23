package starling_goclient_public

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpProxy(t *testing.T) {
	p := NewHttpProxy()
	assert.Equal(t, Domain, p.(*httpProxy).option.httpDomain)
	assert.Equal(t, 10, p.(*httpProxy).option.httpTimeout)
	assert.Equal(t, true, p.(*httpProxy).option.enableHTTPs)

	p = NewHttpProxy(
		WithHTTPDomain("www.starling.com"),
		WithHTTPTimeout(20),
		WithEnableHTTPs(false))
	assert.Equal(t, "www.starling.com", p.(*httpProxy).option.httpDomain)
	assert.Equal(t, 20, p.(*httpProxy).option.httpTimeout)
	assert.Equal(t, false, p.(*httpProxy).option.enableHTTPs)

	key1 := apiVersionGetPackage+"/normal/project/ns/en"
	key2 := apiVersionSimilarFallback+"/normal/project/ns/en"
	key3 := apiVersionGetPackageVersion+"/normal/project/ns/en"
	p = NewHttpProxy(WithLogger(DefaultLogger()), WithMetricer(DefaultMetricer()))

	data, err := p.Retrieve(context.TODO(), "", nil)
	t.Log(data, err)
	assert.NotEmpty(t, err)

	data, err = p.Retrieve(context.TODO(), "key", nil)
	t.Log(data, err)
	assert.NotEmpty(t, err)

	data, err = p.Retrieve(context.TODO(), key1, nil)
	t.Log(data, err)
	assert.Empty(t, err)

	data, err = p.Retrieve(context.TODO(), key2, nil)
	t.Log(data, err)
	assert.Empty(t, err)

	data, err = p.Retrieve(context.TODO(), key3, nil)
	t.Log(data, err)
	assert.Empty(t, err)

	p = NewHttpProxy(WithLogger(DefaultLogger()), WithHTTPDomain("xxxxxxx"))
	rp := NewBackoffRetryPolicy(3, 4000, 1000)
	data, err = p.Retrieve(context.TODO(), key1, rp)
	t.Log(data, err)
	assert.NotEmpty(t, err)
}