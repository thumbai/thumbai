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
	"bufio"
	"errors"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"thumbai/app/models"

	"aahframe.work/essentials"
)

// Lines2MapString method transforms multi-line input to map[string]string by given
// delimiter.
func Lines2MapString(input, delimiter string, hdr bool) (map[string]string, []string) {
	if ess.IsStrEmpty(input) || ess.IsStrEmpty(delimiter) {
		return nil, nil
	}
	result := make(map[string]string)
	errResult := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, delimiter)
		if len(parts) != 2 {
			errResult = append(errResult, line)
			continue
		}
		result[mapKey(parts[0], hdr)] = strings.TrimSpace(parts[1])
	}
	return result, errResult
}

// Lines2Redirects method transforms the multi-lines to slice proxy redirects.
func Lines2Redirects(input string) ([]*models.ProxyRedirect, []string) {
	if ess.IsStrEmpty(input) {
		return nil, nil
	}

	result := make([]*models.ProxyRedirect, 0)
	errResult := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, ",")
		partsLen := len(parts)
		if partsLen == 2 {
			parts[1] = strings.TrimSpace(parts[1])
			isAbs := IsAbsURL(parts[1])
			if !isAbs && parts[1][0] != '/' {
				errResult = append(errResult, parts[1]+" - relative path must start with '/'")
				continue
			}
			result = append(result, &models.ProxyRedirect{
				Match:  strings.TrimSpace(parts[0]),
				Target: parts[1],
				IsAbs:  IsAbsURL(parts[1]),
			})
			continue
		} else if partsLen == 3 {
			parts[1] = strings.TrimSpace(parts[1])
			isAbs := IsAbsURL(parts[1])
			if !isAbs && parts[1][0] != '/' {
				errResult = append(errResult, parts[1]+" - relative path must start with '/'")
				continue
			}
			code, err := strconv.Atoi(strings.TrimSpace(parts[2]))
			if err == nil && IsSupportedRedirectCode(code) {
				result = append(result, &models.ProxyRedirect{
					Match:  strings.TrimSpace(parts[0]),
					Target: parts[1],
					Code:   code,
					IsAbs:  IsAbsURL(parts[1]),
				})
				continue
			}
		}
		errResult = append(errResult, line)
	}
	return result, errResult
}

// Lines2RestrictFiles method transforms the lines into string slice.
func Lines2RestrictFiles(input string) ([]string, []string) {
	if ess.IsStrEmpty(input) {
		return nil, nil
	}
	result := make([]string, 0)
	errResult := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		if line[0] == '{' || line[len(line)-1] == '}' {
			if _, err := regexp.Compile(line[1 : len(line)-1]); err != nil {
				errResult = append(errResult, line)
				continue
			}
		} else if line[0] != '.' {
			errResult = append(errResult, line+" - missing dot prefix in the extension")
			continue
		}
		result = append(result, line)
	}
	return result, errResult
}

// Lines2Statics method transforms the lines into proxy static slice.
func Lines2Statics(input string) ([]*models.ProxyStatic, []string) {
	if ess.IsStrEmpty(input) {
		return nil, nil
	}
	result := make([]*models.ProxyStatic, 0)
	errResult := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			parts[0] = strings.TrimSpace(parts[0])
			if err := IsVaildPath(parts[0]); err != nil {
				errResult = append(errResult, err.Error())
				continue
			}
			parts[1] = strings.TrimSpace(parts[1])
			result = append(result, &models.ProxyStatic{TargetPath: parts[0], StripPrefix: parts[1]})
		} else if len(parts) == 1 {
			parts[0] = strings.TrimSpace(parts[0])
			if err := IsVaildPath(parts[0]); err != nil {
				errResult = append(errResult, err.Error())
				continue
			}
			result = append(result, &models.ProxyStatic{TargetPath: parts[0]})
		} else {
			errResult = append(errResult, line)
		}
	}
	return result, errResult
}

// Lines2HeaderSlice method transforms the lines into header string slice.
func Lines2HeaderSlice(input, shouldNotExists, msg string) ([]string, []string) {
	if ess.IsStrEmpty(input) {
		return nil, nil
	}
	result := make([]string, 0)
	errResult := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		if strings.Index(line, shouldNotExists) > 0 {
			errResult = append(errResult, line+msg)
			continue
		}
		result = append(result, http.CanonicalHeaderKey(line))
	}
	sort.Strings(result)
	return result, errResult
}

// IsSupportedRedirectCode method returns if given code is supported by proxy.
func IsSupportedRedirectCode(code int) bool {
	switch code {
	case http.StatusMovedPermanently, http.StatusFound,
		http.StatusUseProxy, http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	}
	return false
}

// IsAbsURL returns true if its absolute URL otherwise false.
//
// Note: here `url.Parse` is not suitable. since target URL
// might be invaild when regex composition is used.
func IsAbsURL(u string) bool {
	u = strings.TrimSpace(u)
	return strings.HasPrefix(u, "http") || strings.HasPrefix(u, "https")
}

// IsVaildPath method checks the absolute path and existence the returns error
// if not valid otherwise nil.
func IsVaildPath(p string) error {
	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		return errors.New(p + " - absolute path required")
	}
	if !ess.IsFileExists(p) {
		return errors.New(p + " - path not exists on server")
	}
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Unexported methods
//______________________________________________________________________________

func sortHeaderKeys(hdrs map[string]string) []string {
	var keys []string
	for key := range hdrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mapKey(v string, hdr bool) string {
	v = strings.TrimSpace(v)
	if hdr {
		return http.CanonicalHeaderKey(v)
	}
	return v
}
