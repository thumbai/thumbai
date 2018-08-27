// Copyright 2018 Jeevanandam M. (https://github.com/jeevatkm, jeeva@myjeeva.com)
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
	"aahframe.work/aah"
	"gorepositree.com/app/models"
)

// DashboardController defines admin dashboard actions.
type DashboardController struct {
	BaseController
}

// Index method serves the admin dashboard.
func (c *DashboardController) Index() {
	c.Reply().Ok().HTML(aah.Data{
		"IsDashboard":   true,
		"VanityStats":   models.VanityStats(),
		"RedirectStats": models.RedirectStats(),
		"ProxyStats":    models.ProxyStats(),
	})
}

// ToAdminDashboard method redirects path '/@admin' to '/@admin/dashboard'.
func (c *DashboardController) ToAdminDashboard() {
	c.Reply().Redirect(c.RouteURL("dashboard"))
}
