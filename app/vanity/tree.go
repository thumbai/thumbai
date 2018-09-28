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
	"strings"

	"thumbai/app/models"

	"aahframe.work/aah"
)

var errNodeExists = errors.New("tree: node exists")

var tree = &Tree{hosts: make(map[string]*node)}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Lookup method searches the vanity mapping defined in the store for given host
// and request path. If found returns the package info otherwise nil.
func Lookup(host, p string) *models.VanityPackage {
	return tree.lookup(host, p)
}

// Load method creates a vanity tree using data store.
func Load(_ *aah.Event) {
	vanities := All()
	if vanities == nil || len(vanities) == 0 {
		aah.AppLog().Info("Vanities are not yet configured on THUMBAI")
		return
	}

	for _, ps := range vanities {
		for _, p := range ps {
			if err := Add2Tree(p); err != nil {
				aah.AppLog().Error(err)
			}
		}
	}
	aah.AppLog().Info("Successfully created vanity route tree")
}

// Add2Tree method adds vanity package into vanity tree.
func Add2Tree(p *models.VanityPackage) error {
	if err := processVanityPackage(p); err != nil {
		return err
	}
	return tree.add(p.Host, p.Path, p)
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Tree struct and its methods
//______________________________________________________________________________

// Tree implements route implementation of vanity imports using Radix tree.
type Tree struct {
	hosts map[string]*node
}

func (t *Tree) lookupRoot(host string) *node {
	if root, found := t.hosts[strings.ToLower(host)]; found {
		return root
	}
	return nil
}

func (t *Tree) lookup(h, p string) *models.VanityPackage {
	root := t.lookupRoot(h)
	if root == nil {
		return nil
	}

	s, l, sn, pn := strings.ToLower(p), len(p), root, root
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

func (t *Tree) add(host, p string, v *models.VanityPackage) error {
	root := t.lookupRoot(host)
	if root == nil {
		root = &node{edges: make([]*node, 0)}
		t.hosts[host] = root
	}
	s, sn := strings.ToLower(p), root
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
