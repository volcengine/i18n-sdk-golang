package starling_goclient_public

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Proxyer abstracts the procedure to retrieve the text data from the starling
// server of the given key. Any communication protocol, such as HTTP, RPC or
// even raw TCP/UDP be used only if it can connect to the starling server which
// supports the protocol. Currently starling server only support HTTP protocol.
type Proxyer interface {
	// Retrieve does the operation to retrieve the text data of the given key.
	Retrieve(ctx context.Context, key string, rp RetryPolicy) (*Data, error)
}

// Data defines the data structure which is retrieved from the starling server.
type Data struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Lang  string `json:"lang"`
}

// ParsedData is the data structure which is parsed from the returned value.
type ParsedData struct {
	Version int64             `json:"version"`
	Data    map[string]string `json:"data"`
	Lang    string            `json:"lang"`
	Atime   time.Time         `json:"-"`
}

type httpProxy struct {
	httpClient *http.Client
	option     *option
}

// NewHttpProxy creates a `Proxy` which uses the http protocol to retrieve data
// from the starling server.
func NewHttpProxy(opts ...Option) Proxyer {
	o := &option{enableHTTPs: true}
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
	return &httpProxy{httpClient: client, option: o}
}

// Retrieve implements the `Proxyer` interface to get the data from starling server.
func (h *httpProxy) Retrieve(ctx context.Context, key string, rp RetryPolicy) (*Data, error) {
	if len(key) == 0 {
		return nil, errors.New("key not specified")
	}
	keyArr := strings.Split(strings.Trim(key, "/"), "/")
	if len(keyArr) < 5 {
		return nil, fmt.Errorf("key format error, key=%s", key)
	}
	apiVersion := keyArr[0]

	var uri string
	switch apiVersion {
	case apiVersionGetPackage:
		uri = "/v2/get_pack_without_fallback/"
	case apiVersionSimilarFallback:
		uri = "/v2/get_pack/"
	case apiVersionGetPackageVersion:
		uri = "/v3/get_pack_version/"
	default:
		uri = "/v2/get_pack_without_fallback/"
	}
	uri = uri + strings.Join(keyArr[1:], "/") + "/"
	reqUrl := url.URL{
		Scheme: "http",
		Host:   h.option.httpDomain,
		Path:   uri,
	}
	if h.option.enableHTTPs {
		reqUrl.Scheme = "https"
	}

	// Build http request and set the custom header.
	if h.option.logger != nil {
		h.option.logger.Info("prepare sending http request: %s", reqUrl.String())
	}
	req, err := http.NewRequest(http.MethodPost, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Host-IP", LocalIP)
	token := CreateAuthToken(h.option.appKey, keyArr[2], h.option.operator)
	req.Header.Add("Authorization", token)

	// Send http request with retry, record the metrics and parse the response.
	doRequest := func() (*http.Response, error) {
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
	var resp *http.Response
	retryTimes := 0
	for {
		if resp, err = doRequest(); err == nil {
			break
		}
		retryTimes++
		if rp != nil && rp.ShouldRetry(retryTimes, err) {
			time.Sleep(rp.RetryDelay(retryTimes))
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
	if h.option.logger != nil {
		h.option.logger.Debug("do http request: req=%v, resp=%v, err=%v", req, resp, err)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Code    int    `json:"error_code"`
		Message string `json:"message"`
		Cause   string `json:"cause"`
		Data    Data   `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Message) != 0 {
		return nil, errors.New(result.Message)
	}
	return &result.Data, nil
}
