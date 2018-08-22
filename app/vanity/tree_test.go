package vanity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorepositree.com/app/models"
)

func TestTreeAddAndLookup(t *testing.T) {
	tt := &Tree{hosts: make(map[string]*node)}

	for _, importPath := range []string{
		"/aah",
		"/cli",
		"/cache/provider/redis",
	} {
		err := tt.add("aahframe.work", importPath, &models.PackageInfo{Path: importPath})
		assert.Nil(t, err)
		// assert.Equal(t, errNodeExists, err)
	}

	// Lookup
	for _, importPath := range []string{
		"/cache/provider/redis",
		"/",
		"/cli",
		"/cli/aah",
		"/aah/vfs",
		"/aah/cache",
		"/unknown",
	} {
		r := tt.lookup("aahframe.work", importPath)
		if r != nil {
			assert.True(t, strings.HasPrefix(importPath, r.Path))
		}
	}
}

func testdataBaseDir() string {
	wd, _ := os.Getwd()
	if idx := strings.Index(wd, ".testdata"); idx > 0 {
		wd = wd[:idx]
	}
	return filepath.Join(wd, ".testdata")
}
