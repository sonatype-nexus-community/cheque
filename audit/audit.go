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
	"fmt"
	"github.com/sonatype-nexus-community/cheque/config"
	"text/tabwriter"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/bom"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/packages"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"

	"strings"
)

type Audit struct {
	OssiConfig         config.OSSIConfig
}

func New(ossiConfig config.OSSIConfig) *Audit {
    return &Audit{
        OssiConfig:         ossiConfig,
    }
}

func (a Audit) ProcessPaths(libPaths []string, libs []string, files []string) (count int) {
	myBom := packages.Make{}
	var projectList, _ = bom.CreateBom(libPaths, libs, files)
	myBom.Purls = projectList.Projects
	return a.AuditBom(myBom.Purls)
}

func (a Audit) HasProperOssiCredentials() bool {
	return len(a.OssiConfig.Username) > 0 && len(a.OssiConfig.Token) > 0
}

func (a Audit) AuditBom(deps []packageurl.PackageURL) (count int) {
	var canonicalPurls, _ = rootPurls(deps)
	var purls, _ = definedPurls(deps)
	var packageCount = countDistinctLibraries(append(canonicalPurls, purls...))

	if packageCount == 0 {
		return 0
	}

	// For speed purposes, check all possible types simultaneously
	var conanPurls, _ = conanPurls(canonicalPurls)
	var debPurls, _ = debPurls(canonicalPurls)
	var rpmPurls, _ = rpmPurls(canonicalPurls)
	purls = append(purls, conanPurls...)
	purls = append(purls, debPurls...)
	purls = append(purls, rpmPurls...)

	var ossi *ossindex.Server

    if a.HasProperOssiCredentials() {
        ossi = ossindex.New(
            logger.GetLogger(),
            types.Options{
                Version:     "development",
                Tool:        "cheque",
                DBCacheName: "cheque-cache",
                Username:    a.OssiConfig.Username,
                Token:       a.OssiConfig.Token,
            })
    } else {
        ossi = ossindex.New(
            logger.GetLogger(),
            types.Options{
                Version:     "development",
                Tool:        "cheque",
                DBCacheName: "cheque-cache",
            })
    }

	coordinates, err := ossi.AuditPackages(purls)
	if err != nil {
		logger.GetLogger().Error(err)
	}

	var results []types.Coordinate
	lookup := make(map[string]types.Coordinate)

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

	var sb strings.Builder

	w := tabwriter.NewWriter(&sb, 9, 3, 0, '\t', 0)
	_ = w.Flush()

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetTitle("Summary")
	t.AppendRow([]interface{}{"Audited Dependencies", len(results)})
	t.AppendSeparator()
	// t.AppendRow([]interface{}{"Vulnerable Dependencies", au.Bold(au.Red(strconv.Itoa(numVulnerable)))})
	sb.WriteString(t.Render())
	sb.WriteString("\n")

	for _, v := range results {
		if v.IsVulnerable() {
			count++
			LogVulnerablePackage(&sb, false, 0, 0, v)
		}
	}

	fmt.Print(sb.String())

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
func rootPurls(deps []packageurl.PackageURL) (purls []string, err error) {
	for _, dep := range deps {
		if !strings.HasPrefix(dep.Name, "pkg:") {
			purls = append(purls, "pkg:cpp/"+dep.Name+"@"+dep.Version)
			if strings.HasPrefix(dep.Name, "lib") {
				purls = append(purls, "pkg:cpp/"+dep.Name[3:]+"@"+dep.Version)
			}
		}
	}
	return purls, nil
}

func definedPurls(deps []packageurl.PackageURL) (purls []string, err error) {
	for _, dep := range deps {
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
