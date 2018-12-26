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
	"os"
	"strings"

	"thumbai/app/access"
	"thumbai/app/gomod"
	"thumbai/app/models"

	"aahframe.work"
	"aahframe.work/essentials"
)

// GoModController manages settings of go modules proxy server.
type GoModController struct {
	BaseController
}

// Index method display the Go modules settings page.
func (c *GoModController) Index() {
	data := aah.Data{
		"IsGoModules":   true,
		"Stats":         gomod.Settings.Stats,
		"Settings":      gomod.Settings,
		"GoModDisabled": access.GoModDisabled,
	}
	if adminEmail := aah.App().Config().StringDefault("thumbai.admin.contact_email", ""); len(adminEmail) > 0 {
		data["AdminContactEmail"] = adminEmail
	}
	c.Reply().HTML(data)
}

// SaveSettings method saves user settings into data store.
func (c *GoModController) SaveSettings(settings *models.ModuleSettings) {
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

	es := gomod.GetSettings()
	es.GoBinary = settings.GoBinary
	es.GoPath = settings.GoPath
	es.GoProxy = settings.GoProxy
	if err := gomod.SaveSettings(es); err != nil {
		c.Log().Error(err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "error occurred while saving settings",
		})
		return
	}

	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// Publish method go modules into repository.
//
// Supported formats:
//
// aahframe.work/aah@latest    # same (@latest is default for 'go get')
//
// aahframe.work/aah@v0.12.0   # records v0.12.0
//
// aahframe.work/aah@7e312af   # records v0.0.0-20180908054125-7e312af9202b
func (c *GoModController) Publish(pubReq *models.PublishRequest) {
	if !gomod.Settings.Enabled {
		c.Reply().ServiceUnavailable().JSON(aah.Data{
			"message": "Go Proxy Server unavailable due to prerequisites not met on server, please check thumbai logs",
		})
		return
	}

	if len(pubReq.Modules) == 0 {
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "module(s) path required",
		})
		return
	}

	go func() {
		for _, m := range pubReq.Modules {
			if ess.IsStrEmpty(m) {
				continue
			}
			if strings.Contains(m, gomod.FSPathDelimiter) {
				aah.App().Log().Errorf("Publish: invalid module path '%s'", m)
				continue
			}
			parts := strings.Split(m, "@")
			if len(parts) != 2 {
				aah.App().Log().Errorf("Publish: invalid module path '%s'", m)
				continue
			}
			if _, err := gomod.Download(&gomod.Module{Path: parts[0], Version: parts[1]}); err != nil {
				aah.App().Log().Error(err)
			}
		}
	}()

	c.Reply().Accepted().JSON(aah.Data{
		"message": "go module(s) publish request accepted",
	})
}
