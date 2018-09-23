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

package util

import (
	"strconv"
	"strings"

	"thumbai/app/models"

	"aahframe.work/aah/essentials"
)

// ProxyRedirects2Lines method transforms the proxy redirects into display line text.
func ProxyRedirects2Lines(redirectRules []*models.ProxyRedirect) string {
	if len(redirectRules) == 0 {
		return ""
	}
	var redirects []string
	for _, r := range redirectRules {
		str := r.Match + ", " + r.Target
		if r.Code > 0 {
			str += ", " + strconv.Itoa(r.Code)
		}
		redirects = append(redirects, str)
	}
	return strings.Join(redirects, "\n")
}

// MapString2String method transforms the map into multi-line wiyth given delimiter.
func MapString2String(values map[string]string, delimiter, joinstr string) string {
	if len(values) == 0 {
		return ""
	}
	var result []string
	for _, k := range sortHeaderKeys(values) {
		result = append(result, k+delimiter+values[k])
	}
	return strings.Join(result, joinstr)
}

// ProxyStatics2Lines method transforms the static config into mutli-line string.
func ProxyStatics2Lines(statics []*models.ProxyStatic) string {
	if len(statics) == 0 {
		return ""
	}
	var lines []string
	for _, s := range statics {
		str := s.TargetPath
		if !ess.IsStrEmpty(s.StripPrefix) {
			str += ", " + s.StripPrefix
		}
		lines = append(lines, str)
	}
	return strings.Join(lines, "\n")
}

// IsProxyConditionsExists method returns true if path, query params or headers exists
// otherwise false.
func IsProxyConditionsExists(rule *models.ProxyRule) bool {
	if rule == nil {
		return false
	}
	return len(rule.Path) > 0 || len(rule.QueryParams) > 0 || len(rule.Headers) > 0
}

// IsProxyRestrictFilesExists method returns true if rrestrict files config exists
// otherwise false.
func IsProxyRestrictFilesExists(rule *models.ProxyRule) bool {
	if rule == nil || rule.RestrictFiles == nil {
		return false
	}
	return len(rule.RestrictFiles.Extensions) > 0 || len(rule.RestrictFiles.Regexs) > 0
}

// IsProxyRequestHeadersExists method returns true if request headers present
// otherwise false.
func IsProxyRequestHeadersExists(rule *models.ProxyRule) bool {
	if rule == nil || rule.RequestHeaders == nil {
		return false
	}
	return len(rule.RequestHeaders.Add) > 0 || len(rule.RequestHeaders.Remove) > 0
}

// IsProxyResponseHeadersExists method returns true if request headers present
// otherwise false.
func IsProxyResponseHeadersExists(rule *models.ProxyRule) bool {
	if rule == nil || rule.ResponseHeaders == nil {
		return false
	}
	return len(rule.ResponseHeaders.Add) > 0 || len(rule.ResponseHeaders.Remove) > 0
}
