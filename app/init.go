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

// aah application initialization - configuration, server extensions, middleware's, etc.
// Customize it per application use case.

package main

import (
	"html/template"
	"strings"

	"thumbai/app/access"
	"thumbai/app/commands"
	"thumbai/app/datastore"
	"thumbai/app/gomod"
	"thumbai/app/proxy"
	"thumbai/app/settings"
	"thumbai/app/util"
	"thumbai/app/vanity"

	"aahframe.work"
	_ "aahframe.work/minify/html"
)

func init() {
	app := aah.App()

	//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
	// Server Extensions
	// Doc: https://docs.aahframework.org/server-extension.html
	//__________________________________________________________________________
	app.OnInit(CheckConfig, 2)

	app.OnStart(datastore.Connect)
	app.OnStart(vanity.Load, 2)
	app.OnStart(proxy.Load, 2)
	app.OnStart(gomod.Infer)
	app.OnStart(access.Load)
	app.OnStart(settings.Load)

	app.OnPostShutdown(datastore.Disconnect)

	//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
	// Middleware's
	// Doc: https://docs.aahframework.org/middleware.html
	//
	// Executed in the order they are defined. It is recommended; NOT to change
	// the order of pre-defined aah framework middleware's.
	//__________________________________________________________________________
	app.HTTPEngine().Middlewares(
		aah.RouteMiddleware,
		// aah.CORSMiddleware,
		aah.BindMiddleware,
		aah.AntiCSRFMiddleware,
		aah.AuthcAuthzMiddleware,
		aah.ActionMiddleware,
	)

	//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
	// Add Custom Template Functions
	// Doc: https://docs.aahframework.org/template-funcs.html
	//__________________________________________________________________________
	app.AddTemplateFunc(template.FuncMap{
		"redirect2line":            util.ProxyRedirects2Lines,
		"mapstr2str":               util.MapString2String,
		"static2line":              util.ProxyStatics2Lines,
		"proxyconditionexists":     util.IsProxyConditionsExists,
		"proxyrestrictfilesexists": util.IsProxyRestrictFilesExists,
		"proxyrequesthdrexists":    util.IsProxyRequestHeadersExists,
		"proxyresponsehdrexists":   util.IsProxyResponseHeadersExists,
		"join":                     strings.Join,
	})

	if err := app.AddCommand(commands.Generate); err != nil {
		app.Log().Error(err)
	}
}
