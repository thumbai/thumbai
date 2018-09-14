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
	"fmt"
	"os"

	"thumbai/app/gomod"
	"thumbai/app/models"

	"aahframe.work/aah"
	"aahframe.work/aah/essentials"
)

// GoModController manages settings of go modules proxy server.
type GoModController struct {
	BaseController
}

// Index method display the Go modules settings page.
func (c *GoModController) Index() {
	c.Reply().HTML(aah.Data{
		"IsGoModules": true,
		"Stats":       gomod.Stats,
		"Settings":    gomod.Settings,
	})
}

// SaveSettings method saves user settings into data store.
func (c *GoModController) SaveSettings(settings *models.ModuleSettings) {
	fmt.Printf("From Body: %#v\n", settings)
	var fieldErrors []*models.FieldError

	// Existence check
	if !ess.IsFileExists(settings.GoBinary) {
		fieldErrors = append(fieldErrors, &models.FieldError{
			Name:    "goBinary",
			Message: "Go binary does not exists on the server",
		})
	}
	fi, err := os.Lstat(settings.GoPath)
	if err != nil && os.IsNotExist(err) {
		if err = ess.MkDirAll(settings.GoPath, os.FileMode(0755)); err != nil {
			fieldErrors = append(fieldErrors, &models.FieldError{
				Name:    "goPath",
				Message: err.Error(),
			})
		}
	} else if err != nil {
		fieldErrors = append(fieldErrors, &models.FieldError{
			Name:    "goPath",
			Message: err.Error(),
		})
	}
	if len(fieldErrors) > 0 {
		c.Reply().BadRequest().JSON(aah.Data{"errors": fieldErrors})
		return
	}

	if !fi.IsDir() {
		fieldErrors = append(fieldErrors, &models.FieldError{
			Name:    "goPath",
			Message: "Given GOPATH is a file, it must be directory",
		})
	}
	// Validate Go version
	ver := gomod.GoVersion(settings.GoBinary)
	if ver == "0.0.0" || !gomod.InferGo111AndAbove(ver) {
		fieldErrors = append(fieldErrors, &models.FieldError{
			Name:    "goBinary",
			Message: "Requires go1.11 or above",
		})
	}
	if len(fieldErrors) > 0 {
		c.Reply().BadRequest().JSON(aah.Data{"errors": fieldErrors})
		return
	}

	es := models.GoModuleSettings()
	var changed bool
	if gomod.Settings.GoBinary != settings.GoBinary {
		changed = true
		es.GoBinary = settings.GoBinary
	}
	if gomod.Settings.GoPath != settings.GoPath {
		changed = true
		es.GoPath = settings.GoPath
	}
	if changed {
		if err := models.SaveModulesSettings(es); err != nil {
			c.Log().Error(err)
			c.Reply().InternalServerError().JSON(aah.Data{
				"message": "error occurred while saving settings",
			})
			return
		}
	}
	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}
