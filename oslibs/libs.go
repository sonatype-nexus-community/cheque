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
package oslibs

import (
	// "github.com/sonatype-nexus-community/cheque/logger"
  "errors"
	"runtime"
	"os"
	"fmt"
)

func GetLibraryPath(libPaths []string, name string) (path string, err error) {
	switch (runtime.GOOS) {
		case "windows":
				_, _ = fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", runtime.GOOS)
				os.Exit(2)
				return name, errors.New("GetLibraryPath: Unsupported OS")

		case "darwin":
			file, err := findOsxLibFile(libPaths, name)
			if (err != nil || file == "") {
				return file, errors.New("GetLibraryPath: Cannot find path to " + name)
			}
			return file, err

		default:
			file, err := findUnixLibFile(libPaths, name)
			if (err != nil || file == "") {
				return file, errors.New("GetLibraryPath: Cannot find path to " + name)
			}
			return file, err
	}
}

func GetLibraryName(name string) (path string, err error) {
	switch (runtime.GOOS) {
		case "windows":
				_, _ = fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", runtime.GOOS)
				os.Exit(2)
				return name, errors.New("GetLibraryName: Unsupported OS")

		case "darwin":
			return getOsxLibraryName(name)

		default:
			return getUnixLibraryName(name)
	}
}

func GetLibraryVersion(name string) (path string, err error) {
	switch (runtime.GOOS) {
		case "windows":
				_, _ = fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", runtime.GOOS)
				os.Exit(2)
				return name, errors.New("GetLibraryVersion: Unsupported OS")

		case "darwin":
			return getOsxLibraryVersion(name)

		default:
			return getUnixLibraryVersion(name)
	}
}

func GetLibraryPathRegexPattern() (result string) {

	if runtime.GOOS == "windows" {
  }

	if runtime.GOOS == "darwin" {
    return getOsxLibraryPathRegexPattern()
  }

	// Fall back to unix variant
  return getUnixLibraryPathRegexPattern()
}


func GetArchiveFileRegexPattern() (result string) {

	if runtime.GOOS == "windows" {
  }

	if runtime.GOOS == "darwin" {
  }

	// Fall back to unix variant
  return getUnixArchiveFileRegexPattern()
}

func GetLibraryFileRegexPattern() (result string) {

	if runtime.GOOS == "windows" {
  }

	if runtime.GOOS == "darwin" {
    return getOsxLibraryFileRegexPattern()
  }

	// Fall back to unix variant
  return getUnixLibraryFileRegexPattern()
}

/** FIXME: Actually search paths to find actual binary
 */
func GetCommandPath(cmd string) (path string) {
	path = "/usr/bin/" + cmd;
	if _, err := os.Stat(path); os.IsNotExist(err) {
	  return ""
	}
	return path;
}

func GetLibPaths() (paths []string) {
	if runtime.GOOS == "windows" {
		return paths
  }

	if runtime.GOOS == "darwin" {
    return getOsxLibPaths()
  }

	// Fall back to unix variant
  return getLinuxLibPaths()
}
