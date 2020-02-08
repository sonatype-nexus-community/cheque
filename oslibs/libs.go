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
	"fmt"
	"runtime"
)

func GetLibraryPath(libPaths []string, name string) (path string, err error) {
	switch operating := runtime.GOOS; operating {
	case "windows":
		panic(fmt.Sprintf("GetLibraryPath: Unsupported OS: %s\n", operating))
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

func GetLibraryName(name string) (path string, err error) {
	switch operating := runtime.GOOS; operating {
	case "windows":
		panic(fmt.Sprintf("GetLibraryName: Unsupported OS: %s\n", operating))
	case "darwin":
		return getOsxLibraryName(name)
	default:
		return getUnixLibraryName(name)
	}
}

func GetLibraryVersion(name string) (path string, err error) {
	switch operating := runtime.GOOS; operating {
	case "windows":
		panic(fmt.Sprintf("GetLibraryVersion: Unsupported OS: %s\n", operating))
	case "darwin":
		return getOsxLibraryVersion(name)
	default:
		return getUnixLibraryVersion(name)
	}
}

func GetLibraryPathRegexPattern() (result string) {
	switch os := runtime.GOOS; os {
	case "darwin":
		return getOsxLibraryPathRegexPattern()
	case "windows":
		return ""
	default:
		return getUnixLibraryPathRegexPattern()
	}
}

func GetArchiveFileRegexPattern() (result string) {
	switch os := runtime.GOOS; os {
	case "darwin":
		return ""
	case "windows":
		return ""
	default:
		return getUnixArchiveFileRegexPattern()
	}
}

func GetLibraryFileRegexPattern() (result string) {
	switch os := runtime.GOOS; os {
	case "darwin":
		return getOsxLibraryFileRegexPattern()
	case "windows":
		return ""
	default:
		return getUnixLibraryFileRegexPattern()
	}
}

func GetLibPaths() (paths []string) {
	switch os := runtime.GOOS; os {
	case "darwin":
		return getOsxLibPaths()
	case "windows":
		return paths
	default:
		return getLinuxLibPaths()
	}
}
