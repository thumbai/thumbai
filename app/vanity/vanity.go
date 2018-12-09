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
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"thumbai/app/models"

	"aahframe.work"
	"aahframe.work/essentials"
)

var errNodeExists = errors.New("tree: node exists")

// Thumbai vanities instance.
var Thumbai *vanities

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Lookup method searches the vanity mapping defined in the store for given host
// and request path. If found returns the package info otherwise nil.
func Lookup(host, p string) *models.VanityPackage {
	vh := Thumbai.Lookup(host)
	if vh == nil {
		return nil
	}
	if p == "/" || p == "" {
		return vh.Root
	}
	vp := vh.Lookup(p)
	if vp == nil {
		if vh.IsRootVanity(p) { // check root vanity
			return vh.Root
		}
	}
	return vp
}

// Load method creates a vanity tree using data store.
func Load(_ *aah.Event) {
	log := aah.App().Log()
	Thumbai = &vanities{RWMutex: sync.RWMutex{}, Hosts: make(map[string]*vanityHost)}
	allVanities := All()
	if len(allVanities) == 0 {
		log.Info("Vanities are not yet configured on THUMBAI")
		return
	}

	for _, ps := range allVanities {
		for _, p := range ps {
			if err := Add2Tree(p); err != nil {
				log.Error(err)
			}
		}
	}
	log.Info("Successfully created vanity route tree")
}

// Add2Tree method adds vanity package into vanity tree.
func Add2Tree(p *models.VanityPackage) error {
	if err := processVanityPackage(p); err != nil {
		return err
	}
	host := Thumbai.AddHost(p.Host)
	if p.Path == "@" {
		host.AddRootVanity(p)
	} else if err := host.AddVanity2Tree(p.Path, p); err != nil {
		return err
	}
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Vanity struct and its methods
//______________________________________________________________________________

type vanities struct {
	sync.RWMutex
	Hosts map[string]*vanityHost
}

func (v *vanities) Lookup(hostname string) *vanityHost {
	v.RLock()
	defer v.RUnlock()
	if h, f := v.Hosts[strings.ToLower(hostname)]; f {
		return h
	}
	return nil
}

func (v *vanities) AddHost(hostname string) *vanityHost {
	h := v.Lookup(hostname)
	if h == nil {
		h = &vanityHost{
			RWMutex: sync.RWMutex{},
			Name:    hostname,
			Tree:    &node{edges: make([]*node, 0)},
		}
		v.Lock()
		v.Hosts[strings.ToLower(hostname)] = h
		v.Unlock()
	}
	return h
}

func (v *vanities) DelHost(hostname string) {
	v.Lock()
	delete(v.Hosts, hostname)
	v.Unlock()
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// VanityHost struct and its methods
//______________________________________________________________________________

type vanityHost struct {
	sync.RWMutex
	Name        string
	Root        *models.VanityPackage
	RootSubPkgs map[string]bool
	Tree        *node
}

func (vh *vanityHost) AddRootVanity(vp *models.VanityPackage) {
	vh.Lock()
	defer vh.Unlock()
	p := *vp
	p.Path = ""
	vh.Root = &p
	if !ess.IsStrEmpty(vh.Root.RootSubPkgs) {
		pkgs := strings.Split(vh.Root.RootSubPkgs, ",")
		vh.RootSubPkgs = map[string]bool{}
		for _, v := range pkgs {
			vh.RootSubPkgs[strings.TrimSpace(v)] = true
		}
	}
}

func (vh *vanityHost) IsRootVanity(p string) bool {
	p = strings.TrimLeft(p, "/")
	if i := strings.IndexByte(p, '/'); i > 0 {
		p = p[:i]
	}
	_, found := vh.RootSubPkgs[p]
	return found
}

func (vh *vanityHost) Lookup(p string) *models.VanityPackage {
	vh.RLock()
	defer vh.RUnlock()
	if vh.Tree == nil {
		return nil
	}

	s, l, sn, pn := strings.ToLower(p), len(p), vh.Tree, vh.Tree
	ll := l
	for {
		i, max := 0, len(sn.label)
		if ll <= max {
			max = ll
		}
		for i < max && s[i] == sn.label[i] {
			i++
		}
		if i != max {
			return nil
		}

		s = s[i:]
		ll = len(s)
		if ll == 0 {
			goto nomore
		}
		n := sn.findByIdx(s[0])
		if n == nil {
			goto nomore
		}
		sn = n
		if sn.value != nil { // track last non-nil node
			pn = sn
		}
	}

nomore:
	if sn.value == nil {
		return pn.value
	}
	return sn.value
}

func (vh *vanityHost) AddVanity2Tree(p string, v *models.VanityPackage) error {
	vh.Lock()
	defer vh.Unlock()
	s, sn := strings.ToLower(p), vh.Tree
	for {
		i, max := 0, min(len(s), len(sn.label))
		for i < max && s[i] == sn.label[i] {
			i++
		}
		switch {
		case i == 0: // assign to current/root node
			sn.idx = s[0]
			sn.label = s
			if v != nil {
				sn.value = v
			}
		case i < len(sn.label): // split the node
			edge := newNode(sn.label[i:], sn.value, sn.edges)
			sn.idx = sn.label[0]
			sn.label = sn.label[:i]
			sn.value = nil
			sn.edges = []*node{edge}
			if i == len(s) {
				if sn.value != nil {
					return errNodeExists
				}
				sn.value = v
			} else {
				sn.edges = append(sn.edges, newNode(s[i:], v, []*node{}))
			}
		case i < len(s): // navigate, check and add new edge
			s = s[i:]
			if n := sn.findByIdx(s[0]); n != nil {
				sn = n
				continue
			}
			sn.edges = append(sn.edges, newNode(s, v, []*node{}))
		default:
			if v != nil {
				if sn.value != nil {
					return errNodeExists
				}
				sn.value = v
			}
		}
		return nil
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// node struct and its methods
//______________________________________________________________________________

type node struct {
	idx   byte
	label string
	value *models.VanityPackage
	edges []*node
}

func (n *node) findByIdx(i byte) *node {
	for _, e := range n.edges {
		if e.idx == i {
			return e
		}
	}
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Unexported methods
//______________________________________________________________________________

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func newNode(label string, value *models.VanityPackage, edges []*node) *node {
	return &node{
		idx:   label[0],
		label: label,
		value: value,
		edges: edges,
	}
}

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
