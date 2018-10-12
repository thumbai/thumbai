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
	"thumbai/app/gomod"
	"thumbai/app/proxy"
	"thumbai/app/vanity"

	"aahframe.work"
)

// DashboardController defines admin dashboard actions.
type DashboardController struct {
	BaseController
}

// Index method serves the admin dashboard.
func (c *DashboardController) Index() {
	c.Reply().Ok().HTML(aah.Data{
		"IsDashboard":      true,
		"GoModulesEnabled": gomod.Settings.Enabled,
		"GoModulesStats":   gomod.Settings.Stats,
		"VanityStats":      vanity.Stats(),
		"ProxyStats":       proxy.Stats(),
	})
}

// Credits method serves the credits page.
func (c *DashboardController) Credits() {
	c.Reply().HTML(aah.Data{
		"IsCredits": true,
	})
}

// ToAdminDashboard method redirects path '/thumbai' to '/thumbai/dashboard.html'.
func (c *DashboardController) ToAdminDashboard() {
	c.Reply().Redirect(c.RouteURL("dashboard"))
}
