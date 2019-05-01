// gastly provides general purpose HTTP functionality via GhostProxies

// Copyright 2019 Calum MacRae. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package gastly

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const ghostAPI = "https://ghostproxies.com/proxies/api.json"

// Prometheus metrics
var httpReqs = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "gastly_external_http_requests_total",
		Help: "How many external HTTP requests processed, partitioned by status code, method and proxy IP.",
	},
	[]string{"code", "method", "proxy_ip"},
)

// Provider set of account proxy data returned from the GhostProxies API
type Provider struct {
	Data []Container `json:"data"`
}

// Container is a helper set for individual proxies
type Container struct {
	Proxy Proxy `json:"Proxy"`
}

// Proxy is a set of data representing HTTP proxies retrieved from GhostProxies
type Proxy struct {
	IP          string `json:"ip"`
	Status      string `json:"status"`
	PortNum     string `json:"portNum"`
	CityName    string `json:"cityName"`
	RegionName  string `json:"regionName"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	PanelUser   string `json:"panel_user"`
	PanelPass   string `json:"panel_pass"`
}

// RetryOptions is a set of parameters expressing HTTP retry behavior
type RetryOptions struct {
	Max             int
	WaitMinSecs     int
	WaitMaxSecs     int
	BackoffStepSecs int
}

// RandProxy returns a random proxy from a Provider's list of proxies
func (p Provider) RandProxy() Proxy {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	rand := r.Intn(len(p.Data))

	return p.Data[rand].Proxy
}

// NewClient returns a retryablehttp.Client configured to use a random proxy
func (p Provider) NewClient(req *retryablehttp.Request, opts RetryOptions) (*retryablehttp.Client, string, error) {
	proxy := p.RandProxy()
	proxyURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%s", proxy.IP, proxy.PortNum))
	if err != nil {
		return &retryablehttp.Client{}, "", fmt.Errorf("%v", err)
	}

	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = opts.Max
	client.RetryWaitMax = time.Second * time.Duration(opts.WaitMaxSecs)
	client.RetryWaitMin = time.Second * time.Duration(opts.WaitMinSecs)

	client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return (time.Second * time.Duration(opts.BackoffStepSecs)) * time.Duration((attemptNum))
	}

	client.HTTPClient = &http.Client{
		Timeout: (5 * time.Second),
		Transport: &http.Transport{
			Proxy:              http.ProxyURL(proxyURL),
			ProxyConnectHeader: req.Header,
		}}

	return client, proxy.IP, nil
}

// Get performs an HTTP GET request against the given url, with any headers and retry options provided.
// It will use a random proxy to do so
func (p Provider) Get(url string, header http.Header, o RetryOptions) (http.Response, error) {
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return http.Response{}, err
	}

	req.Header = header

	client, proxyIP, err := p.NewClient(req, o)
	if err != nil {
		return http.Response{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return http.Response{}, err
	}

	httpReqs.WithLabelValues(strconv.Itoa(resp.StatusCode), "GET", proxyIP).Inc()

	return *resp, nil
}

// NewProvider returns a configured Provider
func NewProvider(key string) (Provider, error) {
	if key == "" {
		return Provider{}, fmt.Errorf("empty API key")
	}

	p := Provider{}
	client := &http.Client{Timeout: 10 * time.Second}
	r, err := client.Get(ghostAPI + "?key=" + key)
	if err != nil {
		return Provider{}, err
	}
	defer r.Body.Close()
	json.NewDecoder(r.Body).Decode(&p)
	return p, nil
}

// ServeMetrics provides a Prometheus endpoint for monitoring/observability
func ServeMetrics(port int) {
	prometheus.MustRegister(httpReqs)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
