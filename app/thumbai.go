package main

import (
	"path/filepath"
	"strings"

	"aahframe.work/aah"
	"aahframe.work/aah/log"
)

// CheckConfig method subscribes to aah `OnInit` event to check config
// and puts default values as needed.
func CheckConfig(e *aah.Event) {
	cfg := aah.AppConfig()
	cfg.SetString("env.active", cfg.StringDefault("thumbai.env.active", "prod"))
	if !cfg.IsExists("thumbai.admin.host") {
		log.Fatalf("'thumbai.admin.host' value is not configured")
	}

	if tocfg, found := cfg.GetSubConfig("thumbai.server"); found {
		if err := cfg.Merge2Section("env.prod.server", tocfg); err != nil {
			log.Error(err)
		}
	} else {
		log.Errorf("'thumbai.server' configuration not found")
	}

	if tocfg, found := cfg.GetSubConfig("thumbai.log"); found {
		if err := cfg.Merge2Section("env.prod.log", tocfg); err != nil {
			log.Error(err)
		}
	} else {
		log.Errorf("'thumbai.log' configuration not found")
	}

	cfg.SetString("env.prod.server.header", "thumbai "+aah.AppBuildInfo().Version)
	adminHost := cfg.StringDefault("thumbai.admin.host", "")
	if i := strings.IndexByte(adminHost, ':'); i > 0 {
		cfg.SetString("env.prod.routes.domains.thumbai.port", adminHost[i+1:])
		adminHost = adminHost[:i]
	}
	cfg.SetString("env.prod.routes.domains.thumbai.host", adminHost)

	if !cfg.IsExists("thumbai.admin.data_store.location") {
		cfg.SetString("thumbai.admin.data_store.location", filepath.Join(aah.AppBaseDir(), "data"))
	}
}
