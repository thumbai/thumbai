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
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"aahframe.work/aah"
)

var vanityStore *VanityStore

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// LoadVanityStore loads the vanity data store from file system.
func LoadVanityStore(_ *aah.Event) {
	vanityStore = &VanityStore{
		RWMutex: sync.RWMutex{},
		Data:    make(map[string][]*VanityPackage),
		json:    newJSONFile("vanity", "vanitystore"),
	}

	if err := vanityStore.json.Load(&vanityStore.Data); err != nil {
		aah.AppLog().Fatal(err)
	}

	if err := vanityStore.processData(); err != nil {
		aah.AppLog().Fatal(err)
	}
}

// PersistVanityStore persists the vanity configuration into filesystem.
func PersistVanityStore(_ *aah.Event) {
	if aah.AppProfile() == "prod" {
		vanityStore.persist()
	}
}

// AllVanities returns all the vanity host configuration from store.
func AllVanities() map[string][]*VanityPackage {
	vanityStore.RLock()
	defer vanityStore.RUnlock()
	return vanityStore.Data
}

// VanityByHost method returns vanity configured packages by hostname.
func VanityByHost(hostName string) []*VanityPackage {
	return vanityStore.byHost(hostName)
}

// VanityStats method returns stats about vanities.
func VanityStats() map[string]int {
	vanityStore.RLock()
	defer vanityStore.RUnlock()
	stats := make(map[string]int)
	stats["Host"] = len(vanityStore.Data)
	c := 0
	for _, v := range vanityStore.Data {
		c += len(v)
	}
	stats["Packages"] = c
	return stats
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Vanity Store types and its methods
//______________________________________________________________________________

// VanityPackage holds the single vanity Go package for the domain.
type VanityPackage struct {
	Host string `json:"-"`
	Path string `json:"path,omitempty"`
	Repo string `json:"repo,omitempty"`
	VCS  string `json:"vcs,omitempty"`
	Src  string `json:"-"`
}

// VanityStore is data store thats holds vanity domain and its package configuration.
type VanityStore struct {
	sync.RWMutex
	Modified bool
	Data     map[string][]*VanityPackage
	json     *jsonFile
}

func (vs *VanityStore) byHost(hostName string) []*VanityPackage {
	if pkgs, found := vs.Data[strings.ToLower(hostName)]; found {
		return pkgs
	}
	return nil
}

func (vs *VanityStore) processData() error {
	vs.Lock()
	defer vs.Unlock()
	for host, d := range vs.Data {
		for _, p := range d {
			if p.Path == "/" {
				return fmt.Errorf("root path '/' is not valid Go package path for host:%s", host)
			}
			p.Host = host
			p.Path = strings.TrimSuffix(p.Path, "/")

			repo := strings.TrimSuffix(p.Repo, path.Ext(p.Repo))
			if strings.HasPrefix(p.Repo, "https://github.com/") {
				p.Src = fmt.Sprintf("%s %s/tree/master{/dir} %s/blob/master{/dir}/{file}#L{line}", repo, repo, repo)
			} else if strings.HasPrefix(p.Repo, "https://bitbucket.org/") {
				p.Src = fmt.Sprintf("%s %s/src/default{/dir} %s/src/default{/dir}/{file}#{file}-{line}", repo, repo, repo)
			}

			if len(p.VCS) == 0 {
				p.VCS = "git"
			}

			if p.VCS == "git" && filepath.Ext(p.Repo) != ".git" {
				return fmt.Errorf("invalid repo URL for path '%s', it doesn't end with .git", p.Path)
			}
		}
	}
	return nil
}

func (vs *VanityStore) persist() {
	vs.Lock()
	if err := vs.json.Persist(&vs.Data); err != nil {
		aah.AppLog().Error(err)
	}
	vs.Unlock()
}
