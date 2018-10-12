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
	"thumbai/app/datastore"
	"thumbai/app/models"

	"aahframe.work"
)

// All method returns all the proxy configuration from the data store.
func All() map[string][]*models.ProxyRule {
	keys := datastore.BucketKeys(datastore.BucketProxies)
	proxies := map[string][]*models.ProxyRule{}
	for _, k := range keys {
		rules := make([]*models.ProxyRule, 0)
		_ = datastore.Get(datastore.BucketProxies, k, &rules)
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

// Import method deletes existing configuration and then saves given configurations.
// Also reload the new imported configurations.
func Import(configs map[string][]*models.ProxyRule) {
	for k, p := range configs {
		if err := DelHost(k); err != nil && err != datastore.ErrRecordNotFound {
			continue
		}
		if err := datastore.Put(datastore.BucketProxies, k, p); err != nil {
			aah.AppLog().Errorf("Unable to import proxy config for host: %s, error: %v", k, err)
		}
	}
	Load(nil)
}

// AddHost method adds the given host into proxies data store.
func AddHost(proxyInfo *models.FormTargetURL) error {
	proxyInfo.Host = strings.ToLower(proxyInfo.Host)
	if datastore.IsKeyExists(datastore.BucketProxies, proxyInfo.Host) {
		return datastore.ErrRecordAlreadyExists
	}
	proxyRule := &models.ProxyRule{Host: proxyInfo.Host, TargetURL: proxyInfo.TargetURL}
	if err := datastore.Put(datastore.BucketProxies, proxyRule.Host, append([]*models.ProxyRule{}, proxyRule)); err != nil {
		return err
	}
	h := Thumbai.AddHost(proxyInfo.Host)
	return h.AddProxyRule(proxyRule)
}

// DelHost method deletes the given host from proxies store.
func DelHost(hostName string) error {
	return datastore.Del(datastore.BucketProxies, strings.ToLower(hostName))
}

// Get method returns configured proxy rules for the given host.
func Get(host string) []*models.ProxyRule {
	host = strings.ToLower(host)
	rules := make([]*models.ProxyRule, 0)
	_ = datastore.Get(datastore.BucketProxies, host, &rules)
	return rules
}

// GetRule method returns configured proxy rules for the given host.
func GetRule(host, targetURL string) *models.ProxyRule {
	rules := Get(host)
	for _, rule := range rules {
		if rule.TargetURL == targetURL {
			return rule
		}
	}
	return nil
}

// AddRule methods adds new proxy rule for the host.
func AddRule(rule *models.ProxyRule) error {
	rules := Get(rule.Host)
	return datastore.Put(datastore.BucketProxies, rule.Host, append(rules, rule))
}

// UpdateRule method updates the given rule on the exiting rules for the host.
func UpdateRule(oldTargetURL string, rule *models.ProxyRule) error {
	rules := Get(rule.Host)
	if rule.Last { // inactived current last rule
		for _, e := range rules {
			e.Last = false
		}
	}
	for i := range rules {
		if rules[i].TargetURL == oldTargetURL {
			rules[i] = rule
			return datastore.Put(datastore.BucketProxies, rule.Host, rules)
		}
	}
	return nil
}

// DelRule method deletes configured proxy rule for the given host.
func DelRule(host, targetURL string) error {
	rules := Get(host)
	if len(rules) == 0 {
		return datastore.ErrRecordNotFound
	}
	f := -1
	for i, r := range rules {
		if r.TargetURL == targetURL {
			f = i
			break
		}
	}
	if f > -1 {
		rules = append(rules[:f], rules[f+1:]...)
		return datastore.Put(datastore.BucketProxies, host, rules)
	}
	return nil
}
