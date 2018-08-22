// Copyright 2018 Jeevanandam M. (https://github.com/jeevatkm, jeeva@myjeeva.com)
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

// PackageInfo holds single Go package path belongs to a host.
type PackageInfo struct {
	Path string `json:"path,omitempty"`
	Repo string `json:"repo,omitempty"`
	VCS  string `json:"vcs,omitempty"`
	Src  string `json:"-"`
}

// ProxyInfo holds single Reverse Proxy server info.
type ProxyInfo struct {
	URL       string      `json:"url,omitempty"`
	Redirects []*Redirect `json:"redirects,omitempty"`
	ReqHdr    *Hdr        `json:"request_header,omitempty"`
	ResHdr    *Hdr        `json:"response_header,omitempty"`
}

// Redirect holds single redirect for proxy server.
type Redirect struct {
	Match  string
	Target string
	Code   int
}

// Hdr struct holds the request needs to be added or removed.
type Hdr struct {
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}
