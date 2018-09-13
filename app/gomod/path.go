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
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Module and version path taken from 'go mod' cmd source code
// https://github.com/golang/go/blob/master/src/cmd/go/internal/module/module.go

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package methods
//______________________________________________________________________________

// CheckPath checks that a module path is valid.
func CheckPath(path string) error {
	if err := checkPath(path, false); err != nil {
		return fmt.Errorf("malformed module path %q: %v", path, err)
	}
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	if i == 0 {
		return fmt.Errorf("malformed module path %q: leading slash", path)
	}
	if !strings.Contains(path[:i], ".") {
		return fmt.Errorf("malformed module path %q: missing dot in first path element", path)
	}
	if path[0] == '-' {
		return fmt.Errorf("malformed module path %q: leading dash in first path element", path)
	}
	for _, r := range path[:i] {
		if !firstPathOK(r) {
			return fmt.Errorf("malformed module path %q: invalid char %q in first path element", path, r)
		}
	}
	if _, _, ok := SplitPathVersion(path); !ok {
		return fmt.Errorf("malformed module path %q: invalid version", path)
	}
	return nil
}

// SplitPathVersion returns prefix and major version such that prefix+pathMajor == path
// and version is either empty or "/vN" for N >= 2.
// As a special case, gopkg.in paths are recognized directly;
// they require ".vN" instead of "/vN", and for all N, not just N >= 2.
func SplitPathVersion(path string) (prefix, pathMajor string, ok bool) {
	if strings.HasPrefix(path, "gopkg.in/") {
		return splitGopkgIn(path)
	}

	i := len(path)
	dot := false
	for i > 0 && ('0' <= path[i-1] && path[i-1] <= '9' || path[i-1] == '.') {
		if path[i-1] == '.' {
			dot = true
		}
		i--
	}
	if i <= 1 || path[i-1] != 'v' || path[i-2] != '/' {
		return path, "", true
	}
	prefix, pathMajor = path[:i-2], path[i-2:]
	if dot || len(pathMajor) <= 2 || pathMajor[2] == '0' || pathMajor == "/v1" {
		return path, "", false
	}
	return prefix, pathMajor, true
}

// EncodePath returns the safe encoding of the given module path.
// It fails if the module path is invalid.
func EncodePath(path string) (encoding string, err error) {
	if err := CheckPath(path); err != nil {
		return "", err
	}

	return encodeString(path)
}

// DecodePath returns the module path of the given safe encoding.
// It fails if the encoding is invalid or encodes an invalid path.
func DecodePath(encoding string) (path string, err error) {
	path, ok := decodeString(encoding)
	if !ok {
		return "", fmt.Errorf("invalid module path encoding %q", encoding)
	}
	if err := CheckPath(path); err != nil {
		return "", fmt.Errorf("invalid module path encoding %q: %v", encoding, err)
	}
	return path, nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Unexported methods
//______________________________________________________________________________

// firstPathOK reports whether r can appear in the first element of a module path.
// The first element of the path must be an LDH domain name, at least for now.
// To avoid case ambiguity, the domain name must be entirely lower case.
func firstPathOK(r rune) bool {
	return r == '-' || r == '.' ||
		'0' <= r && r <= '9' ||
		'a' <= r && r <= 'z'
}

// fileNameOK reports whether r can appear in a file name.
// For now we allow all Unicode letters but otherwise limit to pathOK plus a few more punctuation characters.
// If we expand the set of allowed characters here, we have to
// work harder at detecting potential case-folding and normalization collisions.
// See note about "safe encoding" below.
func fileNameOK(r rune) bool {
	if r < utf8.RuneSelf {
		// Entire set of ASCII punctuation, from which we remove characters:
		//     ! " # $ % & ' ( ) * + , - . / : ; < = > ? @ [ \ ] ^ _ ` { | } ~
		// We disallow some shell special characters: " ' * < > ? ` |
		// (Note that some of those are disallowed by the Windows file system as well.)
		// We also disallow path separators / : and \ (fileNameOK is only called on path element characters).
		// We allow spaces (U+0020) in file names.
		const allowed = "!#$%&()+,-.=@[]^_{}~ "
		if '0' <= r && r <= '9' || 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' {
			return true
		}
		for i := 0; i < len(allowed); i++ {
			if rune(allowed[i]) == r {
				return true
			}
		}
		return false
	}
	// It may be OK to add more ASCII punctuation here, but only carefully.
	// For example Windows disallows < > \, and macOS disallows :, so we must not allow those.
	return unicode.IsLetter(r)
}

// pathOK reports whether r can appear in an import path element.
// Paths can be ASCII letters, ASCII digits, and limited ASCII punctuation: + - . _ and ~.
// This matches what "go get" has historically recognized in import paths.
// TODO(rsc): We would like to allow Unicode letters, but that requires additional
// care in the safe encoding (see note below).
func pathOK(r rune) bool {
	if r < utf8.RuneSelf {
		return r == '+' || r == '-' || r == '.' || r == '_' || r == '~' ||
			'0' <= r && r <= '9' ||
			'A' <= r && r <= 'Z' ||
			'a' <= r && r <= 'z'
	}
	return false
}

// checkPath checks that a general path is valid.
// It returns an error describing why but not mentioning path.
// Because these checks apply to both module paths and import paths,
// the caller is expected to add the "malformed ___ path %q: " prefix.
// fileName indicates whether the final element of the path is a file name
// (as opposed to a directory name).
func checkPath(path string, fileName bool) error {
	if !utf8.ValidString(path) {
		return fmt.Errorf("invalid UTF-8")
	}
	if path == "" {
		return fmt.Errorf("empty string")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("double dot")
	}
	if strings.Contains(path, "//") {
		return fmt.Errorf("double slash")
	}
	if path[len(path)-1] == '/' {
		return fmt.Errorf("trailing slash")
	}
	elemStart := 0
	for i, r := range path {
		if r == '/' {
			if err := checkElem(path[elemStart:i], fileName); err != nil {
				return err
			}
			elemStart = i + 1
		}
	}
	return checkElem(path[elemStart:], fileName)
}

// checkElem checks whether an individual path element is valid.
// fileName indicates whether the element is a file name (not a directory name).
func checkElem(elem string, fileName bool) error {
	if elem == "" {
		return fmt.Errorf("empty path element")
	}
	if strings.Count(elem, ".") == len(elem) {
		return fmt.Errorf("invalid path element %q", elem)
	}
	if elem[0] == '.' && !fileName {
		return fmt.Errorf("leading dot in path element")
	}
	if elem[len(elem)-1] == '.' {
		return fmt.Errorf("trailing dot in path element")
	}
	charOK := pathOK
	if fileName {
		charOK = fileNameOK
	}
	for _, r := range elem {
		if !charOK(r) {
			return fmt.Errorf("invalid char %q", r)
		}
	}

	// Windows disallows a bunch of path elements, sadly.
	// See https://docs.microsoft.com/en-us/windows/desktop/fileio/naming-a-file
	short := elem
	if i := strings.Index(short, "."); i >= 0 {
		short = short[:i]
	}
	for _, bad := range badWindowsNames {
		if strings.EqualFold(bad, short) {
			return fmt.Errorf("disallowed path element %q", elem)
		}
	}
	return nil
}

// badWindowsNames are the reserved file path elements on Windows.
// See https://docs.microsoft.com/en-us/windows/desktop/fileio/naming-a-file
var badWindowsNames = []string{
	"CON",
	"PRN",
	"AUX",
	"NUL",
	"COM1",
	"COM2",
	"COM3",
	"COM4",
	"COM5",
	"COM6",
	"COM7",
	"COM8",
	"COM9",
	"LPT1",
	"LPT2",
	"LPT3",
	"LPT4",
	"LPT5",
	"LPT6",
	"LPT7",
	"LPT8",
	"LPT9",
}

// splitGopkgIn is like SplitPathVersion but only for gopkg.in paths.
func splitGopkgIn(path string) (prefix, pathMajor string, ok bool) {
	if !strings.HasPrefix(path, "gopkg.in/") {
		return path, "", false
	}
	i := len(path)
	if strings.HasSuffix(path, "-unstable") {
		i -= len("-unstable")
	}
	for i > 0 && ('0' <= path[i-1] && path[i-1] <= '9') {
		i--
	}
	if i <= 1 || path[i-1] != 'v' || path[i-2] != '.' {
		// All gopkg.in paths must end in vN for some N.
		return path, "", false
	}
	prefix, pathMajor = path[:i-2], path[i-2:]
	if len(pathMajor) <= 2 || pathMajor[2] == '0' && pathMajor != ".v0" {
		return path, "", false
	}
	return prefix, pathMajor, true
}

func encodeString(s string) (encoding string, err error) {
	haveUpper := false
	for _, r := range s {
		if r == '!' || r >= utf8.RuneSelf {
			// This should be disallowed by CheckPath, but diagnose anyway.
			// The correctness of the encoding loop below depends on it.
			return "", fmt.Errorf("internal error: inconsistency in EncodePath")
		}
		if 'A' <= r && r <= 'Z' {
			haveUpper = true
		}
	}

	if !haveUpper {
		return s, nil
	}

	var buf []byte
	for _, r := range s {
		if 'A' <= r && r <= 'Z' {
			buf = append(buf, '!', byte(r+'a'-'A'))
		} else {
			buf = append(buf, byte(r))
		}
	}
	return string(buf), nil
}

func decodeString(encoding string) (string, bool) {
	var buf []byte

	bang := false
	for _, r := range encoding {
		if r >= utf8.RuneSelf {
			return "", false
		}
		if bang {
			bang = false
			if r < 'a' || 'z' < r {
				return "", false
			}
			buf = append(buf, byte(r+'A'-'a'))
			continue
		}
		if r == '!' {
			bang = true
			continue
		}
		if 'A' <= r && r <= 'Z' {
			return "", false
		}
		buf = append(buf, byte(r))
	}
	if bang {
		return "", false
	}
	return string(buf), true
}
