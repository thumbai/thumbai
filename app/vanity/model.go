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

package vanity

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"thumbai/app/datastore"
	"thumbai/app/models"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// All method returns all the vanity host configurations from store.
func All() map[string][]*models.VanityPackage {
	keys := datastore.BucketKeys(datastore.BucketGoVanities)
	allVanities := map[string][]*models.VanityPackage{}
	for _, k := range keys {
		vanities := make([]*models.VanityPackage, 0)
		_ = datastore.Get(datastore.BucketGoVanities, k, &vanities)
		allVanities[k] = vanities
	}
	return allVanities
}

// Stats method returns stats about vanities.
func Stats() map[string]int {
	stats := make(map[string]int)
	vanities := All()
	stats["Host"] = len(vanities)
	c := 0
	for _, v := range vanities {
		c += len(v)
	}
	stats["Packages"] = c
	return stats
}

// AddHost method adds the given host into vanities data store.
func AddHost(hostName string) error {
	hostName = strings.ToLower(hostName)
	if datastore.IsKeyExists(datastore.BucketGoVanities, hostName) {
		return datastore.ErrRecordAlreadyExists
	}
	return Add(hostName, nil)
}

// DelHost method deletes the given host from vanities store.
func DelHost(hostName string) error {
	return datastore.Del(datastore.BucketGoVanities, strings.ToLower(hostName))
}

// Get method returns the vanity package configurations for given host.
func Get(host string) []*models.VanityPackage {
	host = strings.ToLower(host)
	vanities := make([]*models.VanityPackage, 0)
	_ = datastore.Get(datastore.BucketGoVanities, host, &vanities)
	return vanities
}

// Add method adds the vanity package into vanities data store for given host.
func Add(host string, vp *models.VanityPackage) error {
	host = strings.ToLower(host)
	vanities := make([]*models.VanityPackage, 0)
	_ = datastore.Get(datastore.BucketGoVanities, host, &vanities)
	if vp == nil {
		return datastore.Put(datastore.BucketGoVanities, host, vanities)
	}
	for _, p := range vanities {
		if p.Path == vp.Path {
			return datastore.ErrRecordAlreadyExists
		}
	}
	return datastore.Put(datastore.BucketGoVanities, host, append(vanities, vp))
}

// Del method deletes vanity package from vanities data store for given host.
func Del(host, p string) error {
	vanities := make([]*models.VanityPackage, 0)
	_ = datastore.Get(datastore.BucketGoVanities, host, &vanities)
	f := -1
	for i, v := range vanities {
		if v.Path == p {
			f = i
			break
		}
	}
	if f > -1 {
		vanities = append(vanities[:f], vanities[f+1:]...)
		return datastore.Put(datastore.BucketGoVanities, host, vanities)
	}
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Unexporetd package methods
//______________________________________________________________________________

func processVanityPackage(p *models.VanityPackage) error {
	if p.Path == "/" {
		return fmt.Errorf("root path '/' is not valid Go package path for host:%s", p.Host)
	}
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
	return nil
}
