package i18n

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	c := NewClient(
		context.TODO(),
		"project",
		"ns",
		WithProxyer(&testProxy{}))
	assert.NotEmpty(t, c)
	defer c.Destroy()
	c1 := NewClient(context.TODO(), "project", "ns")
	assert.Equal(t, c, c1)

	c.EnableFallback()
	assert.Equal(t, true, c.enFblang)
	c.DisableFallback()
	assert.Equal(t, false, c.enFblang)

	c.EnableFallback()
	c.SetDefaultFallbackLang([]string{})
	assert.Equal(t, false, c.enFblang)
	assert.Equal(t, 0, len(c.defFblang))
	c.SetDefaultFallbackLang([]string{"en"})
	assert.Equal(t, []string{"en"}, c.defFblang)

	val, lang := c.GetText("s", "en", ModeGray)
	t.Log(val, lang)
	val, lang = c.GetText("s", "en", ModeTest)
	t.Log(val, lang)
	val, lang = c.GetText("s", "zh-Hans", ModeTest)
	t.Log(val, lang)
	val, lang = c.GetText("key1", "zh-Hans")
	t.Log(val, lang)
	assert.Equal(t, "设置成功", val)
	val, lang, ver := c.GetTextWithFallbackVersion("key1", "zh-Hans", FallbackLangDefault, 1)
	t.Log(val, lang, ver)

	pkg, lang, ver := c.GetPackage("zh-Hans")
	t.Log(pkg, lang, ver)
	pkg, lang, ver = c.GetPackageWithFallbackVersion("zh-Hans", FallbackLangCustom, 1)
	t.Log(pkg, lang, ver)

	t.Log(c.Dump())
	time.Sleep(11*time.Second)
	t.Log(c.Dump())
}

type testProxy struct{}

func (t *testProxy) Retrieve(ctx context.Context, key string, rp RetryPolicy) (*Data, error) {
	if strings.Contains(key, string(ModeGray)) {
		return &Data{}, fmt.Errorf("proxy return error: %s", key)
	}
	if strings.Contains(key, string(ModeTest)) {
		if strings.Contains(key, "en") {
			return &Data{Value: "1234567890{xxx}"}, nil
		}
		return &Data{}, nil
	}
	return &Data{
		Key: key,
		Value: `           {
		"version": 1,
		"data": {
			"key1": "设置成功",
			"key2": "修改头像",
			"key3": "上传头像失败"
		},
		"lang": "zh-Hans"}`,
		Lang:  "zh-Hans",
	}, nil
}
