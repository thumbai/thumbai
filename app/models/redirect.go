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

var redirectStore *RedirectStore

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// LoadRedirectStore loads the redirect data store from file system.
func LoadRedirectStore(_ *aah.Event) {
	redirectStore = &RedirectStore{
		RWMutex: sync.RWMutex{},
		Data:    make(map[string][]*Redirect),
		json:    newJSONFile("redirect", "redirectstore"),
	}

	if err := redirectStore.json.Load(&redirectStore.Data); err != nil {
		aah.AppLog().Fatal(err)
	}
}

// PersistRedirectStore persists the redirect configurations into filesystem.
func PersistRedirectStore(_ *aah.Event) {
	if aah.AppProfile() == "prod" {
		redirectStore.persist()
	}
}

// AllRedirects methos redirects all the redirects configured in the system.
func AllRedirects() map[string][]*Redirect {
	redirectStore.RLock()
	defer redirectStore.RUnlock()
	return redirectStore.Data
}

// RedirectByHost method returns redirects configured by hostname.
func RedirectByHost(hostName string) []*Redirect {
	return redirectStore.byHost(hostName)
}

// RedirectStats method returns stats about redirects configuration.
func RedirectStats() map[string]int {
	redirectStore.RLock()
	defer redirectStore.RUnlock()
	stats := make(map[string]int)
	stats["Host"] = len(redirectStore.Data)
	c := 0
	for _, v := range redirectStore.Data {
		c += len(v)
	}
	stats["Redirects"] = c
	return stats
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Redirect Store types and its methods
//______________________________________________________________________________

// Redirect holds single redirect for proxy server.
type Redirect struct {
	Match  string
	Target string
	Code   int
}

// RedirectStore is data store that holds redirects configuration by host.
type RedirectStore struct {
	sync.RWMutex
	Modified bool
	Data     map[string][]*Redirect
	json     *jsonFile
}

func (rs *RedirectStore) byHost(hostName string) []*Redirect {
	if pkgs, found := rs.Data[strings.ToLower(hostName)]; found {
		return pkgs
	}
	return nil
}

func (rs *RedirectStore) persist() {
	rs.Lock()
	if err := rs.json.Persist(&rs.Data); err != nil {
		aah.AppLog().Error(err)
	}
	rs.Unlock()
}
