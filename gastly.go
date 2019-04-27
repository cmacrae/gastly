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
	"time"
)

const ghostAPI = "https://ghostproxies.com/proxies/api.json"

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

// RandProxy returns a random proxy from a Provider's list of proxies
func (p Provider) RandProxy() Proxy {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	rand := r.Intn(len(p.Data))

	return p.Data[rand].Proxy
}

// NewClient returns a http.Client configured to use a random proxy
func (p Provider) NewClient(req *http.Request) (*http.Client, error) {
	proxy := p.RandProxy()
	proxyURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%s", proxy.IP, proxy.PortNum))
	if err != nil {
		return &http.Client{}, fmt.Errorf("%v", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy:              http.ProxyURL(proxyURL),
			ProxyConnectHeader: req.Header,
		},
		Timeout: (5 * time.Second),
	}, nil
}

// Get performs an HTTP GET request against the given url, with any headers provided.
// It will use a random proxy to do so
func (p Provider) Get(url string, headers map[string]string) (http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.Response{}, err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Set the client to use the proxied internal client
	client, err := p.NewClient(req)
	if err != nil {
		return http.Response{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return http.Response{}, err
	}

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
