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

package gomod

import (
	"thumbai/app/datastore"
	"thumbai/app/models"

	"aahframe.work/aah"
)

// Stats returns go modules statistics.
func Stats() *models.ModuleStats {
	stats := &models.ModuleStats{}
	if err := datastore.Get(datastore.BucketGoModules, "stats", stats); err != nil {
		if err == datastore.ErrRecordNotFound {
			aah.AppLog().Info("Go Modules stats data currently unavailable")
		} else {
			aah.AppLog().Error(err)
		}
	}
	return stats
}

// SaveStats method saves the given stats into data store.
func SaveStats(stats *models.ModuleStats) error {
	return datastore.Put(datastore.BucketGoModules, "stats", stats)
}

// GetSettings method gets the modules settings from data store.
func GetSettings() *models.ModuleSettings {
	settings := &models.ModuleSettings{}
	if err := datastore.Get(datastore.BucketGoModules, "settings", settings); err != nil {
		if err != datastore.ErrRecordNotFound {
			aah.AppLog().Error(err)
		}
	}
	return settings
}

// SaveSettings method saves the given modules into data store.
func SaveSettings(settings *models.ModuleSettings) error {
	return datastore.Put(datastore.BucketGoModules, "settings", settings)
}
