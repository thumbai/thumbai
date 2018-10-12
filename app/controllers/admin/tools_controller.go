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
	"thumbai/app/proxy"
	"thumbai/app/vanity"

	"aahframe.work"
	"aahframe.work/ahttp"
)

// ToolsController defines admin tools actions.
type ToolsController struct {
	BaseController
}

// Index method serves the admin tools page.
func (c *ToolsController) Index() {
	c.Reply().HTML(aah.Data{
		"IsTools": true,
	})
}

// Export method implements THUMBAI configuration dats such as vanity, proxies, etc.
//
// NOTE: It does not export Go modules configuration, since inferred based on target
// enviroment.
func (c *ToolsController) Export() {
	c.Reply().
		Header(ahttp.HeaderContentDisposition, "attachment; filename=thumbai-configurations.json").
		JSON(models.Configuration{
			Vanities: vanity.All(),
			Proxies:  proxy.All(),
		})
}

// Import method implements THUMBAI import configuration.
//
// NOTE: Import overwrites the configuration if exists.
func (c *ToolsController) Import(config *models.Configuration) {
	if len(config.Vanities) > 0 {
		vanity.Import(config.Vanities)
	}
	if len(config.Proxies) > 0 {
		proxy.Import(config.Proxies)
	}
	c.Reply().NoContent()
}
