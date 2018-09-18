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
	"strings"
	"thumbai/app/models"
	"thumbai/app/store"
)

// All method returns all the proxy configuration from the data store.
func All() map[string][]*models.ProxyRule {
	keys := store.BucketKeys(store.BucketProxies)
	proxies := map[string][]*models.ProxyRule{}
	for _, k := range keys {
		rules := make([]*models.ProxyRule, 0)
		_ = store.Get(store.BucketProxies, k, &rules)
		proxies[k] = rules
	}
	return proxies
}

// Stats method returns stats about proxies.
func Stats() map[string]int {
	stats := make(map[string]int)
	proxies := All()
	stats["Host"] = len(proxies)
	c := 0
	for _, v := range proxies {
		c += len(v)
	}
	stats["ProxyRules"] = c
	return stats
}

// AddHost method adds the given host into proxies data store.
func AddHost(proxyRule *models.ProxyRule) error {
	proxyRule.Host = strings.ToLower(proxyRule.Host)
	if store.IsKeyExists(store.BucketProxies, proxyRule.Host) {
		return store.ErrRecordAlreadyExists
	}
	proxyRule.RequestHeader = nil
	proxyRule.ResponseHeader = nil
	proxyRule.RestrictFile = nil
	return store.Put(store.BucketProxies, proxyRule.Host, append([]*models.ProxyRule{}, proxyRule))
}

// DelHost method deletes the given host from proxies store.
func DelHost(hostName string) error {
	return store.Del(store.BucketProxies, strings.ToLower(hostName))
}

// Get method returns configured proxy rules for the given host.
func Get(host string) []*models.ProxyRule {
	host = strings.ToLower(host)
	rules := make([]*models.ProxyRule, 0)
	_ = store.Get(store.BucketProxies, host, &rules)
	return rules
}
