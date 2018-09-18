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
	"fmt"
	"strconv"
	"strings"

	"thumbai/app/models"
)

// MapStringToString method converts map type to comma separated string.
func MapStringToString(input map[string]string) string {
	var str string
	for k, v := range input {
		str += ", " + fmt.Sprintf("%v: %v", k, v)
	}
	return strings.TrimPrefix(str, ", ")
}

// ProxyRedirects2Lines method transforms the proxy redirects into display line text.
func ProxyRedirects2Lines(redirectRules []*models.ProxyRedirect) string {
	if len(redirectRules) == 0 {
		return ""
	}
	var redirects []string
	for _, r := range redirectRules {
		str := r.Match + ", " + r.Target + ", "
		if r.Code == 0 {
			str += "301"
		} else {
			str += strconv.Itoa(r.Code)
		}
		redirects = append(redirects, str)
	}
	return strings.Join(redirects, "\n")
}
