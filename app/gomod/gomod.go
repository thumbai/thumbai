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
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"thumbai/app/models"

	"aahframe.work"
	"aahframe.work/essentials"
)

// errors
var (
	ErrInvalidGoModPath = errors.New("gomod: invalid path")
	ErrExecFailure      = errors.New("gomod: exec failure")
)

// go mod
var (
	Settings = &settings{RWMutex: sync.RWMutex{}, GoVersion: "NA"}
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
	Settings.GoCache = filepath.Join(Settings.GoPath, "pkg", "mod", "cache")
	Settings.ModCachePath = filepath.Join(Settings.GoPath, "pkg", "mod", "cache", "download")

	if ess.IsStrEmpty(Settings.storeSettings.GoProxy) {
		Settings.GoProxy = os.Getenv("GOPROXY")
	} else {
		Settings.GoProxy = Settings.storeSettings.GoProxy
	}

	Settings.Enabled = true

	go func() {
		cnt := Count(Settings.ModCachePath)
		Settings.Lock()
		Settings.Stats.TotalCount = cnt
		_ = SaveStats(Settings.Stats)
		Settings.Unlock()
	}()
}

// FSPathDelimiter is used for mod cache operations.
const FSPathDelimiter = "/@v/"

// InferRequest method parse the go mod request into Request object.
//
// {module}/@v/list fetches a list of all known versions, one per line.
//
// {module}/@v/{version}.info fetches JSON-formatted metadata about that version.
//
// {module}/@v/{version}.mod fetches the go.mod file for that version.
//
// {module}/@v/{version}.zip fetches the zip file for that version.
func InferRequest(modReqPath string) (*Request, error) {
	parts := strings.Split(modReqPath, FSPathDelimiter)
	if len(parts) != 2 {
		return nil, ErrInvalidGoModPath
	}

	req := &Request{Module: parts[0],
		ModuleFilePath: filepath.Join(Settings.ModCachePath, parts[0]),
		FilePath:       filepath.Join(Settings.ModCachePath, modReqPath)}
	if parts[1] == "list" {
		req.Action = parts[1]
	} else {
		i := strings.LastIndexByte(parts[1], '.')
		if i == -1 {
			return nil, ErrInvalidGoModPath
		}
		req.Version = parts[1][:i]
		req.Action = parts[1][i+1:]
	}

	return req, nil
}

var goPrg = []byte(`package main

    import (
        "fmt"
    )
    
    func main() {
        fmt.Println("thumbai temp go project")
    }
	`)

var goMod = `module thumbai.app/tempproject

require %s %s
`

const tempFilePerm = os.FileMode(0644)

var modMutex = map[string]bool{}

// Download method downloads the requested go module using 'go get'
// which populates the mod cache.
func Download(modReq *Request) error {
	app := aah.App()
	mutexModPath := modReq.Module + "@" + modReq.Version
	app.Log().Info("Download request recevied for ", mutexModPath)
	srcZipPath := filepath.Join(Settings.ModCachePath, modReq.Module, "@v", modReq.Version+".zip")
	if ess.IsFileExists(srcZipPath) {
		app.Log().Info("Module ", mutexModPath, " already exists on server")
		return nil
	}
	if _, found := modMutex[mutexModPath]; found {
		// Download already in-progress
		return nil
	}
	defer func() { delete(modMutex, mutexModPath) }()

	decodedPath, err := DecodePath(modReq.Module)
	if err != nil {
		return err
	}
	dirPath, err := ioutil.TempDir("", "tempgoproject")
	if err != nil {
		return err
	}
	defer func() {
		if err = os.RemoveAll(dirPath); err != nil {
			app.Log().Warn(err)
		}
	}()

	if err = ioutil.WriteFile(filepath.Join(dirPath, "main.go"), goPrg, tempFilePerm); err != nil {
		return err
	}
	if err = ioutil.WriteFile(filepath.Join(dirPath, "go.mod"), []byte(fmt.Sprintf(goMod, decodedPath, modReq.Version)), tempFilePerm); err != nil {
		return err
	}

	args := []string{"mod", "download"}
	if modReq.gogetRequired() {
		args = []string{"get", decodedPath + "@" + modReq.Version}
	}
	app.Log().Info("Executing ", Settings.GoBinary, " ", strings.Join(args, " "))
	cmd := exec.Command(Settings.GoBinary, args...)
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOPATH=%s", Settings.GoPath))
	env = append(env, fmt.Sprintf("GOCACHE=%s", Settings.GoCache))
	cmd.Env = env
	cmd.Dir = dirPath

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	scanner := bufio.NewScanner(stderrReader)
	go func() {
		for scanner.Scan() {
			fmt.Fprintln(buf, scanner.Text())
		}
	}()

	status, errInfo := inferExitStatus(cmd, cmd.Run())
	if status != 0 {
		app.Log().Error(strings.TrimSpace(buf.String()))
		app.Log().Error(errInfo)
		return ErrExecFailure
	}
	processModAndUpdateCount(buf)
	app.Log().Infof("Module %s@%s downloaded successfully", decodedPath, modReq.Version)
	return nil
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

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Request type and its methods
//______________________________________________________________________________

// Request struct holds parsed values of module request info.
type Request struct {
	Module         string
	Version        string
	FilePath       string
	ModuleFilePath string
	Action         string
}

func (r *Request) gogetRequired() bool {
	return r.Version == "latest" || r.Version == "master"
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

func processModAndUpdateCount(r io.Reader) {
	Settings.Lock()
	defer Settings.Unlock()
	mods := map[string]bool{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ln := strings.TrimSpace(scanner.Text())
		if len(ln) > 0 {
			if strings.HasPrefix(ln, "go: finding") || strings.HasPrefix(ln, "go: downloading") {
				parts := strings.Fields(ln)[2:]
				if len(parts) >= 2 {
					mods[parts[0]+"@"+parts[1]] = true
				}
			}
		}
	}
	Settings.Stats.TotalCount += int64(len(mods))
	_ = SaveStats(Settings.Stats)
}

func inferGoBinary(current string) (string, error) {
	var currentExists bool
	if !ess.IsStrEmpty(current) {
		currentExists = ess.IsFileExists(current)
	}
	if ess.IsStrEmpty(current) || !currentExists {
		aah.App().Log().Warnf("%s configured go binary is not exists on server, will infer from server if possible", current)
		return exec.LookPath("go")
	}
	return current, nil
}

func inferGopath(current string) string {
	var currentExists bool
	if !ess.IsStrEmpty(current) {
		currentExists = ess.IsFileExists(current)
	}
	if ess.IsStrEmpty(current) || !currentExists {
		aah.App().Log().Warnf("%s GOPATH is not exists on server, will infer from server if possible", current)
		if paths := filepath.SplitList(build.Default.GOPATH); len(paths) > 0 {
			return paths[0]
		}
	}
	return current
}
