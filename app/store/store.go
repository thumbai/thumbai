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

package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"path/filepath"
	"time"

	"aahframe.work/aah"
	bolt "go.etcd.io/bbolt"
)

var thumbaiDB *bolt.DB

// DB Errors
var (
	ErrRecordNotFound = errors.New("db: record not found")
	ErrInvalidValue   = errors.New("db: invalid value")
)

// Bucket Names
var (
	BucketGoModules = "gomodules"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Connect method connects to DB on app start up.
func Connect(_ *aah.Event) {
	opts := &bolt.Options{
		Timeout: 100 * time.Millisecond,
	}
	var err error
	storePath := filepath.Join(aah.AppBaseDir(), "data", "thumbai.db")
	thumbaiDB, err = bolt.Open(storePath, 0644, opts)
	if err != nil {
		aah.AppLog().Fatal(err)
	}
	if err = thumbaiDB.Update(func(tx *bolt.Tx) error {
		var err error
		if _, err = tx.CreateBucketIfNotExists([]byte(BucketGoModules)); err != nil {
			return err
		}
		// futher buckets
		return err
	}); err != nil {
		aah.AppLog().Fatal(err)
	}
	aah.AppLog().Info("Connected to data store successfully at ", storePath)
}

// Disconnect method disconects from DB.
func Disconnect(_ *aah.Event) {
	if thumbaiDB != nil {
		if err := thumbaiDB.Close(); err != nil {
			aah.AppLog().Error(err)
		}
	}
}

// Get method gets the record from data store from the given bucket
// for the give key.
func Get(bucketName, key string, value interface{}) error {
	return thumbaiDB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		k, v := c.Seek([]byte(key))
		if k == nil || string(k) != key {
			return ErrRecordNotFound
		}
		if value == nil {
			return nil
		}
		d := gob.NewDecoder(bytes.NewReader(v))
		return d.Decode(value)
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
