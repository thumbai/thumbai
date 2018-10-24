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

package access

import (
	"aahframe.work"
	"aahframe.work/essentials"
)

// Values
var (
	AdminHost     string
	AllowedIPs    []string
	UserStore     = map[string]User{}
	GoModDisabled bool
)

// Load method configures the thumbai access limits.
func Load(_ *aah.Event) {
	app := aah.App()
	cfg := app.Config()
	AdminHost = cfg.StringDefault("thumbai.admin.host", "")
	GoModDisabled = cfg.BoolDefault("thumbai.admin.disable.gomod_repo", false)

	var found bool
	AllowedIPs, found = cfg.StringList("thumbai.admin.allow_only")
	if found {
		AllowedIPs = append(AllowedIPs, "127.0.0.1", "::1", "[::1]")
	}

	if !cfg.IsExists("thumbai.user_datastore") {
		app.Log().Fatal("'thumbai.user_datastore' configuration is missing")
	}

	// processing user data
	userKeys := cfg.KeysByPath("thumbai.user_datastore")
	keyPrefix := "thumbai.user_datastore."
	for _, k := range userKeys {
		password := cfg.StringDefault(keyPrefix+k+".password", "")
		if ess.IsStrEmpty(password) {
			app.Log().Errorf("password value is missing for user [%s] on 'thumbai.user_datastore.%s.password'", k, k)
			continue
		}
		permissions, _ := cfg.StringList(keyPrefix + k + ".permissions")
		UserStore[k] = User{
			Username:    k,
			Password:    []byte(password),
			Permissions: permissions,
			Locked:      cfg.BoolDefault(keyPrefix+k+".locked", false),
			Expired:     cfg.BoolDefault(keyPrefix+k+".expired", false),
		}
	}
}

// IsAllowedFromIP method is used to check IP address allowed to admin interface.
func IsAllowedFromIP(ip string) bool {
	if len(AllowedIPs) == 0 {
		return true
	}
	for _, ap := range AllowedIPs {
		if ap == ip {
			return true
		}
	}
	return false
}

// User struct represents the THUMBAI application user.
type User struct {
	Username    string
	Password    []byte
	Permissions []string
	Locked      bool
	Expired     bool
}
