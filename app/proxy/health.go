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

package proxy

import (
	"net/http"
	"net/http/httptest"
	"sync"

	"aahframe.work"
	"aahframe.work/ahttp"
)

func Health(ctx *aah.Context) []aah.Data {
	Thumbai.RLock()
	defer Thumbai.RUnlock()
	hosts := make([]aah.Data, 0)
	wg := sync.WaitGroup{}
	wg.Add(len(Thumbai.Hosts))
	for _, h := range Thumbai.Hosts {
		go func(host *host) {
			defer wg.Done()
			data := checkHealth(host)
			hosts = append(hosts, data)
		}(h)
	}
	wg.Wait()
	return hosts
}

func checkHealth(h *host) aah.Data {
	reqPath := h.HealthCheckPath
	if len(reqPath) == 0 {
		reqPath = "/"
	}
	r, _ := http.NewRequest(http.MethodGet, reqPath, nil)
	r.Header.Set(ahttp.HeaderHost, h.Name)
	r.Host = h.Name

	w := httptest.NewRecorder()
	h.LastRule.Proxy.ServeHTTP(w, r)
	code := w.Result().StatusCode
	if code >= http.StatusOK && code < http.StatusBadRequest {
		return aah.Data{
			"status":      "available",
			"host":        h.Name,
			"status_code": code,
		}
	}

	return aah.Data{
		"status":      "unavailable",
		"host":        h.Name,
		"status_code": code,
	}
}
