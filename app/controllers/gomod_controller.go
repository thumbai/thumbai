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

package controllers

import (
	"net/http"
	"path/filepath"

	"thumbai/app/access"
	"thumbai/app/gomod"

	"aahframe.work"
	"aahframe.work/ahttp"
)

// GoModController handles `go mod` requests, this is gonna be future package management way.
type GoModController struct {
	*aah.Context
}

// Handle method handles the go mode requests {list, info, mod, zip}
func (c *GoModController) Handle(modPath string) {
	if !gomod.Settings.Enabled {
		c.Reply().ServiceUnavailable().Text("Go Proxy Server unavailable due to prerequisites not met on server, please check thumbai logs")
		return
	}
	if access.GoModDisabled {
		c.Reply().ServiceUnavailable().Text("Go Mod repository is disabled, please contact your administrator (%s).",
			aah.App().Config().StringDefault("thumbai.admin.contact_email", ""))
		return
	}

	c.Log().Debug("Requested Go Mod URI: ", modPath)
	mod, err := gomod.InferRequest(modPath)
	if err != nil && err != gomod.ErrGoModNotExist {
		c.Log().Warn(err)
		c.Reply().BadRequest().Text("%v", err)
		return
	}

	if err == gomod.ErrGoModNotExist {
		c.Log().Infof("Requested module or version [%s] does not exists in repository, "+
			"let's download it", modPath)
		result, err := gomod.Download(mod)
		if err != nil {
			c.Log().Error(err)
			c.Reply().InternalServerError().Text("%v %s",
				http.StatusInternalServerError,
				http.StatusText(http.StatusInternalServerError))
			return
		}
		mod = result
	}

	if mod.Action == "list" {
		c.Reply().
			ContentType(ahttp.ContentTypePlainText.String()).
			File(filepath.Join(gomod.Settings.ModCachePath, modPath))
		return
	}

	targetFile := filepath.Join(gomod.Settings.ModCachePath, mod.Path, "@v", mod.Version+"."+mod.Action)
	switch mod.Action {
	case "info":
		c.Reply().ContentType(ahttp.ContentTypeJSON.String()).File(targetFile)
	case "mod":
		c.Reply().ContentType(ahttp.ContentTypePlainText.String()).File(targetFile)
	case "zip":
		c.Reply().ContentType(ahttp.ContentTypeOctetStream.String()).File(targetFile)
	default:
		c.Reply().BadRequest().Text("invaild go mod request")
	}
}
