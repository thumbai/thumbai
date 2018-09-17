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

// PublishRequest struct used to accept the module publish request.
type PublishRequest struct {
	Modules []string `json:"modules"`
}

// FieldError to represent HTML field error info on JSON response.
type FieldError struct {
	Name    string
	Message string
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Module types
//______________________________________________________________________________

// ModuleSettings represents the go modules settings.
type ModuleSettings struct {
	GoPath   string `bind:"goPath"`
	GoBinary string `bind:"goBinary"`
	GoProxy  string `bind:"goProxy"`
}

// ModuleStats represents the go modules statics on the server.
type ModuleStats struct {
	TotalCount int64
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Vanity package type
//______________________________________________________________________________

// VanityPackage holds the single vanity Go package for the domain.
type VanityPackage struct {
	Host string `bind:"hostName" json:"host"`
	Path string `bind:"vanityPkgPath" json:"path,omitempty"`
	Repo string `bind:"vanityPkgRepo" json:"repo,omitempty"`
	VCS  string `bind:"vanityPkgVcs" json:"vcs,omitempty"`
	Src  string `json:"-"`
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Proxy Rule, related types
//______________________________________________________________________________

// ProxyRule represents one proxy pass rule.
type ProxyRule struct {
	Last           bool               `json:"last,omitempty"`
	SkipTLSVerify  bool               `json:"skip_tls_verify,omitempty"`
	Host           string             `bind:"hostName" json:"host,omitempty"`
	Path           string             `json:"path,omitempty"`
	TargetURL      string             `json:"target_url,omitempty"`
	QueryParams    map[string]string  `json:"query_params,omitempty"`
	Headers        map[string]string  `json:"headers,omitempty"`
	RequestHeader  *ProxyHeader       `json:"request_header,omitempty"`
	ResponseHeader *ProxyHeader       `json:"response_header,omitempty"`
	RestrictFile   *ProxyRestrictFile `json:"restrict_file,omitempty"`
	Statics        []*ProxyStatic     `json:"statics,omitempty"`
}

// ProxyHeader struct holds the headers request and
// response that needs to be added or removed.
type ProxyHeader struct {
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}

// ProxyStatic struct holds the static files directory mappings.
// It used to check before proxying request to upstream targets.
type ProxyStatic struct {
	StripPrefix string `json:"strip_prefix,omitempty"`
	TargetPath  string `json:"target_path,omitempty"`
}

// ProxyRestrictFile structs holds the restricts configurations of by file
// extension and regex match.
type ProxyRestrictFile struct {
	Extension []string `json:"extension,omitempty"`
	Match     []string `json:"match,omitempty"`
}
