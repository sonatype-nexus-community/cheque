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
	"github.com/sonatype-nexus-community/nancy/types"

	"runtime"
	"os"
)

var libpaths = []string {"/usr/lib/", "/usr/local/lib/", "/usr/lib/x86_64-linux-gnu/"}

// Depending on the operating system, we need to find library versions in
// different ways.
//
//   OSX: otool -L <path>
//   Linux: Get file name (may be symbolically linked)
//   Windows: There is a way...
func GetLibraryId(name string) (project types.Projects, err error) {
  project = types.Projects{}

  switch (runtime.GOOS) {
    case "windows":
        project.Version,err = getWindowsLibraryId(name)

    case "darwin":
        project.Version,err = getOsxLibraryId(name)

    default:
      return getUnixLibraryId(name)
  }
  return project, err;
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
