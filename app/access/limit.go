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
	"aahframe.work/aah"
)

// Values
var (
	AdminHost  string
	AllowedIPs []string
)

// Load method configures the thumbai access limits.
func Load(_ *aah.Event) {
	cfg := aah.AppConfig()
	AdminHost = cfg.StringDefault("thumbai.admin.host", "")
	AllowedIPs, _ = cfg.StringList("thumbai.admin.allow_only")
	AllowedIPs = append(AllowedIPs, "127.0.0.1", "::1", "[::1]")
}

// IsAllowedFromIP method is used to check IP address allowed to admin interface.
func IsAllowedFromIP(ip string) bool {
	for _, ap := range AllowedIPs {
		if ap == ip {
			return true
		}
	}
	return false
}
