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

package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"aahframe.work/aah"
)

func newJSONFile(store, filename string) *jsonFile {
	storePath := aah.AppConfig().StringDefault("thumbai.data_store.location",
		filepath.Join(aah.AppBaseDir(), "data"))
	storePath = filepath.Join(storePath, filename)

	return &jsonFile{
		Store:    store,
		Location: storePath,
	}
}

type jsonFile struct {
	Store    string
	Location string
}

func (j *jsonFile) Load(v interface{}) error {
	aah.AppLog().Infof("Processing %s store from %s", j.Store, j.Location)

	if err := os.MkdirAll(filepath.Dir(j.Location), os.FileMode(0755)); err != nil {
		return err
	}

	f, err := os.OpenFile(j.Location, os.O_RDWR|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("Unable to read %s store from %s", j.Store, j.Location)
	}

	fi, err := f.Stat()
	if err == nil && fi.Size() > 0 {
		if err = json.NewDecoder(f).Decode(v); err != nil {
			return fmt.Errorf("Unable to load %s store info: %v", j.Store, err)
		}
	}

	aah.AppLog().Infof("Successfully processed %s", j.Location)
	return nil
}

func (j *jsonFile) Persist(v interface{}) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("Unable parse %s store data: %v", j.Store, err)
	}

	if err := ioutil.WriteFile(j.Location, buf.Bytes(), os.FileMode(0644)); err != nil {
		return fmt.Errorf("Unable persist %s store data: %v", j.Store, err)
	}

	aah.AppLog().Infof("Successfully persisted %s configuration to %s", j.Store, j.Location)
	return nil
}
