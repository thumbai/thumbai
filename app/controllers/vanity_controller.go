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

package controllers

import (
	"aahframe.work/aah"
	"gorepositree.com/app/proxy"
	"gorepositree.com/app/redirect"
	"gorepositree.com/app/vanity"
)

// VanityController handles the classic `go get` handling, gonna become legacy.
type VanityController struct {
	*aah.Context
}

// Handle method handles Go vanity package request. If not found then it passes
// control over to proxy pass.
func (c *VanityController) Handle() {
	pkg := vanity.Lookup("aahframe.work", c.Req.Path) // TODO Remove
	if pkg == nil {
		if redirect.Do(c.Context) {
			return
		}

		c.Req.Host = "localhost:8080"          // TODO Remove
		c.Req.Unwrap().Host = "localhost:8080" // TODO Remove
		proxy.Do(c.Context)
		return
	}

	c.Reply().HTMLl("goget.html", aah.Data{"Vanity": pkg})
}
