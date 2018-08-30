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

// UserController handles subject related actions like login, logout, etc.
type UserController struct {
	BaseController
}

// BeforeLogin method is an interceptor for action Login.
// func (c *UserController) BeforeLogin() {
// 	if c.Subject().IsAuthenticated() {
// 		c.Reply().Redirect(c.RouteURL("dashboard"))
// 		c.Abort()
// 	}
// }

// Login method does the subject login.
func (c *UserController) Login() {
	if c.Subject().IsAuthenticated() {
		c.Reply().Redirect(c.RouteURL("dashboard"))
		return
	}
	c.Reply().HTMLl("basic.html", nil)
}

// Logout method does the subject logout.
func (c *UserController) Logout() {
	c.Subject().Logout()
	c.Reply().Redirect(c.RouteURL("login"))
}
