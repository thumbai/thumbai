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

package models

import (
	"strings"
	"sync"

	"aahframe.work/aah"
)

var proxyStore *ProxyStore

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// LoadProxyStore loads the proxy data store from file system.
func LoadProxyStore(_ *aah.Event) {
	proxyStore = &ProxyStore{
		RWMutex: sync.RWMutex{},
		Data:    make(map[string][]*ProxyRule),
		json:    newJSONFile("proxy", "proxystore"),
	}

	if err := proxyStore.json.Load(&proxyStore.Data); err != nil {
		aah.AppLog().Fatal(err)
	}
}

// PersistProxyStore persists the proxy configuration into filesystem.
func PersistProxyStore(_ *aah.Event) {
	if aah.AppProfile() == "prod" {
		proxyStore.persist()
	}
}

// AllProxies method returns all the proxy configuration from the store.
func AllProxies() map[string][]*ProxyRule {
	proxyStore.RLock()
	defer proxyStore.RUnlock()
	return proxyStore.Data
}

// ProxyByHost method returns proxy configured by hostname.
func ProxyByHost(hostName string) []*ProxyRule {
	return proxyStore.byHost(hostName)
}

// ProxyStats method returns stats about proxy configuration.
func ProxyStats() map[string]int {
	proxyStore.RLock()
	defer proxyStore.RUnlock()
	c := 0
	for _, v := range proxyStore.Data {
		c += len(v)
	}
	return map[string]int{
		"Host":       len(proxyStore.Data),
		"ProxyRules": c,
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Proxy Store types and its methods
//______________________________________________________________________________

// ProxyStore is data store thats holds proxy rules and definitions.
type ProxyStore struct {
	sync.RWMutex
	Modified bool
	Data     map[string][]*ProxyRule
	json     *jsonFile
}

func (ps *ProxyStore) byHost(hostName string) []*ProxyRule {
	if rules, found := ps.Data[strings.ToLower(hostName)]; found {
		return rules
	}
	return nil
}

func (ps *ProxyStore) persist() {
	ps.Lock()
	if err := ps.json.Persist(&ps.Data); err != nil {
		aah.AppLog().Error(err)
	}
	ps.Unlock()
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Proxy Rule, related types
//______________________________________________________________________________

// ProxyRule represents one proxy pass rule.
type ProxyRule struct {
	Last           bool              `json:"last,omitempty"`
	SkipTLSVerify  bool              `json:"skip_tls_verify,omitempty"`
	Path           string            `json:"path,omitempty"`
	TargetURL      string            `json:"target_url,omitempty"`
	QueryParams    map[string]string `json:"query_params,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	RequestHeader  *Header           `json:"request_header,omitempty"`
	ResponseHeader *Header           `json:"response_header,omitempty"`
}

// Header struct holds the headers request and response that needs to be added or removed.
type Header struct {
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}
