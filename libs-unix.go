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
	"github.com/sonatype-nexus-community/nancy/types"
	"regexp"
	"os"
	"strings"
	// "fmt"
	"path/filepath"

	"os/exec"
	// "bytes"
)

func GetLinuxDistro() (name string) {

	return "Unknown";
}

func GetUnixLibraryId(name string) (project types.Projects, err error) {
  project = types.Projects{}
	file, err := FindUnixLibFile(name)
	// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 1 %s\n", file)

  if (err == nil) {
    if (file == "") {
      return project, nil
    }

		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 2 %s\n", file)
		// distro := strings.ToLower(GetLinuxDistro())
		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3 %s\n", distro)

		// try dpkg
		dpkgCmd := exec.Command("dpkg", "-S", file)
		out,err := dpkgCmd.Output()
		if (err == nil) {
			// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3.1 %s\n", out)
			buf := string(out)
			tokens := strings.Split(buf, ":")
			libname := tokens[0]

			dpkgCmd := exec.Command("dpkg", "-s", libname)
			out,err := dpkgCmd.Output()
			if (err == nil) {
				r, _ := regexp.Compile("Version: ([^\\n]+)")
				matches := r.FindStringSubmatch(string(out))
				if matches != nil {
					project.Name = "pkg:dpkg/ubuntu/" + libname
					project.Version = doParseAptVersionIntoPurl(libname, matches[1])
					// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3.2: %s %s\n", project.Name, project.Version)
					return project,nil
				}
			}
		}
		project.Version,err = GetUnixSymlinkVersion(file)
		return project,err;
  }

  // Try to fallback to pulling a version out of the filename
  if strings.HasSuffix(name, ".so") {
    // Extract a version
    r, err := regexp.Compile("\\.([0-9\\.]+)\\.so")
    if err != nil {
      return project, err
    }
    matches := r.FindStringSubmatch(name)
    if matches == nil {
      return project, nil
    }
		project.Version = matches[1]
    return project, nil
  }

  return project, nil
}

func FindUnixLibFile(name string) (match string, err error) {
	if strings.Contains(name, ".so.") || strings.HasSuffix(name, ".so") {
		// fmt.Fprintf(os.Stderr, "BUH 1 %s\n", name)
    if _, err := os.Stat(name); os.IsNotExist(err) {
      return "", err
    }
		return name,nil
	} else {
		// fmt.Fprintf(os.Stderr, "BUH 2 %s\n", name)

		return FindLibFile("lib", name, ".so")
	}
}

/** In some cases the library is a symbolic link to a file with an embedded version
 * number. Try and extract a version from there.
 */
func GetUnixSymlinkVersion(file string) (version string, err error) {
	path,err = filepath.EvalSymlinks(file)

	// fmt.Fprintf(os.Stderr, "GetUnixSymlinkVersion 2 %s\n", path)

	if err != nil {
		return "", err
	}

	// Extract a version
	r, err := regexp.Compile("\\.so\\.([0-9\\.]+)")
	if err != nil {
	// fmt.Fprintf(os.Stderr, "GetUnixSymlinkVersion 3 %s\n", path)
		return "", err
	}
	// fmt.Fprintf(os.Stderr, "GetUnixSymlinkVersion 4 %s\n", path)
	matches := r.FindStringSubmatch(path)
	if matches == nil {
		r, _ = regexp.Compile("([0-9\\.]+)\\.so")
		matches = r.FindStringSubmatch(path)
		if matches == nil {
			return "", nil
		}
	}

	return matches[1], nil
}

func GetUnixLibraryPathRegexPattern() (result string) {
	return "[a-zA-Z0-9_/\\.\\-]+\\.so\\.[a-zA-Z0-9_/\\.]+";
}


func GetUnixLibraryFileRegexPattern() (result string) {
	return "([a-zA-Z0-9_\\-]+)\\.so\\.[0-9\\.]+"
}
