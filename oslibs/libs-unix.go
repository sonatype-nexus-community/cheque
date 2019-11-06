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
	"github.com/sonatype-nexus-community/cheque/logger"
	"regexp"
	"os"
	"strings"
	// "fmt"
	"path/filepath"
	"errors"
	"bufio"

	"os/exec"
	// "bytes"
)

/** Given a file path, extract a library name from the path.
 */
func getUnixLibraryName(name string) (path string, err error) {
	path, _, err = getUnixLibraryNameAndVersion(name)
	return path, err
}

/** Given a file path, extract a library name from the path.
 */
func getUnixLibraryVersion(name string) (version string, err error) {
	_, version, err = getUnixLibraryNameAndVersion(name)
	return version, err
}

func getLinuxDistro() (name string) {

	return "Unknown";
}

func getUnixArchiveId(name string) (project types.Projects, err error) {
  // fmt.Fprintf(os.Stderr, "getUnixArchiveId 1: %s\n", name)
  project = types.Projects{}
	file, err := findUnixLibFile(name)
	// fmt.Fprintf(os.Stderr, "getUnixArchiveId 1 %s\n", file)
	// fmt.Fprintf(os.Stderr, "getUnixArchiveId 2: %s\n", file)

  if (err == nil) {
		// fmt.Fprintf(os.Stderr, "getUnixArchiveId 3: %s\n", file)
    if (file == "") {
      return project, nil
    }

		// fmt.Fprintf(os.Stderr, "getUnixArchiveId 2 %s\n", file)
		// distro := strings.ToLower(GetLinuxDistro())
		// fmt.Fprintf(os.Stderr, "getUnixArchiveId 3 %s\n", distro)

		// try dpkg
		debProject,err := getDebianPackage(file)
		if err == nil {
			return debProject,err
		}
		// TODO: try rpm

		// fmt.Fprintf(os.Stderr, "getUnixArchiveId 4: %s\n", file)

		// Look for pkgconfig files related to this file
		pkgConfigProject,err := getPkgConfigVersion(file)
		if err == nil {
			return pkgConfigProject,err
		}

		// Fallback to looking at the file version, if we can find one
		project.Version,err = getUnixSymlinkVersion(file)
		// fmt.Fprintf(os.Stderr, "getUnixArchiveId 1: %s\n", project.Version)
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

func getUnixLibraryId(name string) (project types.Projects, err error) {
  // fmt.Fprintf(os.Stderr, "getUnixLibraryId 1: %s\n", name)
  project = types.Projects{}
	file, err := findUnixLibFile(name)
	// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 1 %s\n", file)
	// fmt.Fprintf(os.Stderr, "getUnixLibraryId 2: %s\n", file)

  if (err == nil) {
		// fmt.Fprintf(os.Stderr, "getUnixLibraryId 3: %s\n", file)
    if (file == "") {
      return project, nil
    }

		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 2 %s\n", file)
		// distro := strings.ToLower(GetLinuxDistro())
		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3 %s\n", distro)

		// try dpkg
		debProject,err := getDebianPackage(file)
		if err == nil {
			return debProject,err
		}
		// TODO: try rpm

		// fmt.Fprintf(os.Stderr, "getUnixLibraryId 4: %s\n", file)

		project.Version,err = getUnixSymlinkVersion(file)
		// fmt.Fprintf(os.Stderr, "getUnixLibraryId 1: %s\n", project.Version)
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

func getPkgConfigVersion(fpath string) (project types.Projects, err error) {
	project = types.Projects{}

	path := filepath.Dir(fpath)
	base := filepath.Base(fpath)
	extension := filepath.Ext(base)
	base = base[0:len(base)-len(extension)]
	path = path + "/pkgconfig/" + base + ".pc";

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return project, errors.New("PkgConfig: Cannot find package config")
	}
	// fmt.Fprintf(os.Stderr, "getPkgConfigVersion 1: %s\n", path)

	file, err := os.Open(path)
	if err != nil {
	    logger.Fatal(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name:") {
			project.Name = strings.TrimSpace(line[5:])
		} else if strings.HasPrefix(line, "Version:") {
			project.Version = strings.TrimSpace(line[8:])
		}
	}

	// fmt.Fprintf(os.Stderr, "getPkgConfigVersion 2: %s %s\n", project.Name, project.Version)

	if err := scanner.Err(); err != nil {
	    logger.Fatal(err.Error())
	}

	return project, nil
}

func getDebianPackage(file string) (project types.Projects, err error) {
	project = types.Projects{}

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
	return project, errors.New("Dpkg: Cannot find package")
}

func findUnixLibFile(name string) (match string, err error) {
	if strings.Contains(name, ".so.") || strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".a") {
		// fmt.Fprintf(os.Stderr, "BUH 1 %s\n", name)
    if _, err := os.Stat(name); os.IsNotExist(err) {
      return "", err
    }
		return name,nil
	} else {
		// fmt.Fprintf(os.Stderr, "findUnixLibFile 1 %s\n", name)

		return findLibFile("lib", name, ".so")
	}
}

func getUnixLibraryNameAndVersion(path string) (name string, version string, err error) {

	// Extract a name
	fname := filepath.Base(path)
	r, _ := regexp.Compile("^(.*)\\.so\\.([0-9\\.]+)")
	matches := r.FindStringSubmatch(path)
	if matches == nil {
		r, _ = regexp.Compile("^(.*)\\.([0-9\\.]+)\\.so")
		matches = r.FindStringSubmatch(path)
	}
	if matches == nil {
		return "", "", errors.New("getUnixLibraryNameAndVersion: cannot get name from " + path + " (" + fname + ")")
	}
	name = matches[1]

	// Extract a version
	r, _ = regexp.Compile("\\.so\\.([0-9\\.]+)")
	matches = r.FindStringSubmatch(path)
	if matches != nil {
		return name, matches[1], nil
	}

	r, _ = regexp.Compile("([0-9\\.]+)\\.so")
	matches = r.FindStringSubmatch(path)
	if matches != nil {
		return name, matches[1], nil
	}

	return name, "", errors.New("getUnixLibraryNameAndVersion: cannot get version from " + fname)
}

/** In some cases the library is a symbolic link to a file with an embedded version
 * number. Try and extract a version from there.
 */
func getUnixSymlinkVersion(file string) (version string, err error) {
	path,err := filepath.EvalSymlinks(file)

	// fmt.Fprintf(os.Stderr, "GetUnixSymlinkVersion 2 %s\n", path)

	if err != nil {
		return "", err
	}

	// Extract a version
	r, _ := regexp.Compile("\\.so\\.([0-9\\.]+)")
	matches := r.FindStringSubmatch(path)
	if matches != nil {
		return matches[1], nil
	}

	r, _ = regexp.Compile("([0-9\\.]+)\\.so")
	matches = r.FindStringSubmatch(path)
	if matches != nil {
		return matches[1], nil
	}

	return "", nil
}

func getUnixLibraryPathRegexPattern() (result string) {
	return "[a-zA-Z0-9_/\\.\\-]+\\.so\\.[a-zA-Z0-9_/\\.]+";
}


func getUnixLibraryFileRegexPattern() (result string) {
	return "([a-zA-Z0-9_\\-]+)\\.so\\.[0-9\\.]+"
}

func getUnixArchiveFileRegexPattern() (result string) {
	return "([a-zA-Z0-9_\\-]+)\\.a"
}

/** FIXME: Use gcc/ld to get search Paths
		> ld --verbose | grep SEARCH_DIR | tr -s ' ;' \\012
		SEARCH_DIR("/usr/x86_64-amazon-linux/lib64")
		SEARCH_DIR("/usr/lib64")
		SEARCH_DIR("/usr/local/lib64")
		SEARCH_DIR("/lib64")
		SEARCH_DIR("/usr/x86_64-amazon-linux/lib")
		SEARCH_DIR("/usr/local/lib")
		SEARCH_DIR("/lib")
		SEARCH_DIR("/usr/lib")

		> gcc -print-search-dirs
		install: /usr/lib/gcc/x86_64-amazon-linux/4.8.5/
		programs: =/usr/libexec/gcc/x86_64-amazon-linux/4.8.5/:/usr/libexec/gcc/x86_64-amazon-linux/4.8.5/:/usr/libexec/gcc/x86_64-amazon-linux/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/:/usr/lib/gcc/x86_64-amazon-linux/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../../x86_64-amazon-linux/bin/x86_64-amazon-linux/4.8.5/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../../x86_64-amazon-linux/bin/
		libraries: =/usr/lib/gcc/x86_64-amazon-linux/4.8.5/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../../x86_64-amazon-linux/lib/x86_64-amazon-linux/4.8.5/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../../x86_64-amazon-linux/lib/../lib64/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../x86_64-amazon-linux/4.8.5/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../../lib64/:/lib/x86_64-amazon-linux/4.8.5/:/lib/../lib64/:/usr/lib/x86_64-amazon-linux/4.8.5/:/usr/lib/../lib64/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../../x86_64-amazon-linux/lib/:/usr/lib/gcc/x86_64-amazon-linux/4.8.5/../../../:/lib/:/usr/lib/
 */
func getLinuxLibPaths() (map[string]bool) {
	libPaths := make(map[string]bool)
	// libPaths["/usr/lib/"] = true
	// libPaths["/usr/local/lib/"] = true
	// libPaths["/usr/lib/x86_64-linux-gnu/"] = true

	libPaths["/usr/x86_64-amazon-linux/lib64"] = true
	libPaths["/usr/lib64"] = true
	libPaths["/usr/local/lib64"] = true
	libPaths["/lib64"] = true
	libPaths["/usr/x86_64-amazon-linux/lib"] = true
	libPaths["/usr/local/lib"] = true
	libPaths["/lib"] = true
	libPaths["/usr/lib"] = true

	return libPaths
}
