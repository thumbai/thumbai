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
	"thumbai/app/models"
	"thumbai/app/proxy"
	"thumbai/app/store"

	"aahframe.work/aah"
)

// ProxyController controller manages the host and its proxy rules.
type ProxyController struct {
	BaseController
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// HTML page actions
//______________________________________________________________________________

// List method shows all hosts and proxy rules count.
func (c *ProxyController) List() {
	c.Reply().HTML(aah.Data{
		"IsProxy":    true,
		"AllProxies": proxy.All(),
	})
}

// Show method shows all the proxy rules configuration for the host.
func (c *ProxyController) Show(hostName string) {
	proxyRules := proxy.Get(hostName)
	c.Reply().HTML(aah.Data{
		"IsProxy":       true,
		"ProxyHostName": hostName,
		"ProxyRules":    proxyRules,
	})
}

// AddRule method serves the add proxy rule page.
func (c *ProxyController) AddRule(hostName string) {
	c.Reply().HTMLf("edit.html", aah.Data{
		"IsProxy":       true,
		"ProxyHostName": hostName,
	})
}

// EditRule method serves the edit proxy rule page.
func (c *ProxyController) EditRule(hostName, targetURL string) {
	proxyRules := proxy.Get(hostName)
	var rule *models.ProxyRule
	for _, e := range proxyRules {
		if e.Host == hostName && e.TargetURL == targetURL {
			rule = e
			break
		}
	}
	rule.Redirects = make([]*models.ProxyRedirect, 0)
	rule.Redirects = append(rule.Redirects,
		&models.ProxyRedirect{
			Match:  "{^/v0.5/(.*)$}",
			Target: "https://docs.aahframework.org/{1}",
		},
		&models.ProxyRedirect{
			Match:  "{^/v0.6/(.*)$}",
			Target: "https://docs.aahframework.org/{1}",
		},
		&models.ProxyRedirect{
			Match:  "/v0.7/centralized-error-handler.html",
			Target: "https://docs.aahframework.org/error-handling.html",
		},
		&models.ProxyRedirect{
			Match:  "/v0.7/external-json-library.html",
			Target: "https://docs.aahframework.org",
			Code:   307,
		})
	c.Reply().HTMLf("edit.html", aah.Data{
		"IsProxy":        true,
		"ProxyHostName":  hostName,
		"ProxyTargetURL": targetURL,
		"ProxyRule":      rule,
	})
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// API endpoint actions
//______________________________________________________________________________

// Hosts method returns all proxies configrations.
func (c *ProxyController) Hosts() {
	c.Reply().JSON(aah.Data{
		"hosts": proxy.All(),
	})
}

// Host method returns the proxy rules by host.
func (c *ProxyController) Host(hostName string) {
	rules := proxy.Get(hostName)
	c.Reply().JSON(aah.Data{
		"proxy_rules": rules,
	})
}

// AddHost method adds the new proxy host into proxy store.
func (c *ProxyController) AddHost(proxyRule *models.ProxyRule) {
	fmt.Println("AddHost proxyRule", proxyRule)
	var fieldErrors []*models.FieldError
	if err := proxy.AddHost(proxyRule); err != nil {
		switch {
		case err == store.ErrRecordAlreadyExists:
			fieldErrors = append(fieldErrors, &models.FieldError{
				Name:    "hostName",
				Message: "Proxy host already exists",
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

// DelHost method deletes the proxy host and its configurations from proxy store.
func (c *ProxyController) DelHost(hostName string) {
	if err := proxy.DelHost(hostName); err != nil {
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
