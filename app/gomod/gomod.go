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

package gomod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"thumbai/app/models"
	"time"

	"aahframe.work"
	"aahframe.work/essentials"
)

// errors
var (
	ErrInvalidGoModPath = errors.New("gomod: invalid path")
	ErrGoModNotExist    = errors.New("gomod: mod not exist")
	ErrExecFailure      = errors.New("gomod: exec failure")
)

// go mod
var (
	Settings = &settings{RWMutex: sync.RWMutex{}, GoVersion: "NA"}

	semverPrefixRegex = regexp.MustCompile(`(^v[0-9]+\.)`)
)

type settings struct {
	sync.RWMutex
	Enabled       bool
	GoBinary      string
	GoVersion     string
	GoPath        string
	GoCache       string
	GoProxy       string
	ModCachePath  string
	Stats         *models.ModuleStats
	storeSettings *models.ModuleSettings
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Infer method runtime infer values for gocmd, gopath, and mod cache.
func Infer(_ *aah.Event) {
	Settings.Lock()
	defer Settings.Unlock()
	Settings.storeSettings = GetSettings()
	Settings.Stats = &models.ModuleStats{}
	var err error
	Settings.GoBinary, err = inferGoBinary(Settings.storeSettings.GoBinary)
	if err != nil {
		aah.App().Log().Errorf("Go modules proxy server would unavailable: %v", err)
		return
	}

	Settings.GoVersion = GoVersion(Settings.GoBinary)
	if !InferGo111AndAbove(Settings.GoVersion) {
		aah.App().Log().Errorf("Go version found: %s. Minimum go1.11 & above is required to use go modules proxy server")
		return
	}

	Settings.GoPath = inferGopath(Settings.storeSettings.GoPath)
	Settings.GoCache = inferGoCache()
	Settings.ModCachePath = filepath.Join(Settings.GoPath, "pkg", "mod", "cache", "download")

	if ess.IsStrEmpty(Settings.storeSettings.GoProxy) {
		Settings.GoProxy = os.Getenv("GOPROXY")
	} else {
		Settings.GoProxy = Settings.storeSettings.GoProxy
	}

	Settings.Enabled = true
	go countGoMods()
}

// FSPathDelimiter is used for mod cache operations.
const FSPathDelimiter = "/@v/"

// Module struct to parse JSON output of command `go mod` plus THUMBAI needs.
type Module struct {
	Path        string
	DecodedPath string
	Version     string
	Error       string
	Info        string
	GoMod       string
	Zip         string
	Dir         string
	Action      string
}

// InferRequest method parse the go mod request into Request object.
//
// {module}/@v/list fetches a list of all known versions, one per line.
//
// {module}/@v/{version}.info fetches JSON-formatted metadata about that version.
//
// {module}/@v/{version}.mod fetches the go.mod file for that version.
//
// {module}/@v/{version}.zip fetches the zip file for that version.
func InferRequest(modReqPath string) (*Module, error) {
	parts := strings.Split(modReqPath, FSPathDelimiter)
	if len(parts) != 2 {
		return nil, ErrInvalidGoModPath
	}

	mod := &Module{Path: parts[0]}
	if parts[1] == "list" {
		mod.Action = parts[1]
		if !ess.IsFileExists(filepath.Join(Settings.ModCachePath, modReqPath)) {
			return mod, ErrGoModNotExist
		}
		return mod, nil
	}

	i := strings.LastIndexByte(parts[1], '.')
	if i == -1 {
		return nil, ErrInvalidGoModPath
	}
	mod.Version = parts[1][:i]
	mod.Action = parts[1][i+1:]
	if !ess.IsFileExists(filepath.Join(Settings.ModCachePath, modReqPath)) {
		if err := checkAndCreateInfoFile(mod); err == nil {
			return mod, nil // good to go
		}
		return mod, ErrGoModNotExist
	}
	return mod, nil
}

var goPrg = []byte(`package main

import (
	"fmt"
)

func main() {
	fmt.Println("thumbai temp go project")
}
`)

var goMod = []byte(`module thumbai.app/tempproject`)

const tempFilePerm = os.FileMode(0644)

var modMutex = map[string]bool{}

// Download method downloads the requested go module path using 'go mod' or 'go get'
// which populates the mod cache.
func Download(mod *Module) (*Module, error) {
	defer countGoMods()
	app := aah.App()
	mutexModPath := modPath(mod)
	app.Log().Info("Download request recevied for ", mutexModPath)
	srcZipPath := filepath.Join(Settings.ModCachePath, mod.Path, "@v", mod.Version+".zip")
	if ess.IsFileExists(srcZipPath) {
		app.Log().Info("Module ", mutexModPath, " already exists on repository")
		return nil, nil
	}
	if _, found := modMutex[mutexModPath]; found {
		app.Log().Infof("Download already in-progress for '%s'", mutexModPath)
		return nil, nil
	}
	defer func() {
		delete(modMutex, mutexModPath)
	}()

	var err error
	mod.DecodedPath, err = DecodePath(mod.Path)
	if err != nil {
		return nil, err
	}

	dirPath, err := createTempProject()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = os.RemoveAll(dirPath); err != nil {
			app.Log().Warn(err)
		}
	}()

	downloadMode := "gomod"
	if ess.IsStrEmpty(mod.Version) || !semverPrefixRegex.MatchString(mod.Version) {
		downloadMode = "goget" // such as version => `latest`, `branchname`
	}

	args := []string{}
	switch downloadMode {
	case "gomod":
		args = append(args, "mod", "download", "-json", decodedModPath(mod))
	case "goget":
		args = append(args, "get", decodedModPath(mod))
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("GOPATH=%s", Settings.GoPath))
	env = append(env, fmt.Sprintf("GOCACHE=%s", Settings.GoCache))

	cmd := exec.Command(Settings.GoBinary, args...)
	cmd.Env = env
	cmd.Dir = dirPath
	stdOut, stdErr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdOut, stdErr

	app.Log().Info("Executing ", Settings.GoBinary, " ", strings.Join(args, " "))
	status, errInfo := inferExitStatus(cmd, cmd.Run())
	if status != 0 {
		app.Log().Error(strings.TrimSpace(stdErr.String()))
		app.Log().Error(errInfo)
		return nil, ErrExecFailure
	}

	// handle result based on mode
	resultMod := &Module{}
	switch downloadMode {
	case "gomod":
		if err = json.NewDecoder(stdOut).Decode(resultMod); err != nil {
			return nil, errors.New(stdErr.String())
		}
		if len(resultMod.Error) > 0 {
			return nil, errors.New(resultMod.Error)
		}
		resultMod.Path = mod.Path
		resultMod.DecodedPath = mod.DecodedPath
		resultMod.Action = mod.Action
	case "goget":
		modVersion := inferGoGetModVersion(mod, stdErr.Bytes())
		resultMod = mod
		resultMod.Version = modVersion
		if ess.IsStrEmpty(resultMod.Version) {
			tcmd := exec.Command(Settings.GoBinary, "mod", "download", "-json", decodedModPath(mod))
			tcmd.Env = env
			tcmd.Dir = dirPath
			b, err := tcmd.Output()
			if err == nil {
				resultMod = &Module{}
				if err = json.NewDecoder(bytes.NewReader(b)).Decode(resultMod); err != nil {
					return nil, err
				}
				if len(resultMod.Error) > 0 {
					return nil, errors.New(resultMod.Error)
				}
				resultMod.Path = mod.Path
				resultMod.DecodedPath = mod.DecodedPath
				resultMod.Action = mod.Action
			}
		}
		_ = checkAndCreateInfoFile(resultMod)
	}

	app.Log().Infof("Module [%s@%s] downloaded successfully into repository", resultMod.Path, resultMod.Version)
	return resultMod, nil
}

const modExt = ".mod"

// Count method counts the no of modules in the server filesystem.
func Count(dir string) int64 {
	if ess.IsStrEmpty(dir) {
		dir = Settings.ModCachePath
	}
	var count int64
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == modExt {
			count++
		}
		return nil
	})
	return count
}

// GoVersion method returns go version
func GoVersion(gocmd string) string {
	cmd := exec.Command(gocmd, "version")
	verBytes, err := cmd.CombinedOutput()
	if err != nil {
		aah.App().Log().Errorf("Unable to infer go version: %v", err)
		return "0.0.0"
	}
	return strings.TrimPrefix(strings.Fields(string(verBytes))[2], "go")
}

// InferGo111AndAbove method infers the go version is go1.11 and above
func InferGo111AndAbove(ver string) bool {
	ver = strings.Join(strings.Split(ver, ".")[:2], ".")
	verNum, err := strconv.ParseFloat(ver, 64)
	if err != nil {
		return false
	}
	return verNum >= float64(1.11)
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Unexported methods
//______________________________________________________________________________

func inferExitStatus(cmd *exec.Cmd, err error) (int, string) {
	if err == nil {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return ws.ExitStatus(), ""
	}
	if ee, ok := err.(*exec.ExitError); ok {
		ws := ee.Sys().(syscall.WaitStatus)
		return ws.ExitStatus(), ee.String()
	}
	return 1, err.Error()
}

const (
	modVersionTimeFormat  = "20060102150405"
	modInfoFileTimeFormat = "2006-01-02T15:04:05Z"
)

func checkAndCreateInfoFile(mod *Module) error {
	infoFile := filepath.Join(Settings.ModCachePath, mod.Path, "@v", mod.Version+".info")
	if ess.IsFileExists(infoFile) {
		return errors.New(".info file already exist")
	}
	if !ess.IsFileExists(filepath.Join(Settings.ModCachePath, mod.Path, "@v", mod.Version+".mod")) {
		return errors.New(".mod file not exist")
	}

	if verParts := strings.Split(mod.Version, "-"); len(verParts) == 3 {
		mtime := verParts[1]
		if i := strings.IndexByte(verParts[1], '.'); i > 0 {
			mtime = verParts[1][i+1:]
		}
		if t, err := time.Parse(modVersionTimeFormat, mtime); err == nil {
			var buf bytes.Buffer
			if err = json.NewEncoder(&buf).Encode(map[string]string{
				"Version": mod.Version,
				"Time":    t.Format(modInfoFileTimeFormat),
			}); err != nil {
				return errors.New("unable to create info JSON")
			}
			fname := filepath.Join(Settings.ModCachePath, mod.Path, "@v", mod.Version+".info")
			aah.App().Log().Debug("Creating ", fname)
			if err = ioutil.WriteFile(fname, buf.Bytes(), 0600); err == nil {
				return nil // good to go
			}
		}
	}
	return errors.New("create '.info' file is not applicable")
}

func modPath(mod *Module) string {
	if ess.IsStrEmpty(mod.Version) {
		return mod.Path + "@latest"
	}
	return mod.Path + "@" + mod.Version
}

func decodedModPath(mod *Module) string {
	if ess.IsStrEmpty(mod.Version) {
		return mod.DecodedPath + "@latest"
	}
	return mod.DecodedPath + "@" + mod.Version
}

func createTempProject() (string, error) {
	dirPath, err := ioutil.TempDir("", "tempgoproject")
	if err != nil {
		return "", err
	}
	if err = ioutil.WriteFile(filepath.Join(dirPath, "main.go"), goPrg, tempFilePerm); err != nil {
		return "", err
	}
	if err = ioutil.WriteFile(filepath.Join(dirPath, "go.mod"), goMod, tempFilePerm); err != nil {
		return "", err
	}
	return dirPath, nil
}

func countGoMods() {
	cnt := Count(Settings.ModCachePath)
	Settings.Lock()
	Settings.Stats.TotalCount = cnt
	_ = SaveStats(Settings.Stats)
	Settings.Unlock()
}

func inferGoGetModVersion(mod *Module, b []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(b))
	prefix := fmt.Sprintf("go: downloading %s", mod.Path)
	for scanner.Scan() {
		ln := strings.TrimSpace(scanner.Text())
		if len(ln) > 0 {
			if strings.HasPrefix(ln, prefix) {
				parts := strings.Fields(ln)
				if len(parts) >= 4 {
					return parts[3]
				}
				return ""
			}
		}
	}
	return ""
}

func inferGoBinary(current string) (string, error) {
	app := aah.App()
	if !ess.IsStrEmpty(current) {
		if !ess.IsFileExists(current) {
			app.Log().Warnf("%s configured Go binary is not exists on server, will infer from server if possible", current)
			return exec.LookPath("go")
		}
		return current, nil
	}

	app.Log().Warn("Go binary is not defined within THUMBAI, will infer from server environment, if possible")
	return exec.LookPath("go")
}

func inferGopath(current string) string {
	app := aah.App()
	if !ess.IsStrEmpty(current) {
		if !ess.IsFileExists(current) {
			app.Log().Warnf("GOPATH: %s directory is not exists on server, will create it", current)
			if err := ess.MkDirAll(current, 0755); err != nil {
				app.Log().Error(err)
			}
		}
		return current
	}

	app.Log().Warn("GOPATH value is not defined within THUMBAI, will infer from server environment, if possible")
	if paths := filepath.SplitList(build.Default.GOPATH); len(paths) > 0 {
		app.Log().Infof("Inferred GOPATH is '%s'", paths[0])
		return paths[0]
	}
	return current
}

func inferGoCache() string {
	cmd := exec.Command(Settings.GoBinary, "env", "GOCACHE")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return filepath.Join(Settings.GoPath, "pkg", "go-build-cache")
	}
	return string(b)
}
