// Copyright Jeevanandam M. (https://github.com/jeevatkm, jeeva@myjeeva.com)
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
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"thumbai/app/settings"

	"thumbai/app/models"

	"aahframe.work"
	"aahframe.work/ahttp"
	"aahframe.work/essentials"
)

// Thumbai proxies instance.
var Thumbai *proxies

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Load method reads proxy configurations from store and builds proxy
// engine.
func Load(_ *aah.Event) {
	Thumbai = &proxies{RWMutex: sync.RWMutex{}, Hosts: make(map[string]*host)}
	allProxies := All()
	if allProxies == nil || len(allProxies) == 0 {
		aah.AppLog().Info("Proxies are not yet configured on THUMBAI")
		return
	}

	for h, rules := range allProxies {
		host := Thumbai.AddHost(h)
		for _, r := range rules {
			if err := host.AddProxyRule(r); err != nil {
				aah.AppLog().Error(err)
			}
		}
		if host.LastRule == nil {
			if len(host.ProxyRules) == 1 {
				host.Lock()
				host.LastRule = host.ProxyRules[0]
				host.ProxyRules = nil
				host.Unlock()
			} else {
				aah.AppLog().Errorf("Incomplete proxy configuration for host->%s; last rule not found, reverse proxy may not work properly", host.Name)
			}
		}
	}

	aah.AppLog().Info("Successfully created reverse proxy engine")
}

// Do method performs the reverse proxy based on the header `Host` and proxy rules.
func Do(ctx *aah.Context) {
	host := Thumbai.Lookup(ctx.Req.Host)
	if host == nil {
		ctx.Reply().Status(http.StatusBadGateway).Text("502 Bad Gateway")
		return
	}
	host.RLock()
	defer host.RUnlock()

	var tr *rule
	for _, r := range host.ProxyRules {
		if len(r.Path) > 0 || r.PathRegex != nil {
			var rpath bool
			if r.PathRegex == nil {
				rpath = ctx.Req.Path == r.Path
			} else {
				rpath = r.PathRegex.MatchString(ctx.Req.Path)
			}
			if !rpath {
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

	tr.RLock()
	defer tr.RUnlock()

	// Restrict by file extensions and regex
	if tr.RestrictFile != nil {
		file := path.Base(ctx.Req.Path)
		ext := strings.ToLower(path.Ext(file))
		for _, e := range tr.RestrictFile.Extensions {
			if ext == e {
				ctx.Reply().Forbidden().Text("403 Forbidden")
				return
			}
		}
		for _, re := range tr.RestrictFile.Regexs {
			if re.MatchString(file) {
				ctx.Reply().Forbidden().Text("403 Forbidden")
				return
			}
		}
	}

	// Redirects
	if tr.checkRedirects(ctx) {
		return
	}

	// Static file try from filesystem
	if ctx.Req.Path != "/" {
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
	}

	ctx.Reply().Done()
	if len(settings.ServerHeader) > 0 {
		ctx.Res.Header().Set(ahttp.HeaderServer, settings.ServerHeader)
	}
	tr.Proxy.ServeHTTP(ctx.Res, ctx.Req.Unwrap())
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Proxies type and its methods
//______________________________________________________________________________

type proxies struct {
	sync.RWMutex
	Hosts map[string]*host
}

func (p *proxies) Lookup(hostname string) *host {
	p.RLock()
	defer p.RUnlock()
	if h, f := p.Hosts[strings.ToLower(hostname)]; f {
		return h
	}
	return nil
}

func (p *proxies) AddHost(hostname string) *host {
	h := p.Lookup(hostname)
	if h == nil {
		h = &host{
			RWMutex:    sync.RWMutex{},
			Name:       hostname,
			ProxyRules: make([]*rule, 0),
		}
		p.Lock()
		p.Hosts[strings.ToLower(hostname)] = h
		p.Unlock()
	}
	return h
}

func (p *proxies) DelHost(hostname string) {
	p.Lock()
	delete(p.Hosts, hostname)
	p.Unlock()
}

func (p *proxies) UpdateRule(targetURL string, pr *models.ProxyRule) error {
	if h := p.Lookup(pr.Host); h != nil {
		return h.UpdateProxyRule(targetURL, pr)
	}
	return nil
}

func (p *proxies) DeleteRule(hostname, targetURL string) {
	if h := p.Lookup(hostname); h != nil {
		h.DelProxyRule(targetURL)
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Host type and methods
//______________________________________________________________________________

// Host struct holds the proxy pass and redirects processor.
type host struct {
	sync.RWMutex
	Name       string
	LastRule   *rule
	ProxyRules []*rule
}

type restrictFile struct {
	Extensions []string
	Regexs     []*regexp.Regexp
}

func (h *host) AddProxyRule(pr *models.ProxyRule) error {
	r, err := h.createProxyRule(pr)
	if err != nil {
		return err
	}
	h.Lock()
	if pr.Last { // set last rule
		h.LastRule = r
	} else {
		h.ProxyRules = append(h.ProxyRules, r)
	}
	h.Unlock()

	return nil
}

func (h *host) UpdateProxyRule(targetURL string, pr *models.ProxyRule) error {
	existingRule, i := h.LookupRule(targetURL)
	if existingRule == nil { // no rule found
		return errors.New("proxy rule not found")
	}
	newRule, err := h.createProxyRule(pr)
	if err != nil {
		return err
	}
	h.Lock()
	if i == -1 {
		h.LastRule = newRule
	} else {
		h.ProxyRules[i] = newRule
	}
	h.Unlock()
	return nil
}

func (h *host) DelProxyRule(targetURL string) {
	existingRule, i := h.LookupRule(targetURL)
	if existingRule != nil {
		h.Lock()
		h.ProxyRules = append(h.ProxyRules[:i], h.ProxyRules[i+1:]...)
		h.Unlock()
	}
}

func (h *host) LookupRule(targetURL string) (*rule, int) {
	h.RLock()
	defer h.RUnlock()
	for i, r := range h.ProxyRules {
		if r.TargetURL == targetURL {
			return r, i
		}
	}
	if h.LastRule != nil && h.LastRule.TargetURL == targetURL {
		return h.LastRule, -1
	}
	return nil, -1
}

func (h *host) createProxyRule(pr *models.ProxyRule) (*rule, error) {
	r := &rule{RWMutex: sync.RWMutex{}, TargetURL: pr.TargetURL, host: h}
	r.QueryParams = pr.QueryParams
	r.Headers = pr.Headers

	pl := len(pr.Path)
	if pl > 0 && pr.Path[0] == '{' && pr.Path[pl-1] == '}' {
		regex, err := regexp.Compile(pr.Path[1 : pl-1])
		if err != nil {
			return nil, fmt.Errorf("proxy path config error on host->'%s' match->'%s': %v", h.Name, pr.Path, err)
		}
		r.PathRegex = regex
	} else {
		r.Path = pr.Path
	}

	if len(pr.Redirects) > 0 {
		r.ExactRedirect = make(map[string]*redirectRule)
		r.RegexRedirect = make([]*redirectRule, 0)
		for _, redirect := range pr.Redirects {
			if err := r.AddRedirect(redirect); err != nil {
				aah.AppLog().Error(err)
			}
		}
	}

	r.ReqHdr = pr.RequestHeaders
	r.ResHdr = pr.ResponseHeaders

	if pr.RestrictFiles != nil {
		r.RestrictFile = &restrictFile{}
		if len(pr.RestrictFiles.Extensions) > 0 {
			r.RestrictFile.Extensions = pr.RestrictFiles.Extensions
		}
		if len(pr.RestrictFiles.Regexs) > 0 {
			for _, rr := range pr.RestrictFiles.Regexs {
				regex, err := regexp.Compile(rr[1 : len(rr)-1])
				if err != nil {
					return nil, fmt.Errorf("proxy restrict by regex config has an error on host='%s' regex='%s': %v", h.Name, rr, err)
				}
				r.RestrictFile.Regexs = append(r.RestrictFile.Regexs, regex)
			}
		}
	}

	r.Statics = pr.Statics

	if err := r.createReverseProxy(pr.TargetURL, pr.SkipTLSVerify); err != nil {
		return nil, err
	}
	return r, nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Proxy Rule type and its methods
//______________________________________________________________________________

type rule struct {
	sync.RWMutex
	TargetURL     string
	Path          string
	PathRegex     *regexp.Regexp
	QueryParams   map[string]string
	Headers       map[string]string
	ExactRedirect map[string]*redirectRule
	RegexRedirect []*redirectRule
	RestrictFile  *restrictFile
	Statics       []*models.ProxyStatic
	ReqHdr        *models.ProxyHeader
	ResHdr        *models.ProxyHeader
	Proxy         *httputil.ReverseProxy
	host          *host
}

func (r *rule) EditConditions(pr *models.ProxyRule) error {
	r.Lock()
	defer r.Unlock()
	r.QueryParams = pr.QueryParams
	r.Headers = pr.Headers
	pl := len(pr.Path)
	if pl > 0 && pr.Path[0] == '{' && pr.Path[pl-1] == '}' {
		regex, err := regexp.Compile(pr.Path[1 : pl-1])
		if err != nil {
			return fmt.Errorf("proxy path config error on host->'%s' match->'%s': %v", r.host.Name, pr.Path, err)
		}
		r.PathRegex = regex
	} else {
		r.Path = pr.Path
	}
	return nil
}

func (r *rule) AddRedirect(redirect *models.ProxyRedirect) error {
	vars := make([]string, 0)
	regexVars := make([]string, 0)
	tl := len(redirect.Target)
	for i := 0; i < tl; i++ {
		if redirect.Target[i] == '{' {
			j := i
			for i < tl && redirect.Target[i] != '}' {
				i++
			}
			if _, err := strconv.ParseInt(redirect.Target[j+1:i], 10, 32); err != nil {
				vars = append(vars, redirect.Target[j:i+1])
				continue
			}
			regexVars = append(regexVars, redirect.Target[j:i+1])
		}
	}

	rl := &redirectRule{IsAbs: redirect.IsAbs, Target: redirect.Target,
		Code: redirect.Code, Vars: vars, RegexVars: regexVars}
	if rl.Code == 0 {
		rl.Code = http.StatusMovedPermanently
	}
	ml := len(redirect.Match)
	if redirect.Match[0] == '{' && redirect.Match[ml-1] == '}' {
		regex, err := regexp.Compile(redirect.Match[1 : ml-1])
		if err != nil {
			return fmt.Errorf("redirect config error on host->'%s' match->'%s': %v", r.host.Name, redirect.Match, err)
		}
		rl.Regex = regex
		r.RegexRedirect = append(r.RegexRedirect, rl)
	} else {
		r.ExactRedirect[redirect.Match] = rl
	}
	return nil
}

func (r *rule) checkRedirects(ctx *aah.Context) bool {
	// Exact match
	if rr, found := r.ExactRedirect[ctx.Req.Path]; found {
		ctx.Reply().RedirectWithStatus(rr.ProcessVars(ctx.Req), rr.Code)
		return true
	}
	// Regex match
	rp := ctx.Req.Path
	for _, re := range r.RegexRedirect {
		matches := re.Regex.FindStringSubmatch(rp)
		ml := len(matches)
		if ml > 0 {
			tu := re.ProcessVars(ctx.Req)
			if len(re.RegexVars) > 0 && ml > 1 {
				for i, v := range matches[1:] {
					tu = strings.Replace(tu, re.RegexVars[i], v, 1)
				}
			}
			ctx.Reply().RedirectWithStatus(tu, re.Code)
			return true
		}
	}
	return false
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

		if _, ok := req.Header[ahttp.HeaderUserAgent]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set(ahttp.HeaderUserAgent, "")
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

	// for now use default transport
	// later we can enhance it more options
	transport := http.DefaultTransport.(*http.Transport)
	if skipTLSVerify {
		// #nosec
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	r.Proxy = &httputil.ReverseProxy{Director: director, Transport: transport, ModifyResponse: modifyResponse}

	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Rule type and its methods
//______________________________________________________________________________

type redirectRule struct {
	IsAbs     bool
	Code      int
	Target    string
	Vars      []string
	RegexVars []string
	Regex     *regexp.Regexp
}

func (r *redirectRule) ProcessVars(req *ahttp.Request) string {
	tu := r.Target
	if !r.IsAbs {
		tu = req.Scheme + "://" + req.Host + r.Target
	}
	if len(r.Vars) == 0 {
		return tu
	}
	for _, v := range r.Vars {
		switch v {
		case "{request_uri}":
			tu = strings.Replace(tu, v, req.URL().RequestURI(), -1)
		}
	}
	return tu
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
