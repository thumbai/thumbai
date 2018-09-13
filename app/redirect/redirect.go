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

package redirect

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"thumbai/app/models"

	"aahframe.work/aah"
	"aahframe.work/aah/ahttp"
)

var redirectHosts hosts

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Load method reads redirects configurations from store and builds redirects
// engine.
func Load(_ *aah.Event) {
	redirects := models.AllRedirects()
	if redirects == nil || len(redirects) == 0 {
		aah.AppLog().Info("Redirects are not yet configured on THUMBAI")
		return
	}

	redirectHosts = hosts{}
	for h, rules := range redirects {
		host := redirectHosts.addHost(h)
		for _, r := range rules {
			if err := host.AddRedirect(r); err != nil {
				aah.AppLog().Error(err)
			}
		}
	}

	aah.AppLog().Info("Successfully created redirect engine")
}

// Do method perform redirect check for incoming request. If
// redirect rule satisfy rule, it does redirect request and returns true
// otherwise return false.
func Do(ctx *aah.Context) bool {
	h := redirectHosts.Lookup(ctx.Req.Host)
	if h == nil {
		return false
	}

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

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Host type and its methods
//______________________________________________________________________________

// Host struct holds redirects rules.
type host struct {
	Name          string
	ExactRedirect map[string]*rule
	RegexRedirect []*rule
}

func (h *host) AddRedirect(redirect *models.Redirect) error {
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

	rl := &rule{Target: redirect.Target, Code: redirect.Code, Vars: vars, RegexVars: regexVars}
	ml := len(redirect.Match)
	if redirect.Match[0] == '{' && redirect.Match[ml-1] == '}' {
		regex, err := regexp.Compile(redirect.Match[1 : ml-1])
		if err != nil {
			return fmt.Errorf("redirect config error on host->'%s' match->'%s': %v", h.Name, redirect.Match, err)
		}
		rl.Regex = regex
		h.RegexRedirect = append(h.RegexRedirect, rl)
	} else {
		h.ExactRedirect[redirect.Match] = rl
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
			Name:          hostname,
			ExactRedirect: make(map[string]*rule),
			RegexRedirect: make([]*rule, 0),
		}
		hs[strings.ToLower(hostname)] = h
	}
	return h
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Rule type and its methods
//______________________________________________________________________________

type rule struct {
	Code      int
	Target    string
	Vars      []string
	RegexVars []string
	Regex     *regexp.Regexp
}

func (r *rule) ProcessVars(req *ahttp.Request) string {
	if len(r.Vars) == 0 {
		return r.Target
	}

	tu := r.Target
	for _, v := range r.Vars {
		switch v {
		case "{request_uri}":
			tu = strings.Replace(tu, v, req.URL().RequestURI(), -1)
		}
	}
	return tu
}
