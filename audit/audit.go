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
package audit

import (
	"github.com/sonatype-nexus-community/cheque/bom"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/packages"
	"github.com/sonatype-nexus-community/cheque/types"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	typesNancy "github.com/sonatype-nexus-community/nancy/types"

	"strings"
)

func ProcessPaths(libPaths []string, libs []string, files []string) (count int) {
	myBom := packages.Make{}
	myBom.ProjectList, _ = bom.CreateBom(libPaths, libs, files)
	return AuditBom(myBom.ProjectList)
}

func AuditBom(deps types.ProjectList) (count int) {
	var canonicalPurls, _ = rootPurls(deps)
	var purls, _ = definedPurls(deps)
	var packageCount = countDistinctLibraries(append(canonicalPurls, purls...))

	if packageCount == 0 {
		return 0
	}

	var conanPurls, _ = conanPurls(canonicalPurls)
	var debPurls, _ = debPurls(canonicalPurls)
	var rpmPurls, _ = rpmPurls(canonicalPurls)
	purls = append(purls, conanPurls...)
	purls = append(purls, debPurls...)
	purls = append(purls, rpmPurls...)

	coordinates, err := ossindex.AuditPackages(purls)
	customerrors.Check(err, "Error auditing packages")

	var results []typesNancy.Coordinate
	lookup := make(map[string]typesNancy.Coordinate)

	for i := 0; i < len(coordinates); i++ {
		coordinate := coordinates[i]
		name := getCanonicalNameAndVersion(coordinate.Coordinates)

		// If the lookup is true, and there are vulnerabilities, then this name is done
		if val, ok := lookup[name]; ok {
			if len(val.Vulnerabilities) > 0 {
				continue
			}
		}

		// The lookup keeps getting replaced until one of them has a vulnerability
		lookup[name] = coordinate
	}

	// Now report the final/best value found for the libraries
	for _, v := range lookup {
		// Uncomment this to hide the source of the vulnerability
		// v.Coordinates = "pkg:cpp/" + k
		logger.Info(v.Coordinates)
		results = append(results, v)
	}
	count = audit.LogResults(false, packageCount, results)
	return count
}

func countDistinctLibraries(purls []string) (result int) {
	lookup := make(map[string]bool)

	for _, purl := range purls {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")
		name := tokens[len(tokens)-1]
		name = strings.TrimPrefix(name, "lib")
		lookup[name] = true
	}
	return len(lookup)
}

func getCanonicalNameAndVersion(path string) (result string) {
	tokens := strings.Split(path, ":")
	tokens = strings.Split(tokens[1], "/")

	name := tokens[len(tokens)-1]
	return strings.TrimPrefix(name, "lib")
}

// The root Purls are a generic purl which is not used for querying. It will
// subsequently be used to build *real* PURLs for OSS Index queries.
func rootPurls(deps types.ProjectList) (purls []string, err error) {
	for _, dep := range deps.Projects {
		if !strings.HasPrefix(dep.Name, "pkg:") {
			purls = append(purls, dep.String())
			if strings.HasPrefix(dep.Name, "lib") {
				libDep := dep
				libDep.Name = libDep.Name[3:]
				purls = append(purls, libDep.String())
			}
		}
	}
	return purls, nil
}

func definedPurls(deps types.ProjectList) (purls []string, err error) {
	for _, dep := range deps.Projects {
		if strings.HasPrefix(dep.Name, "pkg:") {
			purls = append(purls, dep.Name+"@"+dep.Version)
		}
	}
	return purls, nil
}

func conanPurls(purls []string) (results []string, err error) {
	for _, purl := range purls {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")

		// For now lets assume conan OR bincrafters. These are actually only two namespaces
		// of possibly many. We will want to get more clever here.
		results = append(results, "pkg:conan/conan/"+tokens[len(tokens)-1])
		results = append(results, "pkg:conan/bincrafters/"+tokens[len(tokens)-1])
	}
	return results, nil
}

func debPurls(misses []string) (results []string, err error) {
	for _, purl := range misses {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")

		results = append(results, "pkg:deb/debian/"+tokens[len(tokens)-1])
	}
	return results, nil
}

func rpmPurls(misses []string) (results []string, err error) {
	for _, purl := range misses {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")

		results = append(results, "pkg:rpm/fedora/"+tokens[len(tokens)-1])
	}
	return results, nil
}
