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
package main

import (
  // "fmt"
  "os"

	// Required to get OS
	"runtime"

	// Required to search file system
	"path/filepath"
)

var libpaths = []string {"/usr/lib/", "/usr/local/lib/"}

// Depending on the operating system, we need to find library versions in
// different ways.
//
//   OSX: otool -L <path>
//   Linux: Get file name (may be symbolically linked)
//   Windows: There is a way...
func GetLibraryVersion(name string) (version string, err error) {

	if runtime.GOOS == "windows" {
    return GetWindowsLibraryVersion(name)
  }

	if runtime.GOOS == "darwin" {
    return GetOsxLibraryVersion(name)
  }

	// Fall back to unix variant
  return GetUnixLibraryVersion(name)
}


func FindLibFile(prefix string, lib string, suffix string) (match string, err error) {

  for _, libpath := range libpaths {
    // _, _ = fmt.Fprintf(os.Stderr, "TRY 1  %s\n", libpath + prefix + lib + suffix)

   	matches, err := filepath.Glob(libpath + prefix + lib + suffix)

  	if err == nil {
    	if len(matches) != 0 {

        if _, err := os.Stat(matches[0]); os.IsNotExist(err) {
          // Do nothing
        } else {
    	     return matches[0], nil
        }
    	}
    }
  }
	return "", nil
}
