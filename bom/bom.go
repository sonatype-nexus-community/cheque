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
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/types"

	"fmt"
)

var UseRuntime = runtime.GOOS

var RPMExtCmd ExternalCommand
var DEBExtCmd ExternalCommand
var LDDCommand ExternalCommand
var OtoolCommand ExternalCommand

func init() {
	LDDCommand = LddExternalCommand{}
	RPMExtCmd = RpmExternalCommand{}
	DEBExtCmd = DebExternalCommand{}
	OtoolCommand = OtoolExternalCommand{}
}

// CreateBom does stuff
func CreateBom(libPaths []string, libs []string, files []string) (deps types.ProjectList, err error) {
	// Library names
	lookup := make(map[string]bool)
	deps.FileLookup = make(map[string]string)

	// Recursively get all transitive library paths
	for _, lib := range libs {
		lookup, err = recursiveGetLibraryPaths(lookup, libPaths, lib)
		if err != nil {
			fmt.Printf(err.Error() + "\n")
			continue
		}
	}

	for path := range lookup {
		project, err := getLibraryCoordinate(path)
		if err != nil {
			fmt.Printf(err.Error() + "\n")
			continue
		}

		deps.Projects = append(deps.Projects, project)
		deps.FileLookup[project.ToString()] = path
		deps = checkIfRpmOrDebAppendLibIfNot(project, path, deps)
	}

	for _, lib := range files {
		pattern := GetLibraryFileRegexPattern()
		rn, _ := regexp.Compile(pattern)
		fname := filepath.Base(lib)
		nameMatch := rn.FindStringSubmatch(fname)

		if nameMatch != nil {
			// If we have a nameMatch then this is a dynamic library
			project, err := getLibraryCoordinate(lib)
			if err != nil {
				fmt.Printf(err.Error() + "\n")
				continue
			}

			// We need both the "lib<name>" and "<name>" versions, since which is used
			// depends on the repo.
			deps.Projects = append(deps.Projects, project)
			deps.FileLookup[project.ToString()] = lib
			deps = checkIfRpmOrDebAppendLibIfNot(project, lib, deps)
		} else {
			// If we get here then this is a concrete non-dynamic library file. This might be a static library,
			// but may be any number of other supported files as well.
			purl, err := getFileCoordinate(lib)
			if err != nil {
				fmt.Printf(err.Error() + "\n")
				continue
			}

			// We need both the "lib<name>" and "<name>" versions, since which is used
			// depends on the repo.
			deps.Projects = append(deps.Projects, purl)
			deps.FileLookup[purl.ToString()] = lib
			deps = checkIfRpmOrDebAppendLibIfNot(purl, lib, deps)
		}
	}
	return deps, nil
}

func getTransitiveDependencies(path string) (results []byte, err error) {
	switch UseRuntime {
	case "windows":
		return
	case "darwin":
		return OtoolCommand.ExecCommand("-L", path)
	default:
		return LDDCommand.ExecCommand(path)
	}
}

func recursiveGetLibraryPaths(lookup map[string]bool, libPaths []string, lib string) (results map[string]bool, err error) {
	path, err := GetLibraryPath(libPaths, lib)
	if err != nil {
		logger.Debug(fmt.Sprintf("%v", err))
		return lookup, nil
	}
	if lookup[path] { // Does it already exist?
		return lookup, nil
	}

	lookup[path] = true

	out, err := getTransitiveDependencies(path)
	if err == nil {
		buf := string(out)
		lines := strings.Split(buf, "\n")
		for _, line := range lines {
			if line != "" {
				tokens := strings.Split(strings.TrimSpace(line), " ")
				token := strings.TrimSpace(tokens[0])
				lookup, err = recursiveGetLibraryPaths(lookup, append(libPaths, filepath.Dir(path)), token)
				if err != nil {
					return lookup, err
				}
			}
		}
	} else {
		return lookup, err
	}

	return lookup, nil
}

func checkIfRpmOrDebAppendLibIfNot(project packageurl.PackageURL, path string, deps types.ProjectList) types.ProjectList {
	if project.Type != "rpm" && project.Type != "deb" {
		if !strings.HasPrefix(project.Name, "lib") {
			project.Name = "lib" + project.Name
		} else {
			project.Name = project.Name[3:]
		}
		deps.Projects = append(deps.Projects, project)
		deps.FileLookup[project.ToString()] = path
	}
	return deps
}

func getLibraryCoordinate(path string) (purl packageurl.PackageURL, err error) {
	var collector Collector

	collector = pkgConfigCollector{path: path}
	purl, err = collector.GetPurlObject()
	if err == nil {
		return purl, nil
	}

	collector = rpmCollector{path: path, externalCommand: RPMExtCmd}
	if collector.IsValid() {
		purl, err = collector.GetPurlObject()
		if err == nil {
			return purl, nil
		}
	}

	collector = debCollector{path: path, externalCommand: DEBExtCmd}
	if collector.IsValid() {
		purl, err = collector.GetPurlObject()
		if err == nil {
			return purl, nil
		}
	}

	collector = pathCollector{path: path}
	purl, err = collector.GetPurlObject()
	if err == nil {
		return purl, nil
	}

	return
}

func getFileCoordinate(path string) (purl packageurl.PackageURL, err error) {
	var collector Collector

	collector = pkgConfigCollector{path: path}
	purl, err = collector.GetPurlObject()
	if err == nil {
		return purl, nil
	}

	collector = pathCollector{path: path}
	purl, err = collector.GetPurlObject()
	if err == nil {
		return purl, nil
	}

	collector = archiveCollector{path: path}
	purl, err = collector.GetPurlObject()
	if err == nil {
		return purl, nil
	}

	return
}
