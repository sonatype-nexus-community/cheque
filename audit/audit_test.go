// Copyright 2020 Sonatype Inc.
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
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/package-url/packageurl-go"
	"github.com/shopspring/decimal"
	"github.com/sonatype-nexus-community/cheque/types"
	nancyTypes "github.com/sonatype-nexus-community/nancy/types"
)

var projectList = types.ProjectList{}

func setupProjectList() {
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:cpp/name@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:cpp/libname@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:deb/debian/name1@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:deb/debian/libname1@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:deb/ubuntu/name2@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:deb/ubuntu/libname2@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:rpm/redhat/name3@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:rpm/redhat/libname3@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:rpm/fedora/name4@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:rpm/fedora/libname4@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:conan/name5@1.0.0"))
	projectList.Projects = append(projectList.Projects, newPackageURL("pkg:conan/libname5@1.0.0"))
}

func newPackageURL(purl string) (newPurl packageurl.PackageURL) {
	newPurl, _ = packageurl.FromString(purl)
	return
}

func setupVulnerability(id string, title string, description string, cvssScore string, vector string, cve string, reference string) nancyTypes.Vulnerability {
	dec, _ := decimal.NewFromString(cvssScore)
	return nancyTypes.Vulnerability{
		Id:          id,
		Title:       title,
		Description: description,
		CvssScore:   dec,
		CvssVector:  vector,
		Cve:         cve,
		Reference:   reference,
	}
}

func TestAuditBom(t *testing.T) {
	httpmock.Activate()

	jsonCoordinates, _ := json.Marshal([]nancyTypes.Coordinate{
		{
			Coordinates: "pkg:rpm/fedora/name1@1.0.0",
			Reference:   "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name1@1.0.0",
			Vulnerabilities: []nancyTypes.Vulnerability{
				setupVulnerability("id", "title", "description", "5.8", "vector", "cve-123", "http://website")},
		},
		{
			Coordinates: "pkg:rpm/fedora/name2@1.0.0",
			Reference: "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name2@1.0.0",
			Vulnerabilities: []nancyTypes.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		httpmock.NewStringResponder(200, string(jsonCoordinates)))

	defer httpmock.DeactivateAndReset()

	setupProjectList()
	i := AuditBom(projectList)

	if i != 1 {
		t.Errorf("There is an error, expected 1, got %d", i)
	}
}

func TestProcessPaths(t *testing.T) {
	i := ProcessPaths(
		[]string{"/usrdefined/path"},
		[]string{"bob", "ken"},
		[]string{"/lib/libpng.so", "/lib/libtiff.a"},
	)

	if i != 0 {
		t.Errorf("There is an error, expected 0, got %d", i)
	}
}
