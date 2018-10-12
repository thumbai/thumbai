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
	"thumbai/app/access"

	"aahframe.work"
)

// BaseController for admin controllers.
// Created for any common abstraction of admin controllers.
type BaseController struct {
	*aah.Context
}

// Before method is an interceptor for admin path.
func (c *BaseController) Before() {
	if c.Req.Host != access.AdminHost || !access.IsAllowedFromIP(c.Req.ClientIP()) {
		c.Reply().Forbidden().Text("403 Forbidden")
		c.Abort()
		return
	}
}
