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
	"flag"
	"fmt"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/cheque/packages"
	"os"
	"strings"
)

var bom *bool
var isBom *bool
var isMakefile *bool
var isLog *bool
var noColorPtr *bool
var path string

func main() {
	args := os.Args[1:]

	bom = flag.Bool("bom", false, "generate a Bill Of Materials only")
	isBom = flag.Bool("isBom", false, "The input file is a Bill Of Materials")
	isMakefile = flag.Bool("isMakefile", false, "The input file is a Makefile")
	isLog = flag.Bool("isLog", false, "The input file is a Make log")
	noColorPtr = flag.Bool("noColor", false, "indicate output should not be colorized")
	version := flag.Bool("version", false, "prints current auditcpp version")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: \nauditcpp [options] </path/to/Makefile>\n\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Parse flags from the command line output
	flag.Parse()

	if *version {
		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		os.Exit(0)
	}

	path = args[len(args)-1]

	// Currently only checks Makefile, can eventually check directory, cmake, etc...
	doCheckExistenceAndParse()
}

func doCheckExistenceAndParse() {
	// switch {
	// case strings.Contains(path, "Makefile"):
		dep := packages.Make{}
		dep.MakefilePath = path
		if dep.CheckExistenceOfManifest() {
			if (*isBom) {
				dep.ProjectList, _ = ParseBom(path)
			} else {
				dep.ProjectList, _ = ParseMakefile(path)
			}
			if (*bom) {
				for _,dep := range dep.ProjectList.Projects {
					if strings.HasPrefix(dep.Name, "pkg:") {
						_, _ = fmt.Printf("%s@%s\n", dep.Name, dep.Version)
					} else {
						_, _ = fmt.Printf("pkg:cpp/%s@%s\n", dep.Name, dep.Version)
					}
				}

				os.Exit(0)
			}

			var canonicalPurls,_ = RootPurls(dep.ProjectList)
			var purls,_ = DefinedPurls(dep.ProjectList)
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

			if count := audit.LogResults(*noColorPtr, packageCount, results); count > 0 {
				os.Exit(count)
			}
		}
	// default:
	// 	os.Exit(3)
	// }
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
