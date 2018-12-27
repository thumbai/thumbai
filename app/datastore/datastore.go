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

package datastore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"path/filepath"
	"time"

	"aahframe.work"
	"aahframe.work/essentials"
	bolt "go.etcd.io/bbolt"
)

var thumbaiDB *bolt.DB

// DB Errors
var (
	ErrRecordNotFound      = errors.New("db: record not found")
	ErrInvalidValue        = errors.New("db: invalid value")
	ErrRecordAlreadyExists = errors.New("db: record alredy exists")
)

// Bucket Names
var (
	BucketGoModules  = "gomodules"
	BucketGoVanities = "govanities"
	BucketProxies    = "proxies"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Connect method connects to DB on app start up.
func Connect(_ *aah.Event) {
	var err error
	app := aah.App()
	storeBasePath := app.Config().StringDefault("thumbai.admin.data_store.directory", "")
	if !ess.IsFileExists(storeBasePath) {
		if err = ess.MkDirAll(storeBasePath, 0755); err != nil {
			app.Log().Fatal(err)
		}
	}
	storePath := filepath.Join(storeBasePath, "thumbai.db")
	thumbaiDB, err = bolt.Open(storePath, 0644, &bolt.Options{Timeout: 100 * time.Millisecond})
	if err != nil {
		app.Log().Fatal(err)
	}
	if err = thumbaiDB.Update(func(tx *bolt.Tx) error {
		var err error
		if _, err = tx.CreateBucketIfNotExists([]byte(BucketGoModules)); err != nil {
			return err
		}
		if _, err = tx.CreateBucketIfNotExists([]byte(BucketGoVanities)); err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(BucketProxies))
		return err
	}); err != nil {
		app.Log().Fatal(err)
	}
	app.Log().Info("Connected to thumbai data store successfully at ", storePath)
}

// Disconnect method disconects from DB.
func Disconnect(_ *aah.Event) {
	if thumbaiDB != nil {
		if err := thumbaiDB.Close(); err != nil {
			aah.App().Log().Error(err)
		}
	}
}

// Get method gets the record from data store from the given bucket
// for the give key.
func Get(bucketName, key string, dst interface{}) error {
	return thumbaiDB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		k, v := c.Seek([]byte(key))
		if k == nil || string(k) != key {
			return ErrRecordNotFound
		}
		if dst == nil {
			return nil
		}
		d := gob.NewDecoder(bytes.NewReader(v))
		return d.Decode(dst)
	})
}

// Put method puts the value for the given key on the given bucket.
func Put(bucketName, key string, value interface{}) error {
	if value == nil {
		return ErrInvalidValue
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return err
	}
	return thumbaiDB.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Put([]byte(key), buf.Bytes())
	})
}

// Del method deletes the value from given bucket for the given key.
func Del(bucketName, key string) error {
	return thumbaiDB.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		k, _ := c.Seek([]byte(key))
		if k == nil || string(k) != key {
			return ErrRecordNotFound
		}
		return c.Delete()
	})
}

// BucketKeys method returns all the bucket keys for given name.
func BucketKeys(name string) []string {
	var keys []string
	_ = thumbaiDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if err := b.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		}); err != nil {
			aah.App().Log().Error("store.BucketKeys ", err)
		}
		return nil
	})
	return keys
}

// IsKeyExists method returns true if given key exists on given bucket.
func IsKeyExists(bucketName, key string) bool {
	return thumbaiDB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		k, _ := c.Seek([]byte(key))
		if k != nil && string(k) == key {
			return ErrRecordNotFound
		}
		return nil
	}) == ErrRecordNotFound
}

// Encode method encodes the Go object into bytes.
// func Encode(value interface{}) ([]byte, error) {
// 	if value == nil {
// 		return nil, nil
// 	}
// 	var buf bytes.Buffer
// 	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

// Decode method decodes the bytes into Go object.
// func Decode(dst interface{}, data []byte) error {
// 	if dst == nil || len(data) == 0 {
// 		return nil
// 	}
// 	d := gob.NewDecoder(bytes.NewReader(data))
// 	return d.Decode(dst)
// }
