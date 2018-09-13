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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"aahframe.work/aah"
	"aahframe.work/aah/essentials"
	"aahframe.work/aah/log"
)

// errors
var (
	ErrInvalidGoModPath = errors.New("gomod: invalid path")
	ErrExecFailure      = errors.New("gomod: exec failure")
)

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

var (
	gopath   string
	modCache string
	gocmd    string
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// Infer method runtime infer values for gocmd, gopath, and mod cache.
func Infer(_ *aah.Event) {
	var err error
	if gocmd, err = exec.LookPath("go"); err != nil {
		log.Fatal(err)
	}
	paths := filepath.SplitList(build.Default.GOPATH)
	if len(paths) > 0 {
		gopath = paths[0]
	}
	modCache = filepath.Join(gopath, "pkg", "mod", "cache", "download")
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
		ModuleFilePath: filepath.Join(modCache, parts[0]),
		FilePath:       filepath.Join(modCache, modReqPath)}
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
	mutexModPath := modReq.Module + "@" + modReq.Version
	aah.AppLog().Info("Download request recevied for ", mutexModPath)
	srcZipPath := filepath.Join(modCache, modReq.Module, "@v", modReq.Version+".zip")
	if ess.IsFileExists(srcZipPath) {
		aah.AppLog().Info("Module ", mutexModPath, " already exists on server")
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
			aah.AppLog().Warn(err)
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
		args = []string{"get", "-v", decodedPath + "@" + modReq.Version}
	}
	aah.AppLog().Info("Executing ", gocmd, " ", strings.Join(args, " "))
	cmd := exec.Command(gocmd, args...)
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOPATH=%s", gopath))
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
		aah.AppLog().Info(buf.String())
		aah.AppLog().Error(errInfo)
		return ErrExecFailure
	}
	fmt.Println(buf.String())

	aah.AppLog().Infof("Module %s@%s downloaded successfully", decodedPath, modReq.Version)
	return nil
}

const modExt = ".mod"

// Count method counts the no of modules in the server filesystem.
func Count(dir string) int64 {
	if ess.IsStrEmpty(dir) {
		dir = modCache
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
