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

// ProxyController controller manages the host and its proxy rules.
type ProxyController struct {
	BaseController
}

// List method shows all hosts and proxy rules count.
func (c *ProxyController) List() {
	c.Reply().HTML(aah.Data{
		"IsProxy":    true,
		"AllProxies": models.AllProxies(),
	})
}

// Index method shows all the proxy rules configuration for the host.
func (c *ProxyController) Index(hostName string) {
	proxyRules := models.ProxyByHost(hostName)
	c.Reply().HTML(aah.Data{
		"IsProxy":    true,
		"ProxyRules": proxyRules,
	})
}
