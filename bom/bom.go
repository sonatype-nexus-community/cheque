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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/oslibs"

	"fmt"
)

/**
 * - Create Bom through a variety of mechanisms.
 * - Export Bom to file
 * - Import Bom from file
 */

func CreateBom(libPaths []string, libs []string, files []string) (deps []packageurl.PackageURL, err error) {
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

	for path, _ := range lookup {
		purl, err := getDllCoordinate(path)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		// Minor repair to names to make them consistent
		if !strings.HasPrefix(purl.Name, "lib") {
			purl.Name = "lib" + purl.Name
		}
		deps = append(deps, purl)
	}

	// Paths to libraries
	for _, lib := range files {
		rn, _ := regexp.Compile(oslibs.GetLibraryFileRegexPattern())
		nameMatch := rn.FindStringSubmatch(lib)

		if nameMatch != nil {
			// This is a dynamic library (DLL)
			purl, err := getDllCoordinate(lib)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			// We need both the "lib<name>" and "<name>" versions, since which is used
			// depends on the repo.
			deps = append(deps, purl)
			if !strings.HasPrefix(purl.Name, "lib") {
				purl.Name = "lib" + purl.Name
				deps = append(deps, purl)
			} else {
				purl.Name = purl.Name[3:]
				deps = append(deps, purl)
			}
		} else {
			purl, err := getArchiveCoordinate(lib)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			// We need both the "lib<name>" and "<name>" versions, since which is used
			// depends on the repo.
			deps = append(deps, purl)
			if !strings.HasPrefix(purl.Name, "lib") {
				purl.Name = "lib" + purl.Name
				deps = append(deps, purl)
			} else {
				purl.Name = purl.Name[3:]
				deps = append(deps, purl)
			}
		}
	}
	return deps, nil
}

func recursiveGetLibraryPaths(lookup map[string]bool, libPaths []string, lib string) (results map[string]bool, err error) {
	path, err := oslibs.GetLibraryPath(libPaths, lib)
	if err != nil {
		logger.Debug(fmt.Sprintf("%v", err))
		return lookup, nil
	}

	lookup[path] = true

	lddCmd := exec.Command("ldd", path)
	out, err := lddCmd.Output()
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

func getDllCoordinate(path string) (purl packageurl.PackageURL, err error) {
	var collector Collector
	// Check each collector in turn to see which gives us a good result.

	// pkgconfig_collector
	pc := pkgconfig_collector{path: path}
	purl, err = pc.GetPurl()
	if err == nil {
		return
	}

	// rpm_collector
	collector = rpm_collector{path: path}
	purl, err = collector.GetPurl()
	if err == nil {
		return
	}

	// deb_collector

	// path_collector
	collector = path_collector{path: path}
	purl, err = collector.GetPurl()
	if err == nil {
		return
	}

	// name_collector

	return
}

func getArchiveCoordinate(path string) (purl packageurl.PackageURL, err error) {
	// Check each collector in turn to see which gives us a good result.

	// pkgconfig_collector
	pc := pkgconfig_collector{path: path}
	purl, err = pc.GetPurl()
	if err == nil {
		return
	}

	// rpm_collector

	// deb_collector

	// path_collector
	collector := path_collector{path: path}
	purl, err = collector.GetPurl()
	if err == nil {
		return
	}

	return
}
