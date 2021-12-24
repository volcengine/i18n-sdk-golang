package i18n

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// client provides all functionality about retrieving data from the remote
// server of the given project and namespace. It will cache the data and
// fetch the data in background to avoid the performance overhead.
type client struct {
	projectID   int64
	namespaceID int64
	options     []Option
	mu          sync.RWMutex
	data        sync.Map
	sf          Group
	shutdownCh  chan struct{}
}

// GetPackage returns a whole package of the given language.
func (c *client) GetPackage(ctx context.Context, lang string, opts ...Option) (*Package, error) {
	o := op.get()
	defer op.put(o)
	return c.getPackage(ctx, o, lang, opts...)
}

// GetText retrieves the text string of the given key in a given language i18n package.
func (c *client) GetText(ctx context.Context, lang, key string, opts ...Option) (val string, err error) {
	o := op.get()
	defer op.put(o)
	pkg, pkgErr := c.getPackage(ctx, o, lang, opts...)
	if pkgErr != nil {
		err = pkgErr
		return
	}
	raw, ok := pkg.Data[key]
	if !ok {
		o.logger.Warn("[starling-client-go] text not existed: key=%v", key)
		o.metricer.EmitCounter(clientKeyEmptyMetricsKey, 1, map[string]string{
			"projectID":   strconv.FormatInt(o.projectID, 10),
			"namespaceID": strconv.FormatInt(o.namespaceID, 10),
			"language":    lang,
			"env":         o.env,
			"key":         key,
		})
		err = ErrKeyNotExist
		return
	}
	val = raw
	if o.pluralCount != nil {
		val, err = c.processPlural(raw, lang, o.pluralDefaultLang, o.pluralCount)
	}
	if len(o.arguments) != 0 {
		val, err = c.processVars(val, o.arguments, o.leftDelimiter, o.rightDelimiter)
	}
	return
}

// AddOption implements the `Client` interface's method.
func (c *client) AddOption(opts ...Option) {
	c.options = append(c.options, opts...)
}

// Shutdown cleans the resources and exits gracefully.
func (c *client) Shutdown() {
	if c.shutdownCh != nil {
		close(c.shutdownCh)
		c.shutdownCh = nil
	}
}

func (c *client) getPackage(ctx context.Context, o *option, lang string, opts ...Option) (data *Package, err error) {
	defer func() {
		if err != nil {
			o.metricer.EmitCounter(clientPackageEmptyMetricsKey, 1, map[string]string{"error": err.Error()})
		}
	}()
	var optArr []Option
	optArr, err = c.handleOptions(o, lang, opts...)
	if err != nil {
		o.logger.Warn("[starling-client-go] handle options failed: %v %v", o, err)
		return
	}

	cacheKey := buildCacheKey(o.projectID, o.namespaceID, o.env, o.language)
	now := time.Now()
	if val, exist := c.data.Load(cacheKey); exist {
		realVal, ok := val.(*Package)
		if ok && (len(o.version) == 0 || realVal.ReleaseVersion == o.version) {
			realVal.atime = &now
			c.data.Store(cacheKey, realVal)
			data = realVal
			return
		}
	}
	var p interface{}
	p, err = c.sf.Do(cacheKey, func() (interface{}, error) {
		return c.getFromProxy(ctx, cacheKey, o, optArr...)
	})
	if err != nil {
		o.logger.Error("starling: first fetch key %s err=%v", cacheKey, err)
		return
	}
	if got, ok := p.(*Package); ok {
		data = got
		data.projectID, data.namespaceID, data.env, data.atime = o.projectID, o.namespaceID, o.env, &now
		if len(o.version) == 0 {
			c.data.Store(cacheKey, data)
		}
	} else {
		err = ErrBackToSourceFailed
	}
	return
}

func (c *client) handleOptions(o *option, lang string, opts ...Option) ([]Option, error) {
	if o == nil { // the object to handle should not be empty
		return nil, ErrInvalidParams
	}
	optArr := make([]Option, len(c.options)+len(opts))
	copy(optArr, c.options)
	copy(optArr[len(c.options):], opts)
	if len(lang) != 0 { // required lang param has highest priority
		optArr = append(optArr, WithLanguage(lang))
	}
	for _, f := range optArr {
		f(o)
	}
	if o.projectID == 0 { // use global project ID if request-level not given
		o.projectID = c.projectID
		optArr = append(optArr, WithProjectID(o.projectID))
	}
	if o.namespaceID == 0 { // use global namespace ID if request-level not given
		o.namespaceID = c.namespaceID
		optArr = append(optArr, WithNamespaceID(o.namespaceID))
	}
	if o.projectID <= 0 || o.namespaceID <= 0 {
		return nil, ErrInvalidParams
	}
	if len(o.env) == 0 {
		o.env = EnvNormal
		optArr = append(optArr, WithEnv(o.env))
	}
	return optArr, nil
}

func (c *client) getFromProxy(ctx context.Context, key string, o *option, opts ...Option) (*Package, error) {
	if o.onlyVersion {
		if len(o.version) == 0 {
			o.logger.Debug("starling: retrieve version from proxy: %v@latest", key)
		} else {
			o.logger.Debug("starling: retrieve version from proxy: %v@%v", key, o.version)
		}
		ver, rel, err := o.fetcher.FetchVersion(ctx, o.projectID, o.namespaceID, o.language, opts...)
		if err != nil {
			o.metricer.EmitCounter(clientRetrieveErrorMetricsKey, 1, map[string]string{"key": key})
			o.logger.Warn("starling: fetch version from proxy failed: key=%s, err=%v", key, err)
			return nil, err
		}
		return &Package{
			Version:        strconv.FormatInt(ver, 10),
			ReleaseVersion: rel,
			Language:       o.language,
		}, nil
	}

	if len(o.version) == 0 {
		o.logger.Debug("starling: retrieve data from proxy: %v@latest", key)
	} else {
		o.logger.Debug("starling: retrieve data from proxy: %v@%v", key, o.version)
	}
	data, err := o.fetcher.Fetch(ctx, o.projectID, o.namespaceID, o.language, opts...)
	if err != nil {
		o.metricer.EmitCounter(clientRetrieveErrorMetricsKey, 1, map[string]string{"key": key})
		o.logger.Warn("starling: fetch data from proxy failed: key=%s, err=%v", key, err)
		return nil, err
	}
	return data, nil
}

func (c *client) processPlural(raw, lang, defLang string, count interface{}) (text string, err error) {
	varName, msg, icuErr := ParseICU(raw)
	if icuErr != nil {
		err = icuErr
		return
	}
	if len(defLang) == 0 {
		defLang = lang
	}
	localize := goi18n.NewLocalizer(goi18n.NewBundle(language.Make(defLang)), lang)
	localized, _ := localize.Localize(&goi18n.LocalizeConfig{
		DefaultMessage: msg,
		PluralCount:    count,
	})
	// Replace the `#` or `{varName}` with the plural count number.
	countStr := fmt.Sprintf("%v", count)
	text = strings.ReplaceAll(localized, "#", countStr)
	if len(varName) != 0 {
		r := regexp.MustCompile(`{\s*` + varName + `\s*}`)
		text = r.ReplaceAllString(text, countStr)
	}
	return
}

func (c *client) processVars(raw string, vars map[string]interface{}, left, right string) (string, error) {
	if len(left) == 0 {
		left = defaultLeftDelimiter
	}
	if len(right) == 0 {
		right = defaultRightDelimiter
	}
	var b bytes.Buffer
	var i int
	for i < len(raw) {
		hasLeft, hasRight := true, true
		tmp := raw[i:]
		pos := strings.Index(tmp, left)
		if pos != -1 {
			b.WriteString(raw[i : i+pos])
			b.WriteString("{{.")
			i += pos + len(left)
		} else {
			hasLeft = false
		}

		tmp = raw[i:]
		pos = strings.Index(tmp, right)
		if pos != -1 {
			b.WriteString(raw[i : i+pos])
			b.WriteString("}}")
			i += pos + len(right)
		} else {
			hasRight = false
		}
		if !hasLeft && !hasRight {
			break
		}
	}
	b.WriteString(raw[i:])
	t := template.Must(template.New("icu").Parse(b.String()))
	var out bytes.Buffer
	if err := t.Execute(&out, vars); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c *client) refresher(ctx context.Context) {
	o := op.get()
	defer op.put(o)
	for _, f := range c.options {
		f(o)
	}
	ticker := time.NewTicker(o.refreshInterval)
	defer ticker.Stop()
	doRefresh := func() {
		duration := o.cacheDuration
		data := make(map[string]interface{})
		c.data.Range(func(key, value interface{}) bool {
			if k, ok := key.(string); ok {
				data[k] = value
			}
			return true
		})
		for k, v := range data {
			now := time.Now()
			realVal, ok := v.(*Package)
			if !ok {
				c.data.Delete(k)
				continue
			}
			if realVal.atime.Add(duration).Before(now) {
				c.data.Delete(k)
				continue
			}

			o.projectID, o.namespaceID, o.env, o.language = realVal.projectID, realVal.namespaceID, realVal.env, realVal.Language
			newVal, err := c.getFromProxy(ctx, k, o,
				WithProjectID(realVal.projectID),
				WithNamespaceID(realVal.namespaceID),
				WithEnv(realVal.env),
				WithLanguage(realVal.Language))
			if err != nil {
				o.logger.Info("starling: refresh key %s failed: %v", k, err)
				continue
			}

			newVal.projectID, newVal.namespaceID, newVal.env, newVal.atime = realVal.projectID, realVal.namespaceID, realVal.env, realVal.atime
			c.data.Store(k, newVal)
		}
	}
	for {
		select {
		case <-c.shutdownCh:
			o.logger.Info("starling: exit background refresher for client=%v:%v", c.projectID, c.namespaceID)
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						o.logger.Warn("starling: refresher panic for client=%v:%v, %v", c.projectID, c.namespaceID, r)
					}
				}()
				doRefresh()
			}()
		}
	}
}
