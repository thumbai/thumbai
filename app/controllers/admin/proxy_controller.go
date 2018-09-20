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
	"strings"
	"thumbai/app/models"
	"thumbai/app/proxy"
	"thumbai/app/store"
	"thumbai/app/util"

	"aahframe.work/aah"
	"aahframe.work/aah/essentials"
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

// AddRulePage method serves the add proxy rule page.
func (c *ProxyController) AddRulePage(hostName string) {
	c.Reply().HTMLf("edit.html", aah.Data{
		"IsProxy": true,
		"Rule":    &models.ProxyRule{Host: hostName},
	})
}

// EditRulePage method serves the edit proxy rule page.
func (c *ProxyController) EditRulePage(hostName, targetURL string) {
	rule := proxy.GetRule(hostName, targetURL)
	c.Reply().HTMLf("edit.html", aah.Data{
		"IsProxy": true,
		"Rule":    rule,
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
func (c *ProxyController) AddHost(proxyInfo *models.FormTargetURL) {
	var fieldErrors []*models.FieldError
	if err := proxy.AddHost(proxyInfo); err != nil {
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

// EditTargetURL method handles values of TargetURL, LastRule and SkipTLSVerify.
func (c *ProxyController) EditTargetURL(info *models.FormTargetURL) {
	if ess.IsStrEmpty(info.OldTargetURL) {
		if err := proxy.AddRule(&models.ProxyRule{
			Host:          info.Host,
			TargetURL:     info.TargetURL,
			Last:          info.Last,
			SkipTLSVerify: info.SkipTLSVerify,
		}); err != nil {
			c.Log().Errorf("Unable to added new proxy rule '%s' for %#v", err, info)
			c.Reply().InternalServerError().JSON(aah.Data{
				"message": "Unable to add new proxy rule!",
			})
			return
		}
		c.Reply().JSON(aah.Data{
			"message": "success",
		})
		return
	}

	rule := proxy.GetRule(info.Host, info.OldTargetURL)
	if rule == nil {
		c.Log().Errorf("Proxy rule not found for %#v", info)
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "Proxy rule not found",
		})
		return
	}
	rule.TargetURL = info.TargetURL
	rule.Last = info.Last
	rule.SkipTLSVerify = info.SkipTLSVerify
	if err := proxy.UpdateRule(info.OldTargetURL, rule); err != nil {
		c.Log().Errorf("EditTargetURL: Unable to update proxy rule %s", err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "Unable to update proxy rule",
		})
		return
	}
	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// EditConditions method handles Conditions values of proxy rule.
func (c *ProxyController) EditConditions(info *models.FormConditions) {
	rule := proxy.GetRule(info.Host, info.TargetURL)
	if rule == nil {
		c.Log().Errorf("Proxy rule not found for %#v", info)
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "Proxy rule not found",
		})
		return
	}

	path := strings.TrimSpace(info.Path)
	if len(path) > 0 {
		rule.Path = path
	}

	var fieldErrors []*models.FieldError
	queryParams, errs := util.Lines2MapString(info.QueryParams, "=")
	if len(errs) > 0 {
		c.Log().Errorf("Proxy conditions error on Query Param values %s", strings.Join(errs, ", "))
		fieldErrors = append(fieldErrors, &models.FieldError{
			Name:    "queryParams",
			Message: "Query params has invalid values: \n" + strings.Join(errs, "\n"),
		})
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "failed",
			"errors":  fieldErrors,
		})
		return
	}
	rule.QueryParams = queryParams

	headers, errs := util.Lines2MapString(info.Headers, "=")
	if len(errs) > 0 {
		c.Log().Errorf("Proxy conditions error on Header values %s", strings.Join(errs, ", "))
		fieldErrors = append(fieldErrors, &models.FieldError{
			Name:    "headers",
			Message: "Headers has invalid values: \n" + strings.Join(errs, "\n"),
		})
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "failed",
			"errors":  fieldErrors,
		})
		return
	}
	rule.Headers = headers

	if err := proxy.UpdateRule(info.TargetURL, rule); err != nil {
		c.Log().Errorf("EditConditions: Unable to update proxy rule %s", err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "Unable to update proxy rule",
		})
		return
	}
	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// EditRedirects method handles proxy redirects configurations.
func (c *ProxyController) EditRedirects(info *models.FormRedirects) {
	rule := proxy.GetRule(info.Host, info.TargetURL)
	if rule == nil {
		c.Log().Errorf("Proxy rule not found for %#v", info)
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "Proxy rule not found",
		})
		return
	}

	redirects, errs := util.Lines2Redirects(info.Redirects)
	if len(errs) > 0 {
		c.Log().Errorf("Proxy redirects have errors on values %s", strings.Join(errs, ", "))
		fieldErrors := append([]*models.FieldError{}, &models.FieldError{
			Name:    "redirects",
			Message: "Redirects has invalid values: \n" + strings.Join(errs, "\n"),
		})
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "failed",
			"errors":  fieldErrors,
		})
		return
	}
	rule.Redirects = redirects
	if err := proxy.UpdateRule(info.TargetURL, rule); err != nil {
		c.Log().Errorf("EditRedirects: Unable to update proxy rule %s", err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "Unable to update proxy rule",
		})
		return
	}

	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// EditRestricts method handles the file restricts by extension and regex.
func (c *ProxyController) EditRestricts(info *models.FormRestricts) {
	rule := proxy.GetRule(info.Host, info.TargetURL)
	if rule == nil {
		c.Log().Errorf("Proxy rule not found for %#v", info)
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "Proxy rule not found",
		})
		return
	}

	exts, errs := util.Lines2RestrictFiles(info.ByExt)
	if len(errs) > 0 {
		c.Log().Errorf("Proxy restrict by ext have errors on values %s", strings.Join(errs, ", "))
		fieldErrors := append([]*models.FieldError{}, &models.FieldError{
			Name:    "restrictsByExt",
			Message: "Restrict by extension has invalid values: \n" + strings.Join(errs, "\n"),
		})
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "failed",
			"errors":  fieldErrors,
		})
		return
	}
	if rule.RestrictFile == nil && len(exts) > 0 {
		rule.RestrictFile = &models.ProxyRestrictFile{Extension: exts}
	} else {
		rule.RestrictFile.Extension = exts
	}

	regexs, errs := util.Lines2RestrictFiles(info.ByRegex)
	if len(errs) > 0 {
		c.Log().Errorf("Proxy restrict by regex have errors on values %s", strings.Join(errs, ", "))
		fieldErrors := append([]*models.FieldError{}, &models.FieldError{
			Name:    "restrictsByRegex",
			Message: "Restrict by regex has invalid values: \n" + strings.Join(errs, "\n"),
		})
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "failed",
			"errors":  fieldErrors,
		})
		return
	}
	if rule.RestrictFile == nil && len(regexs) > 0 {
		rule.RestrictFile = &models.ProxyRestrictFile{Match: regexs}
	} else {
		rule.RestrictFile.Match = regexs
	}

	if err := proxy.UpdateRule(info.TargetURL, rule); err != nil {
		c.Log().Errorf("EditRedirects: Unable to update proxy rule %s", err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "Unable to update proxy rule",
		})
		return
	}

	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}

// EditStatics method handles static files directory configuration.
func (c *ProxyController) EditStatics(info *models.FormStatics) {
	fmt.Printf("%#v\n", info)
	rule := proxy.GetRule(info.Host, info.TargetURL)
	if rule == nil {
		c.Log().Errorf("Proxy rule not found for %#v", info)
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "Proxy rule not found",
		})
		return
	}

	statics, errs := util.Lines2Statics(info.Statics)
	if len(errs) > 0 {
		c.Log().Errorf("Proxy static directories mapping have errors on values %s", strings.Join(errs, ", "))
		fieldErrors := append([]*models.FieldError{}, &models.FieldError{
			Name:    "staticDirs",
			Message: "Static file directories has invalid values: \n" + strings.Join(errs, "\n"),
		})
		c.Reply().BadRequest().JSON(aah.Data{
			"message": "failed",
			"errors":  fieldErrors,
		})
		return
	}

	rule.Statics = statics
	if err := proxy.UpdateRule(info.TargetURL, rule); err != nil {
		c.Log().Errorf("EditStatics: Unable to update proxy rule %s", err)
		c.Reply().InternalServerError().JSON(aah.Data{
			"message": "Unable to update proxy rule",
		})
		return
	}

	c.Reply().JSON(aah.Data{
		"message": "success",
	})
}
