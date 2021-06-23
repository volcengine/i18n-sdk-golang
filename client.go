package starling_goclient_public

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var gClientMap sync.Map

// Client provides all functionality about retrieving data from the remote
// starling server of the given project and namespace. It will cache the data
// and fetch the data in background to avoid the overload.
type Client struct {
	ctx       context.Context
	cancel    context.CancelFunc
	option    *option
	mu        sync.RWMutex
	data      sync.Map
	sf        Group
	project   string
	namespace string
	enFblang  bool
	defFblang []string
}

// NewClient creates an instance of the Client for the given project and
// namespace as well as the custom options.
func NewClient(ctx context.Context, project, namespace string, opts ...Option) *Client {
	key := composeClientKey(project, namespace)
	if c, exist := gClientMap.Load(key); exist {
		return c.(*Client)
	}

	o := &option{}
	for i := range opts {
		opts[i](o)
	}
	if o.retryPolicy == nil {
		o.retryPolicy = NewBackoffRetryPolicy(3, 4000, 500)
	}
	if o.logger == nil {
		o.logger = DefaultLogger()
	}
	if o.metricer == nil {
		o.metricer = DefaultMetricer()
	}
	if o.proxyer == nil {
		o.proxyer = NewHttpProxy(append([]Option{
			WithLogger(o.logger),
			WithMetricer(o.metricer),
			WithRetryPolicy(o.retryPolicy),
		}, opts...)...)
	}
	if int64(o.refreshInterval) < int64(time.Second) || int64(o.refreshInterval) > int64(time.Minute) {
		o.refreshInterval = defaultRefreshInterval
	}
	if int64(o.cacheDuration) < int64(time.Minute) {
		o.cacheDuration = defaultCacheDuration
	}
	c := &Client{
		option:    o,
		project:   project,
		namespace: namespace,
		enFblang:  true,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	gClientMap.Store(key, c)
	go c.refresher()
	return c
}

// Destroy cleans all resource for current client instance with thread-safety
// and should be called by the creator in a deferred form for better practice.
func (c *Client) Destroy() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
	}
	gClientMap.Delete(composeClientKey(c.project, c.namespace))
}

// EnableFallback enables the fallback lang strategy.
func (c *Client) EnableFallback() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enFblang = true
}

// DisableFallback disables the fallback lang strategy.
func (c *Client) DisableFallback() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enFblang = false
}

// SetDefaultFallbackLang set the default fallback lang if fallback strategy
// enabled and will disable it if the given lang is empty.
func (c *Client) SetDefaultFallbackLang(lang []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(lang) == 0 {
		c.enFblang = false
	}
	c.defFblang = lang
}

// GetText gets a translated text for the given key and language.
func (c *Client) GetText(key string, lang string, mode ...ModeType) (string, string) {
	langs := append([]string{lang}, c.defaultFallbackLang()...)
	return c.GetTextWithFallback(key, langs, mode...)
}

// GetTextWithFallback gets a translated text for the given key and language
// with provided fallback languages.
func (c *Client) GetTextWithFallback(key string, langs []string, mode ...ModeType) (string, string) {
	if len(langs) == 0 {
		return "", ""
	}
	m := ModeNormal
	if len(mode) != 0 {
		m = mode[0]
	}
	var data *ParsedData
	if c.option.enableSimilar {
		data = c.get(langs, &key, m, FallbackLangSimilar)
	} else {
		data = c.get(langs, &key, m, FallbackLangCustom)
	}
	return data.Data[key], data.Lang
}

func (c *Client) GetTextWithFallbackVersion(key, lang string, fb FallbackType, ver int, mode ...ModeType) (string, string, int64) {
	m := ModeNormal
	if len(mode) != 0 {
		m = mode[0]
	}
	data := c.getWithVersion(lang, &key, m, fb, strconv.Itoa(ver))
	return data.Data[key], data.Lang, data.Version
}

// GetPackage returns the whole text package of the given language.
func (c *Client) GetPackage(lang string, mode ...ModeType) (map[string]string, string, int64) {
	langs := append([]string{lang}, c.defaultFallbackLang()...)
	return c.GetPackageWithFallback(langs, mode...)
}

// GetPackageWithFallback returns the whole package with fallback language.
func (c *Client) GetPackageWithFallback(langs []string, mode ...ModeType) (map[string]string, string, int64) {
	if len(langs) == 0 {
		return map[string]string{}, "", 0
	}
	m := ModeNormal
	if len(mode) != 0 {
		m = mode[0]
	}
	var data *ParsedData
	if c.option.enableSimilar {
		data = c.get(langs, nil, m, FallbackLangSimilar)
	} else {
		data = c.get(langs, nil, m, FallbackLangCustom)
	}
	mapData := make(map[string]string, len(data.Data))
	for k, v := range data.Data {
		mapData[k] = v
	}
	return mapData, data.Lang, data.Version
}

// GetPackageWithFallbackVersion returns the whole package with fallback language and version.
func (c *Client) GetPackageWithFallbackVersion(lang string, fb FallbackType, ver int, mode ...ModeType) (map[string]string, string, int64) {
	m := ModeNormal
	if len(mode) != 0 {
		m = mode[0]
	}
	data := c.getWithVersion(lang, nil, m, fb, strconv.Itoa(ver))
	mapData := make(map[string]string, len(data.Data))
	for k, v := range data.Data {
		mapData[k] = v
	}
	return mapData, data.Lang, data.Version
}

// Dump returns all data in the local cache and is commonly used for testing.
func (c *Client) Dump() map[string]interface{} {
	data := make(map[string]interface{})
	c.data.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			data[k] = value
		}
		return true
	})
	return data
}

func (c *Client) defaultFallbackLang() []string {
	if !c.enFblang {
		return []string{}
	}
	if len(c.defFblang) > 0 {
		return c.defFblang
	}
	return []string{LocalLang}
}

func (c *Client) get(lang []string, key *string, mode ModeType, fb FallbackType) *ParsedData {
	getOne := func(lg string) (*ParsedData, string){
		result := &ParsedData{}
		var proxyKey string
		if fb == FallbackLangSimilar {
			proxyKey = fmt.Sprintf(keyFmt, apiVersionSimilarFallback, mode, c.project, c.namespace, lg)
		} else {
			proxyKey = fmt.Sprintf(keyFmt, apiVersionGetPackage, mode, c.project, c.namespace, lg)
		}
		if len(strings.Trim(lg, " ")) == 0 {
			return result, proxyKey
		}
		result = c.getFromCache(proxyKey, result)
		return result, proxyKey
	}

	data, pk := getOne(lang[0])
	if key == nil && len(data.Data) == 0 {
		c.option.metricer.EmitCounter(clientPackageEmptyMetricsKey, 1, map[string]string{"lang": lang[0], "proxy_key": pk})
		for _, lgb := range lang[1:] {
			data, pk := getOne(lgb)
			if len(data.Data) != 0 {
				return data
			}
			c.option.metricer.EmitCounter(clientPackageEmptyMetricsKey, 1, map[string]string{"lang": lgb, "proxy_key": pk})
		}
	}

	if key != nil {
		if _, exist := data.Data[*key]; exist {
			return data
		}
		c.option.metricer.EmitCounter(clientKeyEmptyMetricsKey, 1, map[string]string{"lang": lang[0], "proxy_key": pk})
		for _, lgb := range lang[1:] {
			data, pk := getOne(lgb)
			if len(data.Data) == 0 {
				c.option.metricer.EmitCounter(clientKeyEmptyMetricsKey, 1, map[string]string{"lang": lgb, "proxy_key": pk})
				continue
			}
			if _, exist := data.Data[*key]; exist {
				return data
			}
		}
	}
	return data
}

func (c *Client) getWithVersion(lang string, key *string, mode ModeType, fb FallbackType, version string) *ParsedData {
	result := &ParsedData{}
	fbStr := strconv.Itoa(int(fb))
	proxyKey := fmt.Sprintf(keyFmtVersion,
		apiVersionGetPackageVersion, mode, c.project, c.namespace, lang, version, fbStr)
	result = c.getFromCache(proxyKey, result)
	if len(result.Data) == 0 {
		c.option.metricer.EmitCounter(clientPackageEmptyMetricsKey, 1, map[string]string{"lang": lang, "proxy_key": proxyKey})
		return result
	}
	if key != nil {
		if _, exist := result.Data[*key]; exist {
			return result
		}
		c.option.metricer.EmitCounter(clientKeyEmptyMetricsKey, 1, map[string]string{"lang": lang, "proxy_key": proxyKey})
	}
	return result
}

func (c *Client) getFromCache(key string, defaultVal *ParsedData) *ParsedData {
	now := time.Now()
	if val, exist := c.data.Load(key); exist {
		if realVal, ok := val.(*ParsedData); ok {
			realVal.Atime = now
			c.data.Store(key, realVal)
			return realVal
		}
	}

	p, err := c.sf.Do(key, func() (interface{}, error) {
		return c.getFromProxy(key)
	})
	if err != nil {
		c.option.logger.Error("starling: first fetch key %s err=%v", key, err)
		return defaultVal
	}
	got := p.(*ParsedData)
	got.Atime = now
	c.data.Store(key, got)
	return got
}

func (c *Client) getFromProxy(key string) (*ParsedData, error) {
	c.option.logger.Debug("starling: retrieve key %s from proxy", key)
	data, err := c.option.proxyer.Retrieve(c.ctx, key, c.option.retryPolicy)
	if err != nil {
		c.option.logger.Warn("starling: retreive data from proxy failed: key=%s, err=%v", key, err)
		return nil, err
	}
	if data == nil || len(data.Value) <= 13 {
		return nil, errors.New("retrieve invalid data or with size less then 13")
	}

	var obj ParsedData
	if err := json.Unmarshal([]byte(data.Value[11:]), &obj); err != nil {
		c.option.metricer.EmitCounter(clientRetrieveErrorMetricsKey, 1, map[string]string{"key": key})
		c.option.logger.Error("starling: parse retrieved data failed: key=%s, err=%v", key, err)
		return nil, err
	}
	obj.Lang = data.Lang
	return &obj, nil
}

func (c *Client) refresher() {
	ticker := time.NewTicker(c.option.refreshInterval)
	defer ticker.Stop()
	doRefresh := func() {
		duration := c.option.cacheDuration
		data := c.Dump()
		for k, v := range data {
			now := time.Now()
			realVal, ok := v.(*ParsedData)
			if !ok {
				c.data.Delete(k)
				continue
			}
			if realVal.Atime.Add(duration).Before(now) {
				c.data.Delete(k)
				continue
			}
			newVal, err := c.getFromProxy(k)
			if err != nil {
				c.option.logger.Info("starling: refresh key %s failed: %v", k, err)
				continue
			}
			newVal.Atime = now
			c.data.Store(k, newVal)
		}
	}
	for {
		select {
		case <-c.ctx.Done():
			c.option.logger.Info("starling: exit background refresher for client=%s:%s", c.project, c.namespace)
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						c.option.logger.Warn("starling: refresher panic for client=%s:%s, %v", c.project, c.namespace, r)
					}
				}()
				doRefresh()
			}()
		}
	}
}
