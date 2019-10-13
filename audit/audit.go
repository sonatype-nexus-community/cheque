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
  "github.com/sonatype-nexus-community/cheque/oslibs"
  "github.com/sonatype-nexus-community/cheque/packages"
  "github.com/sonatype-nexus-community/nancy/types"
  "github.com/sonatype-nexus-community/nancy/customerrors"
  "github.com/sonatype-nexus-community/nancy/audit"
  "github.com/sonatype-nexus-community/nancy/ossindex"
  "github.com/golang/glog"
  "os"
	"regexp"
  "strings"
)

func ProcessPaths(libPaths []string, libs []string, files []string) {
  glog.Info("libPaths: " + strings.Join(libPaths, ", "))
  glog.Info("libs: " + strings.Join(libs, ", "))
  glog.Info("files: " + strings.Join(files, ", "))

  bom := packages.Make{}
  bom.ProjectList, _ = CreateBom(libPaths, libs, files)
  AuditBom(bom.ProjectList)
}

func CreateBom(_ []string, libs []string, files []string) (deps types.ProjectList, err error) {
  // Library names
  for _,lib := range libs {
    glog.Info("CreateBom 1: " + lib)
    project, err := oslibs.GetLibraryId(lib)
    customerrors.Check(err, "Error finding file/version")

    if (project.Version != "") {
      // Add the simple name
        if (project.Name == "") {
          project.Name = "lib" + lib
        }
      deps.Projects = append(deps.Projects, project)
    } else {
      glog.Error("Cannot find " + lib + " library... skipping")
    }
  }

  // Paths to libraries
  for _,lib := range files {
    glog.Info("CreateBom 2: " + lib)
    rn, _ := regexp.Compile(oslibs.GetLibraryFileRegexPattern())
    nameMatch := rn.FindStringSubmatch(lib)

    project, err := oslibs.GetLibraryId(lib)
    customerrors.Check(err, "Error finding file/version")

    if (project.Version != "") {
      if (project.Name == "") {
        project.Name = nameMatch[1];
      }
      deps.Projects = append(deps.Projects, project)
    } else {
      glog.Error("Cannot find " + lib + " library... skipping")
    }
  }
  return deps,nil
}

func AuditBom(deps types.ProjectList) {
  var canonicalPurls,_ = RootPurls(deps)
  var purls,_ = DefinedPurls(deps)
  var packageCount = CountDistinctLibraries(append(canonicalPurls, purls...))

  // For speed purposes, check all possible types simultaneously
  var conanPurls,_ = ConanPurls(canonicalPurls)
  var debPurls,_ = DebPurls(canonicalPurls)
  var rpmPurls,_ = RpmPurls(canonicalPurls)
  purls = append(purls, conanPurls...)
  purls = append(purls, debPurls...)
  purls = append(purls, rpmPurls...)

  coordinates, err := ossindex.AuditPackages(purls)
  customerrors.Check(err, "Error auditing packages")

  var results []types.Coordinate
  lookup := make(map[string]types.Coordinate)

  for i := 0; i < len(coordinates); i++ {
    coordinate := coordinates[i]
    name := GetCanonicalNameAndVersion(coordinate.Coordinates)

    // If the lookup is true, and there are vulnerabilities, then this name is done
    if val,ok := lookup[name]; ok {
      if (len(val.Vulnerabilities) > 0) {
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
    results = append(results, v)
  }

  if count := audit.LogResults(false, packageCount, results); count > 0 {
    os.Exit(count)
  }
}

func CountDistinctLibraries(purls []string) (result int) {
	lookup := make(map[string]bool)

	for _,purl := range purls {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")
		name := tokens[len(tokens) - 1];
		name = strings.TrimPrefix(name, "lib")
		lookup[name] = true
	}
	return len(lookup);
}

func GetCanonicalNameAndVersion(path string) (result string) {
	tokens := strings.Split(path, ":")
	tokens = strings.Split(tokens[1], "/")

	name := tokens[len(tokens) - 1];
	return strings.TrimPrefix(name, "lib")
}

// The root Purls are a generic purl which is not used for querying. It will
// subsequently be used to build *real* PURLs for OSS Index queries.
func RootPurls(deps types.ProjectList) (purls []string, err error) {
	for _,dep := range deps.Projects {
		if !strings.HasPrefix(dep.Name, "pkg:") {
			purls = append(purls, "pkg:cpp/" + dep.Name + "@" + dep.Version);
		}
	}
	return purls, nil
}

func DefinedPurls(deps types.ProjectList) (purls []string, err error) {
	for _,dep := range deps.Projects {
		if strings.HasPrefix(dep.Name, "pkg:") {
			purls = append(purls, dep.Name + "@" + dep.Version);
		}
	}
	return purls, nil
}

func ConanPurls(purls []string) (results []string, err error) {
	for _,purl := range purls {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")

		// For now lets assume bincrafters. This is actually only one namespace
		// of possibly many. We will want to get more clever here.
		results = append(results, "pkg:conan/bincrafters/" + tokens[len(tokens) - 1]);
	}
	return results, nil
}

func DebPurls(misses []string) (results []string, err error) {
	for _,purl := range misses {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")

		results = append(results, "pkg:deb/debian/" + tokens[len(tokens) - 1]);
	}
	return results, nil
}

func RpmPurls(misses []string) (results []string, err error) {
	for _,purl := range misses {
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")

		results = append(results, "pkg:rpm/fedora/" + tokens[len(tokens) - 1]);
	}
	return results, nil
}
