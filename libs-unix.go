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
	"regexp"
	"os"
	"strings"
	"path/filepath"
)

func GetUnixLibraryVersion(name string) (version string, err error) {
	file, err := FindUnixLibFile(name)

  if (err == nil) {
    if (file == "") {
      return "", nil
    }

    return GetUnixSymlinkVersion(file)

    // return GetOtoolVersion("lib" + name, file)
    return "", nil
  }

  // Try to fallback to pulling a version out of the filename
  if strings.HasSuffix(name, ".so") {
    // Extract a version
    r, err := regexp.Compile("\\.([0-9\\.]+)\\.so")
    if err != nil {
      return "", err
    }
    matches := r.FindStringSubmatch(name)
    if matches == nil {
      return "", nil
    }

    return matches[1], nil
  }

  return "", nil
}

func FindUnixLibFile(name string) (match string, err error) {
	if strings.HasSuffix(name, ".so") {
    if _, err := os.Stat(name); os.IsNotExist(err) {
      return "", err
    }
		return name,nil
	} else {
		return FindLibFile("lib", name, ".so")
	}
}

/** In some cases the library is a symbolic link to a file with an embedded version
 * number. Try and extract a version from there.
 */
func GetUnixSymlinkVersion(file string) (version string, err error) {
	path,err = filepath.EvalSymlinks(file)

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
