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
	"thumbai/app/models"

	"aahframe.work/aah"
)

// RedirectController controller manages the host and its redirects.
type RedirectController struct {
	BaseController
}

// List method shows all hosts and redirect counts.
func (c *RedirectController) List() {
	c.Reply().HTML(aah.Data{
		"IsRedirect":   true,
		"AllRedirects": models.AllRedirects(),
	})
}

// Index method shows all the redirect configuration for the host.
func (c *RedirectController) Index(hostName string) {
	redirects := models.RedirectByHost(hostName)
	c.Reply().HTML(aah.Data{
		"IsRedirect": true,
		"Redirects":  redirects,
	})
}
