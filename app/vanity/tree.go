package vanity

import (
	"errors"
	"strings"

	"aahframe.work/aah"
	"gorepositree.com/app/data"
	"gorepositree.com/app/models"
)

var errNodeExists = errors.New("gorepositree/tree: node exists")

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Lookup method searches the vanity mapping defined in the store for given host
// and request path. If found returns the package info otherwise nil.
func Lookup(host, p string) *models.PackageInfo {
	host = "aahframe.work" // TODO remove
	return tree.lookup(host, p)
}

// Load method creates a vanity tree using data store.
func Load(_ *aah.Event) {
	s := data.Store()
	for h, ps := range s.Data.Vanites {
		for _, p := range ps {
			if err := tree.add(h, p.Path, p); err != nil {
				aah.AppLog().Error(err)
			}
		}
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Tree struct and its methods
//______________________________________________________________________________

// Tree implements route implementation of vanity imports using Radix tree.
type Tree struct {
	hosts map[string]*node
}

var tree = &Tree{hosts: make(map[string]*node)}

func (t *Tree) lookupRoot(host string) *node {
	if root, found := t.hosts[strings.ToLower(host)]; found {
		return root
	}
	return nil
}

func (t *Tree) lookup(h, p string) *models.PackageInfo {
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

func (t *Tree) add(host, p string, v *models.PackageInfo) error {
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
	value *models.PackageInfo
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

func newNode(label string, value *models.PackageInfo, edges []*node) *node {
	return &node{
		idx:   label[0],
		label: label,
		value: value,
		edges: edges,
	}
}
