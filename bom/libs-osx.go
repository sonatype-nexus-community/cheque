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
	// "github.com/sonatype-nexus-community/cheque/logger"

	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	// Required to run external commands
	"os/exec"
)

/** Given a file path, extract a library name from the path.
 */
func getOsxLibraryName(name string) (path string, err error) {
	path, _, err = getOsxLibraryNameAndVersion(name)
	return path, err
}

/** Given a file path, extract a library name from the path.
 */
func getOsxLibraryVersion(name string) (version string, err error) {
	_, version, err = getOsxLibraryNameAndVersion(name)
	return version, err
}

func getOsxLibraryNameAndVersion(path string) (name string, version string, err error) {

	// Extract a name
	fname := filepath.Base(path)
	r, _ := regexp.Compile("^(.*?)\\.([0-9\\.]+)dylib")
	matches := r.FindStringSubmatch(fname)
	if matches == nil {
		return "", "", errors.New("getOsxLibraryNameAndVersion: cannot get name/version from " + path + " (" + fname + ")")
	}
	name = matches[1]

	// Extract a version
	r, _ = regexp.Compile("\\.([0-9\\.]+)\\.dylib")
	matches = r.FindStringSubmatch(fname)
	if matches != nil {
		return name, matches[1], nil
	}

	return name, "", errors.New("getOsxLibraryNameAndVersion: cannot get version from " + fname)
}

/** Run otool and get version strings from it
 *
 * 	/usr/lib/libpam.2.dylib (compatibility version 3.0.0, current version 3.0.0)
 */
func getOtoolVersion(name string, file string) (version string, err error) {
	outbytes, err := exec.Command("otool", "-L", file).Output()
	if err != nil {
		return "", err
	}
	out := string(outbytes)

	lines := strings.Split(out, "\n")

	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "* %s .. %s\n", name, line)
	}

	return "", nil
}

func findOsxLibFile(libPaths []string, name string) (match string, err error) {
	if strings.HasSuffix(name, ".dylib") {
		// First check if this is an absolute path
		_, err = AppFs.Stat(name)
		if !os.IsNotExist((err)) {
			return name, nil
		}

		return findLibFile(libPaths, "", name, "")
	}
	return findLibFile(libPaths, "lib", name, ".dylib")
}

func getOsxSymlinkVersion(file string) (version string, err error) {
	path, err := filepath.EvalSymlinks(file)

	if err != nil {
		return "", err
	}

	// Extract a version
	r, err := regexp.Compile("\\.([0-9\\.]+)\\.dylib")
	if err != nil {
		return "", err
	}
	matches := r.FindStringSubmatch(path)
	if matches == nil {
		return "", nil
	}

	return matches[1], nil
}

func getOsxLibraryPathRegexPattern() (result string) {
	return "[a-zA-Z0-9_/\\.]+\\.dylib"
}

func getOsxLibraryFileRegexPattern() (result string) {
	return "([a-zA-Z0-9_]+)\\.[0-9\\.]+\\.dylib"
}

func getOsxLibPaths() (paths []string) {
	// clang -Xlinker -v
	clangCmd := exec.Command("clang", "-Xlinker", "-v")

	stderr, err := clangCmd.StderrPipe()
	if err != nil {
		// log.Fatalf("could not get stderr pipe: %v", err)
	}
	stdout, err := clangCmd.StdoutPipe()
	if err != nil {
		// log.Fatalf("could not get stdout pipe: %v", err)
	}
	go func() {
		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)
		matching := false
		for scanner.Scan() {
			msg := scanner.Text()
			if matching {
				if strings.HasPrefix(msg, "\t") {
					paths = append(paths, strings.TrimSpace(msg))
				} else {
					matching = false
				}
			}
			if strings.HasPrefix(msg, "Library search paths") {
				matching = true
			}
		}
	}()
	if err := clangCmd.Run(); err != nil {
		// log.Fatalf("could not run clangCmd: %v", err)
	}
	if err != nil {
		// log.Fatalf("could not wait for clangCmd: %v", err)
	}
	return paths
}
