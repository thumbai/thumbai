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

	"thumbai/app/gomod"

	"aahframe.work/aah"
	"aahframe.work/aah/ahttp"
	"aahframe.work/aah/essentials"
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

	modReq, err := gomod.InferRequest(modPath)
	if err != nil {
		c.Log().Warn(err)
		c.Reply().BadRequest().Text("%v", err)
		return
	}

	if !ess.IsFileExists(modReq.ModuleFilePath) || !ess.IsFileExists(modReq.FilePath) {
		c.Log().Infof("Requested module or version is not exists on server, let's download it '%s@%s'",
			modReq.Module, modReq.Version)
		if err := gomod.Download(modReq); err != nil {
			c.Log().Error(err)
			c.Reply().InternalServerError().Text("%v %s",
				http.StatusInternalServerError,
				http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	switch modReq.Action {
	case "list":
		c.Reply().ContentType(ahttp.ContentTypePlainText.String()).File(modReq.FilePath)
	case "info":
		c.Reply().ContentType(ahttp.ContentTypeJSON.String()).File(modReq.FilePath)
	case "mod":
		c.Reply().ContentType(ahttp.ContentTypePlainText.String()).File(modReq.FilePath)
	case "zip":
		c.Reply().ContentType(ahttp.ContentTypeOctetStream.String()).File(modReq.FilePath)
	default:
		c.Reply().BadRequest().Text("invaild go mod request")
	}
}
