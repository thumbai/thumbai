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
	"thumbai/app/datastore"
	"thumbai/app/models"
	"thumbai/app/vanity"

	"aahframe.work/aah"
)

// VanityController controller manages the Vanity host and its packages.
type VanityController struct {
	BaseController
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// HTML page actions
//______________________________________________________________________________

// List method shows all vanity hosts configured in the vanities.
func (c *VanityController) List() {
	c.Reply().HTML(aah.Data{
		"IsVanity": true,
	})
}

// Show method shows all the vanity packages configured for the host.
func (c *VanityController) Show(hostName string) {
	c.Reply().HTML(aah.Data{
		"IsVanity":       true,
		"VanityHostName": hostName,
	})
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// API endpoint actions
//______________________________________________________________________________

// Hosts method returns all vanities configrations.
func (c *VanityController) Hosts() {
	c.Reply().JSON(aah.Data{
		"hosts": vanity.All(),
	})
}

// Host method returns the vanity packages by host.
func (c *VanityController) Host(hostName string) {
	pkgs := vanity.Get(hostName)
	c.Reply().JSON(aah.Data{
		"packages": pkgs,
	})
}

// AddHost method adds new host into vanity store.
func (c *VanityController) AddHost(hostName string) {
	var fieldErrors []*models.FieldError
	if err := vanity.AddHost(hostName); err != nil {
		switch {
		case err == datastore.ErrRecordAlreadyExists:
			fieldErrors = append(fieldErrors, &models.FieldError{
				Name:    "hostName",
				Message: "Vanity host already exists",
			})
			c.Reply().BadRequest().JSON(aah.Data{
				"message": "failed",
				"errors":  fieldErrors,
			})
			return
		}
		c.Log().Error(err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "failed",
		})
		return
	}
	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// DelHost method deletes the host and its vanity package configurations from vanity store.
func (c *VanityController) DelHost(hostName string) {
	if err := vanity.DelHost(hostName); err != nil {
		c.Log().Error(err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "failed",
		})
		return
	}
	c.Reply().NoContent()
}

// AddVanityPackage method adds the vanity package config into vanity store.
func (c *VanityController) AddVanityPackage(vp *models.VanityPackage) {
	var fieldErrors []*models.FieldError
	if err := vanity.Add(vp.Host, vp); err != nil {
		switch {
		case err == datastore.ErrRecordAlreadyExists:
			fieldErrors = append(fieldErrors, &models.FieldError{
				Name:    "vanityPkgPath",
				Message: "Vanity package already exists",
			})
			c.Reply().BadRequest().JSON(aah.Data{
				"message": "failed",
				"errors":  fieldErrors,
			})
			return
		}
		c.Log().Error(err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "failed",
		})
		return
	}
	if err := vanity.Add2Tree(vp); err != nil {
		c.Log().Error(err)
	}
	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// DelVanityPackage method deletes the vanity package config from vanity store.
func (c *VanityController) DelVanityPackage(hostName, pkg string) {
	if err := vanity.Del(hostName, pkg); err != nil {
		c.Log().Error(err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "failed",
		})
		return
	}
	c.Reply().NoContent()
}
