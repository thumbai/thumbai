package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"aahframe.work/aah"
	"gorepositree.com/app/models"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Store method returns gorepositree data store instance.
func Store() *Gorepositree {
	return store
}

// Load method subscribes to application startup event to load vanity configurations.
func Load(e *aah.Event) {
	storePath := aah.AppConfig().StringDefault("gorepositree.store.path", "")
	if len(storePath) == 0 {
		aah.AppLog().Fatal("gorepositree: 'gorepositree.store.path' configuration is required")
	}

	var err error
	if store, err = loadStore(storePath); err != nil {
		aah.AppLog().Fatal(err)
	}

	intervalStr := aah.AppConfig().StringDefault("gorepositree.store.persist_interval", "1h")
	d, err := time.ParseDuration(intervalStr)
	if err != nil {
		aah.AppLog().Error("gorepositree: invalid value 'gorepositree.store.persist_interval', fallback to default 1 hour")
		d = 1 * time.Hour
	}
	store.persistTicker = time.NewTicker(d)
	go func(s *Gorepositree) {
		for _ = range s.persistTicker.C {
			if s.Modified {
				s.persistStore()
			}
		}
	}(store)
}

// Persist method saves vanity store back to file system.
func Persist(_ *aah.Event) {
	if store == nil {
		return
	}
	store.persistStore()
	store.persistTicker.Stop()
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Gorepositree type and its methods
//______________________________________________________________________________

// Gorepositree holds vanity store data
type Gorepositree struct {
	sync.RWMutex
	Modified      bool
	Location      string
	persistTicker *time.Ticker
	Data          *Data
}

// Data holds the gorepositree configuration data.
type Data struct {
	Vanites map[string][]*models.PackageInfo `json:"vanities"`
	Proxies map[string]*models.ProxyInfo     `json:"proxies"`
}

var store *Gorepositree

func (s *Gorepositree) persistStore() {
	s.Lock()
	defer s.Unlock()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(s.Data); err != nil {
		aah.AppLog().Errorf("gorepositree: unable parse store data: %v", err)
		return
	}

	if err := ioutil.WriteFile(s.Location, buf.Bytes(), 0644); err != nil {
		aah.AppLog().Errorf("gorepositree: unable save store data: %v", err)
	}
}

func (s *Gorepositree) processVanityData() error {
	s.Lock()
	defer s.Unlock()
	for _, d := range s.Data.Vanites {
		for _, p := range d {
			if p.Path == "/" {
				return errors.New("gorepositree: root path '/' is not valid Go package path")
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
				return fmt.Errorf("gorepositree: invalid repo URL for path '%s', it doesn't end with .git", p.Path)
			}
		}
	}
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Unexported methods
//______________________________________________________________________________

func loadStore(storePath string) (*Gorepositree, error) {
	fi, err := os.Lstat(storePath)
	if err != nil {
		return nil, fmt.Errorf("gorepositree: unable to stat data store: %v", err)
	}
	if fi.Mode().IsDir() {
		return nil, fmt.Errorf("gorepositree: invalid data store is directory")
	}

	f, err := os.Open(storePath)
	if err != nil {
		return nil, fmt.Errorf("gorepositree: unable to load data store from %s", storePath)
	}

	store := &Gorepositree{
		RWMutex:  sync.RWMutex{},
		Location: storePath,
		Data:     &Data{},
	}
	if err = json.NewDecoder(f).Decode(&store.Data); err != nil {
		return nil, fmt.Errorf("gorepositree: unable to load data store info: %v", err)
	}

	return store, store.processVanityData()
}
