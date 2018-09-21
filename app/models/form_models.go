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

package models

// FormTargetURL represents fields of `formTargetURL` on page `/admin/proxy/edit.html`.
type FormTargetURL struct {
	Last          bool   `bind:"lastRule" json:"last,omitempty"`
	SkipTLSVerify bool   `bind:"skipTLSVerify" json:"skip_tls_verify,omitempty"`
	Host          string `bind:"hostName" json:"host,omitempty"`
	TargetURL     string `bind:"targetURL" json:"target_url,omitempty"`
	OldTargetURL  string `bind:"oldTargetURL" json:"old_target_url,omitempty"`
}

// FormConditions represents fields of `formConditions` on page `/admin/proxy/edit.html`.
type FormConditions struct {
	Host        string `bind:"hostName" json:"host,omitempty"`
	TargetURL   string `bind:"targetURL" json:"target_url,omitempty"`
	Path        string `bind:"path" json:"path,omitempty"`
	QueryParams string `bind:"queryParams" json:"query_params,omitempty"`
	Headers     string `bind:"headers" json:"headers,omitempty"`
}

// FormRedirects represents fields of `formRedirects` on page `/admin/proxy/edit.html`.
type FormRedirects struct {
	Host      string `bind:"hostName" json:"host,omitempty"`
	TargetURL string `bind:"targetURL" json:"target_url,omitempty"`
	Redirects string `bind:"redirects" json:"redirects,omitempty"`
}

// FormRestricts represents fields of `formRestricts` on page `/admin/proxy/edit.html`.
type FormRestricts struct {
	Host      string `bind:"hostName" json:"host,omitempty"`
	TargetURL string `bind:"targetURL" json:"target_url,omitempty"`
	ByExt     string `bind:"restrictsByExt" json:"by_ext,omitempty"`
	ByRegex   string `bind:"restrictsByRegex" json:"by_regex,omitempty"`
}

// FormStatics represents fields of `formStatics` on page `/admin/proxy/edit.html`.
type FormStatics struct {
	Host      string `bind:"hostName" json:"host,omitempty"`
	TargetURL string `bind:"targetURL" json:"target_url,omitempty"`
	Statics   string `bind:"staticDirs" json:"statics,omitempty"`
}

// FormRequestHeaders represents fields of `formRequestHeaders` on page `/admin/proxy/edit.html`.
type FormRequestHeaders struct {
	Host      string `bind:"hostName" json:"host,omitempty"`
	TargetURL string `bind:"targetURL" json:"target_url,omitempty"`
	Add       string `bind:"requestHeadersAdd" json:"add_headers,omitempty"`
	Remove    string `bind:"requestHeadersRemove" json:"remove_headers,omitempty"`
}

// FormResponseHeaders represents fields of `formResponseHeaders` on page `/admin/proxy/edit.html`.
type FormResponseHeaders struct {
	Host      string `bind:"hostName" json:"host,omitempty"`
	TargetURL string `bind:"targetURL" json:"target_url,omitempty"`
	Add       string `bind:"responseHeadersAdd" json:"add_headers,omitempty"`
	Remove    string `bind:"responseHeadersRemove" json:"remove_headers,omitempty"`
}
