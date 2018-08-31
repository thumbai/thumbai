// Copyright 2018 Jeevanandam M. (https://github.com/jeevatkm, jeeva@myjeeva.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"thumbai/app/models"

	"aahframe.work/aah"
	"aahframe.work/aah/essentials"
)

var proxyHosts hosts

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Load method reads proxy configurations from store and builds proxy
// engine.
func Load(_ *aah.Event) {
	proxies := models.AllProxies()
	if proxies == nil || len(proxies) == 0 {
		aah.AppLog().Info("Proxies are not yet configured on THUMBAI")
		return
	}

	proxyHosts = hosts{}
	for h, rules := range proxies {
		host := proxyHosts.addHost(h)
		for _, r := range rules {
			if err := host.AddProxyRule(r); err != nil {
				aah.AppLog().Error(err)
			}
		}
		if host.LastRule == nil {
			if len(host.ProxyRules) == 1 {
				host.LastRule = host.ProxyRules[0]
				host.ProxyRules = nil
			} else {
				aah.AppLog().Errorf("Incomplete proxy configuration for host->%s, reverse proxy may not work properly", host.Name)
			}
		}
	}

	aah.AppLog().Info("Successfully created reverse proxy engine")
}

// Do method performs the reverse proxy based on the header `Host` and proxy rules.
func Do(ctx *aah.Context) {
	host := proxyHosts.Lookup(ctx.Req.Host)
	if host == nil {
		ctx.Reply().Status(http.StatusBadGateway).Text("502 Bad Gateway")
		return
	}

	var tr *rule
	for _, r := range host.ProxyRules {
		if len(r.Path) > 0 || r.PathRegex != nil {
			path := false
			if r.PathRegex == nil {
				path = ctx.Req.Path == r.Path
			} else {
				path = r.PathRegex.MatchString(ctx.Req.Path)
			}
			if !path {
				continue
			}
		}

		if len(r.QueryParams) > 0 {
			query := true
			for k, v := range r.QueryParams {
				if ctx.Req.QueryValue(k) != v {
					query = false
					break
				}
			}
			if !query {
				continue
			}
		}

		if len(r.Headers) > 0 {
			hdr := true
			for k, v := range r.Headers {
				if ctx.Req.Header.Get(k) != v {
					hdr = false
					break
				}
			}
			if !hdr {
				continue
			}
		}

		tr = r
		break
	}

	if tr == nil {
		tr = host.LastRule
	}
	if tr == nil {
		ctx.Reply().Status(http.StatusBadGateway).Text("502 Bad Gateway")
		return
	}

	for _, sf := range tr.Statics {
		tp := ctx.Req.Path
		if len(sf.StripPrefix) > 0 {
			tp = strings.TrimPrefix(tp, sf.StripPrefix)
		}

		tp = filepath.Join(sf.TargetPath, tp)
		if ess.IsFileExists(tp) {
			ctx.Reply().File(tp)
			return
		}
	}

	ctx.Reply().Done()
	tr.Proxy.ServeHTTP(ctx.Res, ctx.Req.Unwrap())
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Host type and methods
//______________________________________________________________________________

// Host struct holds the proxy pass and redirects processor.
type host struct {
	Name       string
	LastRule   *rule
	ProxyRules []*rule
}

type rule struct {
	Path        string
	PathRegex   *regexp.Regexp
	QueryParams map[string]string
	Headers     map[string]string
	ReqHdr      *models.ProxyHeader
	ResHdr      *models.ProxyHeader
	Statics     []*models.ProxyStatic
	Proxy       *httputil.ReverseProxy
	host        *host
}

func (h *host) AddProxyRule(pr *models.ProxyRule) error {
	r := &rule{
		host:        h,
		QueryParams: pr.QueryParams,
		Headers:     pr.Headers,
	}

	pl := len(pr.Path)
	if pl > 0 && pr.Path[0] == '{' && pr.Path[pl-1] == '}' {
		regex, err := regexp.Compile(pr.Path[1 : pl-1])
		if err != nil {
			return fmt.Errorf("proxy path config error on host->'%s' match->'%s': %v", h.Name, pr.Path, err)
		}
		r.PathRegex = regex
	} else {
		r.Path = pr.Path
	}

	if pr.RequestHeader != nil {
		reqHdr := *pr.RequestHeader
		r.ReqHdr = &reqHdr
	}

	if pr.ResponseHeader != nil {
		resHdr := *pr.ResponseHeader
		r.ResHdr = &resHdr
	}

	if len(pr.Statics) > 0 {
		for _, v := range pr.Statics {
			t := *v
			r.Statics = append(r.Statics, &t)
		}
	}

	if err := r.createReverseProxy(pr.TargetURL, pr.SkipTLSVerify); err != nil {
		return err
	}

	if pr.Last { // set last rule
		h.LastRule = r
	} else {
		h.ProxyRules = append(h.ProxyRules, r)
	}

	return nil
}

func (r *rule) createReverseProxy(targetURL string, skipTLSVerify bool) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("proxy target URL error on host->'%s' match->'%s': %v", r.host.Name, targetURL, err)
	}

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		if r.ReqHdr != nil {
			for k, v := range r.ReqHdr.Add {
				req.Header.Add(k, v)
			}
			for _, k := range r.ReqHdr.Remove {
				req.Header.Del(k)
			}
		}
	}
	modifyResponse := func(w *http.Response) error {
		if r.ResHdr != nil {
			for k, v := range r.ResHdr.Add {
				w.Header.Add(k, v)
			}
			for _, k := range r.ResHdr.Remove {
				w.Header.Del(k)
			}
		}
		return nil
	}

	if skipTLSVerify {
		r.Proxy = &httputil.ReverseProxy{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			Director: director, ModifyResponse: modifyResponse}
	} else {
		r.Proxy = &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse}
	}

	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Hosts type and its methods
//______________________________________________________________________________

type hosts map[string]*host

func (hs hosts) Lookup(hostname string) *host {
	if h, f := hs[strings.ToLower(hostname)]; f {
		return h
	}
	return nil
}

func (hs hosts) addHost(hostname string) *host {
	h := hs.Lookup(hostname)
	if h == nil {
		h = &host{
			Name:       hostname,
			ProxyRules: make([]*rule, 0),
		}
		hs[strings.ToLower(hostname)] = h
	}
	return h
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Unexported methods
//______________________________________________________________________________

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
