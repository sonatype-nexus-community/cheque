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
	"strings"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/types"

	"fmt"
)

var RPMExtCmd ExternalCommand
var DEBExtCmd ExternalCommand
var LDDCommand ExternalCommand

func init() {
	LDDCommand = LddExternalCommand{}
	RPMExtCmd = RpmExternalCommand{}
	DEBExtCmd = DebExternalCommand{}
}

// CreateBom does stuff
func CreateBom(libPaths []string, libs []string, files []string) (deps types.ProjectList, err error) {
	// Library names
	lookup := make(map[string]bool)

	// Recursively get all transitive library paths
	for _, lib := range libs {
		lookup, err = recursiveGetLibraryPaths(lookup, libPaths, lib)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
	}

	for path := range lookup {
		project, err := getLibraryCoordinate(path)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		deps.Projects = append(deps.Projects, project)
		deps = checkIfRpmOrDebAppendLibIfNot(project, deps)
	}

	for _, lib := range files {
		pattern := GetLibraryFileRegexPattern()
		rn, _ := regexp.Compile(pattern)
		fname := filepath.Base(lib)
		nameMatch := rn.FindStringSubmatch(fname)

		if nameMatch != nil {
			project, err := getLibraryCoordinate(lib)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			// We need both the "lib<name>" and "<name>" versions, since which is used
			// depends on the repo.
			deps.Projects = append(deps.Projects, project)
			deps = checkIfRpmOrDebAppendLibIfNot(project, deps)
		} else {
			purl, err := getArchiveCoordinate(lib)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			// We need both the "lib<name>" and "<name>" versions, since which is used
			// depends on the repo.
			deps.Projects = append(deps.Projects, purl)
			deps = checkIfRpmOrDebAppendLibIfNot(purl, deps)
		}
	}
	return deps, nil
}

func recursiveGetLibraryPaths(lookup map[string]bool, libPaths []string, lib string) (results map[string]bool, err error) {
	path, err := GetLibraryPath(libPaths, lib)
	if err != nil {
		logger.Debug(fmt.Sprintf("%v", err))
		return lookup, nil
	}

	lookup[path] = true

	out, err := LDDCommand.ExecCommand(path)
	if err == nil {
		buf := string(out)
		lines := strings.Split(buf, "\n")
		for _, line := range lines {
			if line != "" {
				tokens := strings.Split(line, " ")
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

func checkIfRpmOrDebAppendLibIfNot(project packageurl.PackageURL, deps types.ProjectList) types.ProjectList {
	if project.Type != "rpm" && project.Type != "deb" {
		if !strings.HasPrefix(project.Name, "lib") {
			project.Name = "lib" + project.Name
		} else {
			project.Name = project.Name[3:]
		}
		deps.Projects = append(deps.Projects, project)
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

func getArchiveCoordinate(path string) (purl packageurl.PackageURL, err error) {
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

	return
}
