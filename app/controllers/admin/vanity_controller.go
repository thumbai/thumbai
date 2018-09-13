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

package admin

import (
	"thumbai/app/models"

	"aahframe.work/aah"
)

// VanityController controller manages the Vanity host and its packages.
type VanityController struct {
	BaseController
}

// List method shows all vanity hosts configured in the vanities.
func (c *VanityController) List() {
	c.Reply().HTML(aah.Data{
		"IsVanity":    true,
		"AllVanities": models.AllVanities(),
	})
}

// Index method shows all the vanity packages configured for the host.
func (c *VanityController) Index(hostName string) {
	pkgs := models.VanityByHost(hostName)
	c.Reply().HTML(aah.Data{
		"IsVanity": true,
		"Packages": pkgs,
	})
}
