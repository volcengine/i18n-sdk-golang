package i18n

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Fetcher abstracts the procedure to retrieve the text data from the remote
// server of the given key. Any communication protocol, such as HTTP, RPC or
// even raw TCP/UDP be used only if it can connect to the starling server which
// supports the protocol. Currently only support HTTP protocol.
type Fetcher interface {
	// Fetch performs the fetching procedure of the given project/namespace/lang and optional params.
	Fetch(ctx context.Context, projectID, namespaceID int64, lang string, opts ...Option) (*Package, error)
	// FetchVersion gets the timestamp version of the given text package with given language.
	FetchVersion(ctx context.Context, projectID, namespaceID int64, lang string, opts ...Option) (int64, string, error)
}

// Package is the data structure which is parsed from the returned value.
type Package struct {
	Version        string            `json:"version"`
	ReleaseVersion string            `json:"release_version"`
	Data           map[string]string `json:"data"`
	Language       string            `json:"language"`

	atime       *time.Time `json:"-"`
	projectID   int64      `json:"-"`
	namespaceID int64      `json:"-"`
	env         string     `json:"-"`
}

type httpFetcher struct {
	httpClient *http.Client
	option     *option
}

// NewHttpProxy creates a `Fetcher` which uses the http protocol to retrieve data
// from the starling server.
func NewHttpFetcher(opts ...Option) *httpFetcher {
	o := &option{enableHTTPs: false}
	for i := range opts {
		opts[i](o)
	}
	if len(o.httpDomain) == 0 {
		o.httpDomain = Domain
	}
	if o.httpTimeout <= 0 {
		o.httpTimeout = 10
	}

	timeout := time.Second * time.Duration(o.httpTimeout)
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 20 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
	}
	return &httpFetcher{httpClient: client, option: o}
}

// Fetch implements the `Fetcher` interface to get the data from server.
func (h *httpFetcher) Fetch(ctx context.Context, pid, nid int64, lang string, opts ...Option) (pkg *Package, err error) {
	opt := op.get()
	defer op.put(opt)
	for _, f := range opts {
		f(opt)
	}

	// Send http request with retry, record the metrics and parse the response.
	doRequest := func(req *http.Request) (*http.Response, error) {
		begin := time.Now()
		resp, err := h.httpClient.Do(req)
		if h.option.metricer != nil {
			tag := map[string]string{"status": "success"}
			if err != nil {
				tag["status"] = "failed"
			}
			elapsed := time.Now().Sub(begin)
			h.option.metricer.EmitCounter(httpProxyMetricsKeyThroughput, 1, tag)
			h.option.metricer.EmitCounter(httpProxyMetricsKeyLatency, elapsed.Milliseconds(), tag)
		}
		return resp, err
	}

	// Perform requesting with retry policy.
	var req *http.Request
	var resp *http.Response
	retryTimes := 0
	retry := h.option.retryPolicy
	if retry == nil {
		retry = opt.retryPolicy
	}
	for {
		// Build HTTP request with the given params from primary storage.
		req, err = h.buildHTTPRequest(pid, nid, lang, false, false, opt)
		if err != nil {
			return
		}
		resp, err = doRequest(req)
		if h.option.logger != nil {
			h.option.logger.Debug("do http request: req=%v, resp=%v, err=%v", req, resp, err)
		}
		if err == nil {
			break
		}

		// Build HTTP request with the given params from backup storage if not disabled.
		if !opt.disableBackupStorage {
			req, err = h.buildHTTPRequest(pid, nid, lang, true, false, opt)
			if err != nil {
				return
			}
			resp, err = doRequest(req)
			if h.option.logger != nil {
				h.option.logger.Debug("do http request: req=%v, resp=%v, err=%v", req, resp, err)
			}
			if err == nil {
				break
			}
		}

		retryTimes++
		if retry != nil && retry.ShouldRetry(retryTimes, err) {
			time.Sleep(retry.RetryDelay(retryTimes))
			if h.option.logger != nil {
				h.option.logger.Info("retry request %d times for err=%v", retryTimes, err)
			}
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gr, e := gzip.NewReader(resp.Body)
	if e != nil {
		return nil, e
	}
	var obj Package
	if err := json.NewDecoder(gr).Decode(&obj); err != nil && err != io.EOF {
		return nil, err
	}
	return &obj, nil
}

// FetchVersion implements the `Fetcher` interface to get the version from server.
func (h *httpFetcher) FetchVersion(ctx context.Context, pid, nid int64, lang string, opts ...Option) (ver int64, rel string, err error) {
	opt := op.get()
	defer op.put(opt)
	for _, f := range opts {
		f(opt)
	}

	// Send http request with retry, record the metrics and parse the response.
	doRequest := func(req *http.Request) (*http.Response, error) {
		begin := time.Now()
		resp, err := h.httpClient.Do(req)
		if h.option.metricer != nil {
			tag := map[string]string{"status": "success"}
			if err != nil {
				tag["status"] = "failed"
			}
			elapsed := time.Now().Sub(begin)
			h.option.metricer.EmitCounter(httpProxyMetricsKeyThroughput, 1, tag)
			h.option.metricer.EmitCounter(httpProxyMetricsKeyLatency, elapsed.Milliseconds(), tag)
		}
		return resp, err
	}

	// Perform requesting with retry policy.
	var req *http.Request
	var resp *http.Response
	retryTimes := 0
	retry := h.option.retryPolicy
	if retry == nil {
		retry = opt.retryPolicy
	}
	for {
		// Build HTTP request with the given params from primary storage.
		req, err = h.buildHTTPRequest(pid, nid, lang, false, true, opt)
		if err != nil {
			return
		}
		resp, err = doRequest(req)
		if h.option.logger != nil {
			h.option.logger.Debug("do http request: req=%v, resp=%v, err=%v", req, resp, err)
		}
		if err == nil {
			break
		}

		// Build HTTP request with the given params from backup storage if not disabled.
		if !opt.disableBackupStorage {
			req, err = h.buildHTTPRequest(pid, nid, lang, true, true, opt)
			if err != nil {
				return
			}
			resp, err = doRequest(req)
			if h.option.logger != nil {
				h.option.logger.Debug("do http request: req=%v, resp=%v, err=%v", req, resp, err)
			}
			if err == nil {
				break
			}
		}

		retryTimes++
		if retry != nil && retry.ShouldRetry(retryTimes, err) {
			time.Sleep(retry.RetryDelay(retryTimes))
			if h.option.logger != nil {
				h.option.logger.Info("retry request %d times for err=%v", retryTimes, err)
			}
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			continue
		}
		break
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result struct {
		Status int    `json:"status"`
		Data   string `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}
	if result.Status != 0 {
		err = errors.New(result.Data)
		return
	}
	ver, _ = strconv.ParseInt(opt.version, 10, 64)
	rel = result.Data
	return
}

func (h *httpFetcher) buildHTTPRequest(pid, nid int64, lang string, useBackupStorage, onlyVersion bool, opt *option) (*http.Request, error) {
	// Build http request and set the custom header.
	var path string
	pidStr, nidStr := strconv.FormatInt(pid, 10), strconv.FormatInt(nid, 10)
	if onlyVersion {
		path = "/api/v4/version/" + pidStr + "/" + nidStr + "/" + lang + "/"
	} else {
		path = "/api/v4/package/" + pidStr + "/" + nidStr + "/" + lang + "/"
	}
	reqUrl := url.URL{
		Scheme: "http",
		Host:   h.option.httpDomain,
		Path:   path,
	}
	if h.option.enableHTTPs || opt.enableHTTPs {
		reqUrl.Scheme = "https"
	}
	if len(opt.httpDomain) != 0 {
		reqUrl.Host = opt.httpDomain
	}

	// Format the optional request query params.
	param := url.Values{}
	param.Add("env", opt.env)
	if len(opt.version) != 0 {
		param.Add("version", opt.version)
	}
	if !opt.disableBackupLang { // default use backup language except disabled explicitly
		param.Add("use_backup_lang", "true")
	}
	if len(opt.backupLang) != 0 {
		param.Add("backup_lang", strings.Join(opt.backupLang, "|"))
	}
	if useBackupStorage {
		param.Add("use_backup_storage", "true")
	}
	reqUrl.RawQuery = param.Encode()
	if h.option.logger != nil {
		h.option.logger.Info("prepare sending http request: %s", reqUrl.String())
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Host-IP", LocalIP)
	key := h.option.appKey
	if len(key) == 0 {
		key = opt.appKey
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("no app-key given")
	}
	oper := h.option.operator
	if len(oper) == 0 {
		oper = opt.operator
	}
	token := CreateAuthToken(pid, nid, key, oper)
	req.Header.Add("Authorization", token)
	return req, nil
}
