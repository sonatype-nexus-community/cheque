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
	"fmt"
	"runtime"

	"github.com/spf13/afero"
)

// AppFs is using afero to wrap os, that way we can switch it out for testing
var AppFs afero.Fs

// Goose I'm Goose, you're Maverick
var Goose = runtime.GOOS

func init() {
	AppFs = afero.NewOsFs()
}

// GetLibraryPath depending on your operating system (see Goose), returns the path to your library if it exists
func GetLibraryPath(libPaths []string, name string) (path string, err error) {
	switch Goose {
	case "windows":
		panic(fmt.Sprintf("GetLibraryPath: Unsupported OS: %s\n", Goose))
	case "darwin":
		file, err := findOsxLibFile(libPaths, name)
		if err != nil || file == "" {
			return file, fmt.Errorf("GetLibraryPath: Cannot find path to %s", name)
		}
		return file, err
	default:
		file, err := findUnixLibFile(libPaths, name)
		if err != nil || file == "" {
			return file, fmt.Errorf("GetLibraryPath: Cannot find path to %s", name)
		}
		return file, err
	}
}

// GetLibraryName depending on your operating system (see Goose), returns the name of your library, given a full name
func GetLibraryName(path string) (nam string, err error) {
	switch Goose {
	case "windows":
		panic(fmt.Sprintf("GetLibraryName: Unsupported OS: %s\n", Goose))
	case "darwin":
		return getOsxLibraryName(path)
	default:
		return getUnixLibraryName(path)
	}
}

// GetLibraryVersion depending on your operating system (see Goose), returns the version of your library, given a full name
func GetLibraryVersion(path string) (version string, err error) {
	switch Goose {
	case "windows":
		panic(fmt.Sprintf("GetLibraryVersion: Unsupported OS: %s\n", Goose))
	case "darwin":
		return getOsxLibraryVersion(path)
	default:
		return getUnixLibraryVersion(path)
	}
}

func GetLibraryPathRegexPattern() (result string) {
	switch Goose {
	case "darwin":
		return getOsxLibraryPathRegexPattern()
	case "windows":
		return ""
	default:
		return getUnixLibraryPathRegexPattern()
	}
}

func GetArchiveFileRegexPattern() (result string) {
	switch Goose {
	case "darwin":
		return ""
	case "windows":
		return ""
	default:
		return getUnixArchiveFileRegexPattern()
	}
}

func GetLibraryFileRegexPattern() (result string) {
	switch Goose {
	case "darwin":
		return getOsxLibraryFileRegexPattern()
	case "windows":
		return ""
	default:
		return getUnixLibraryFileRegexPattern()
	}
}

func GetLibPaths() (paths []string) {
	switch Goose {
	case "darwin":
		return getOsxLibPaths()
	case "windows":
		return paths
	default:
		return getLinuxLibPaths()
	}
}
