package proxypass

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"aahframe.work/aah"
	"aahframe.work/aah/ahttp"
	"gorepositree.com/app/data"
	"gorepositree.com/app/models"
)

// ProxyHosts holds the processed host configuration includes redirects and proxy pass.
var ProxyHosts = &hosts{}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Do method performs the reverse proxy based on the header `Host`.
func Do(ctx *aah.Context) {
	host := ProxyHosts.Lookup(ctx.Req.Host)
	if host == nil {
		ctx.Reply().Status(http.StatusBadGateway).Text("502 Bad Gateway")
		return
	}

	if host.HandleRedirect(ctx) {
		return
	}

	// this is proxy request, just write as-is from upstream
	ctx.Reply().Done()
	r := ctx.Req.Unwrap()
	r.Host = "localhost:8080" // TODO remove
	host.ReverseProxy.ServeHTTP(ctx.Res, r)
}

// Load method creates the reverse proxy instances based
// gorepositree store config.
func Load(e *aah.Event) {
	s := data.Store()
	for h, pi := range s.Data.Proxies {
		ProxyHosts.AddReverseProxy(h, pi)

		for _, r := range pi.Redirects {
			ProxyHosts.AddRedirectRule(h, r)
		}
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Host type and methods
//______________________________________________________________________________

// Host struct holds the proxy pass and redirects processor.
type host struct {
	ExactRedirect map[string]*redirectRule
	RegexRedirect []*redirectRule
	ReverseProxy  *httputil.ReverseProxy
}

func (h *host) HandleRedirect(ctx *aah.Context) bool {
	// Exact match
	if rr, found := h.ExactRedirect[ctx.Req.Path]; found {
		ctx.Reply().RedirectWithStatus(rr.ProcessVars(ctx.Req), rr.Code)
		return true
	}

	// Regex match
	rp := ctx.Req.Path
	for _, re := range h.RegexRedirect {
		matches := re.Regex.FindStringSubmatch(rp)
		ml := len(matches)
		if ml > 0 {
			tu := re.ProcessVars(ctx.Req)
			if ml > 1 {
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

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// RedirectRule type and methods
//______________________________________________________________________________

type redirectRule struct {
	StartsWith bool
	EndsWith   bool
	Code       int
	Target     string
	Vars       []string
	RegexVars  []string
	Regex      *regexp.Regexp
}

func (rr *redirectRule) ProcessVars(r *ahttp.Request) string {
	if len(rr.Vars) == 0 {
		return rr.Target
	}

	tu := rr.Target
	for _, v := range rr.Vars {
		switch v {
		case "{request_uri}":
			tu = strings.Replace(tu, v, r.URL().RequestURI(), -1)
		}
	}
	return tu
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Hosts type and methods
//______________________________________________________________________________

type hosts map[string]*host

func (hs hosts) AddReverseProxy(hostname string, pi *models.ProxyInfo) {
	target, err := url.Parse(pi.URL)
	if err != nil {
		aah.AppLog().Errorf("gorepositree/proxy: %v", err)
		return
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

		// Add or Append
		for k, v := range pi.ReqHdr.Add {
			req.Header.Add(k, v)
		}

		// Remove
		for _, k := range pi.ReqHdr.Remove {
			req.Header.Del(k)
		}
	}
	modifyResponse := func(w *http.Response) error {
		// Add or Append
		for k, v := range pi.ResHdr.Add {
			w.Header.Add(k, v)
		}

		// Remove
		for _, k := range pi.ResHdr.Remove {
			w.Header.Del(k)
		}
		return nil
	}

	hs.addHost(hostname).ReverseProxy = &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse}
}

func (hs hosts) AddRedirectRule(hostname string, redirect *models.Redirect) {
	vars := []string{}
	regexVars := []string{}
	tl := len(redirect.Target)
	for i := 0; i < tl; i++ {
		if redirect.Target[i] == '{' {
			j := i
			for i < tl && redirect.Target[i] != '}' {
				i++
			}
			_, err := strconv.ParseInt(redirect.Target[j+1:i], 10, 32)
			if err != nil {
				vars = append(vars, redirect.Target[j:i+1])
			} else {
				regexVars = append(regexVars, redirect.Target[j:i+1])
			}
		}
	}

	h := hs.addHost(hostname)
	rl := &redirectRule{Target: redirect.Target, Code: redirect.Code, Vars: vars, RegexVars: regexVars}

	ml := len(redirect.Match)
	if redirect.Match[0] == '{' && redirect.Match[ml-1] == '}' {
		regex, err := regexp.Compile(redirect.Match[1 : ml-1])
		if err != nil {
			aah.AppLog().Errorf("gorepositree/proxy: redirect config on '%s': %v", redirect.Match, err)
			return
		}
		rl.Regex = regex
		h.RegexRedirect = append(h.RegexRedirect, rl)
	} else {
		h.ExactRedirect[redirect.Match] = rl
	}
}

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
			ExactRedirect: make(map[string]*redirectRule),
			RegexRedirect: make([]*redirectRule, 0),
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
