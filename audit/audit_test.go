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
	"errors"
	"github.com/sonatype-nexus-community/cheque/config"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/package-url/packageurl-go"
	"github.com/shopspring/decimal"
	"github.com/sonatype-nexus-community/cheque/types"
  ossiTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
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

func setupVulnerability(id string, title string, description string, cvssScore string, vector string, cve string, reference string) ossiTypes.Vulnerability {
	dec, _ := decimal.NewFromString(cvssScore)
	return ossiTypes.Vulnerability{
		ID:          id,
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

	jsonCoordinates, _ := json.Marshal([]ossiTypes.Coordinate{
		{
			Coordinates: "pkg:rpm/fedora/name1@1.0.0",
			Reference:   "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name1@1.0.0",
			Vulnerabilities: []ossiTypes.Vulnerability{
				setupVulnerability("id", "title", "description", "5.8", "vector", "cve-123", "http://website")},
		},
		{
			Coordinates: "pkg:rpm/fedora/name2@1.0.0",
			Reference: "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name2@1.0.0",
			Vulnerabilities: []ossiTypes.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		httpmock.NewStringResponder(200, string(jsonCoordinates)))

	defer httpmock.DeactivateAndReset()

	setupProjectList()
	audit := New(config.OSSIConfig{})
	auditResults := audit.AuditBom(projectList.Projects)

	if auditResults.Count != 1 {
		t.Errorf("There is an error, expected 1, got %d", auditResults.Count)
	}
}

func TestPassesOssiCredentials(t *testing.T) {
	httpmock.Activate()

	jsonCoordinates, _ := json.Marshal([]ossiTypes.Coordinate{
		{
			Coordinates: "pkg:rpm/fedora/name1@1.0.0",
			Reference:   "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name1@1.0.0",
			Vulnerabilities: []ossiTypes.Vulnerability{
				setupVulnerability("id", "title", "description", "5.8", "vector", "cve-123", "http://website")},
		},
		{
			Coordinates: "pkg:rpm/fedora/name2@1.0.0",
			Reference: "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name2@1.0.0",
			Vulnerabilities: []ossiTypes.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		func(req *http.Request) (*http.Response, error) {
			auth := req.Header.Get("Authorization")
			//If we are missing auth, then its a problem, since we sent them.
			if auth != "Basic dXNlcjp0b2tlbjE=" {
				return httpmock.NewStringResponse(403, ""), errors.New("no authorization found")
			}
			return httpmock.NewStringResponse(200, string(jsonCoordinates)), nil
		})

	defer httpmock.DeactivateAndReset()

	setupProjectList()
	audit := New(config.OSSIConfig{
		Username: "user",
		Token: "token1",
	})

	//Make sure we have proper creds
	if !audit.HasProperOssiCredentials() {
		t.Error("Audit should have a proper config")
	}

	auditResults := audit.AuditBom(projectList.Projects)

	if auditResults.Count != 1 {
		t.Errorf("There is an error, expected 1, got %d", auditResults.Count)
	}
}

func TestWorksWithAbsenceOfOssiCredentials(t *testing.T) {
	httpmock.Activate()

	jsonCoordinates, _ := json.Marshal([]ossiTypes.Coordinate{
		{
			Coordinates: "pkg:rpm/fedora/name1@1.0.0",
			Reference:   "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name1@1.0.0",
			Vulnerabilities: []ossiTypes.Vulnerability{
				setupVulnerability("id", "title", "description", "5.8", "vector", "cve-123", "http://website")},
		},
		{
			Coordinates: "pkg:rpm/fedora/name2@1.0.0",
			Reference: "https://ossindex.sonatype.org/component/pkg:rpm/fedora/name2@1.0.0",
			Vulnerabilities: []ossiTypes.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		func(req *http.Request) (*http.Response, error) {
			auth := req.Header.Get("Authorization")
			//If we have auth, we should error.  we did not send any.
			if auth == "Basic dXNlcjp0b2tlbjE=" {
				return httpmock.NewStringResponse(403, ""), errors.New("we shouldn't have auth")
			}
			return httpmock.NewStringResponse(200, string(jsonCoordinates)), nil
		})

	defer httpmock.DeactivateAndReset()

	setupProjectList()
	audit := New(config.OSSIConfig{})

	//We should not have proper creds
	if audit.HasProperOssiCredentials() {
		t.Error("Audit should have a proper config")
	}

	auditResults := audit.AuditBom(projectList.Projects)

	if auditResults.Count != 1 {
		t.Errorf("There is an error, expected 1, got %d", auditResults.Count)
	}
}

func TestProcessPaths(t *testing.T) {
	audit := New(config.OSSIConfig{})

	auditResults := audit.ProcessPaths(
		[]string{"/usrdefined/path"},
		[]string{"bob", "ken"},
		[]string{"/lib/libpng.so", "/lib/libtiff.a"},
	)

	if auditResults.Count != 0 {
		t.Errorf("There is an error, expected 0, got %d", auditResults.Count)
	}
}
