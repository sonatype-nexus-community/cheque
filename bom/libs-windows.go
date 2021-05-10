// Copyright 2019 Sonatype Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bom

import (
	// "os"
	// "regexp"
	// "strings"

	// "github.com/sonatype-nexus-community/cheque/logger"
	// "github.com/sonatype-nexus-community/nancy/types"

	// // "fmt"
	// "bufio"
	// "errors"
	"errors"
	"path/filepath"
	"regexp"
	// "os/exec"
	// "bytes"
)

/** Given a file path, extract a library name from the path.
 */
func getWindowsLibraryName(name string) (path string, err error) {
	path, _, err = getWindowsLibraryNameAndVersion(name)
	return path, err
}

/** Given a file path, extract a library name from the path.
 */
func getWindowsLibraryVersion(name string) (version string, err error) {
	_, version, err = getWindowsLibraryNameAndVersion(name)
	return version, err
}

func getWindowsLibraryNameAndVersion(path string) (name string, version string, err error) {
	// Extract a name
	fname := filepath.Base(path)
	r, _ := regexp.Compile("^(.*?)[\\.\\-_]([0-9\\._\\-]+)\\.dll")
	matches := r.FindStringSubmatch(path)
	if matches == nil {
		return "", "", errors.New("Cannot get name/version from " + path + " (" + fname + ")")
	}
	name = matches[1]

	// Extract a version
	r, _ = regexp.Compile("[\\._\\-]([0-9\\._\\-]+)\\.dll")
	matches = r.FindStringSubmatch(path)
	if matches != nil {
		return name, matches[1], nil
	}

	return name, "", errors.New("Cannot get version from " + fname)
}
