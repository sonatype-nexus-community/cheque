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

/** Check dpkg (if it exists).
 */
import (
	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/logger"

	"errors"
	"os/exec"
	"regexp"
	"strings"
)

/** Identify the coordinate using file path information
 */
type rpm_collector struct {
	path      string
	pkgconfig string
	name      string
	version   string
	dist      string
}

func (c rpm_collector) GetName() (string, error) {
	if c.dist == "" {
		c.findPackage()
	}
	if c.name != "" {
		return c.name, nil
	}
	return "", errors.New("rpm_collector: Cannot get name for " + c.path)
}

func (c rpm_collector) GetVersion() (string, error) {
	if c.dist == "" {
		c.findPackage()
	}
	if c.version != "" {
		return c.version, nil
	}
	return "", errors.New("rpm_collector: Cannot get version for " + c.path)
}

func (c rpm_collector) GetPurl() (purl packageurl.PackageURL, err error) {
	name, err := c.GetName()
	if err != nil {
		return
	}
	version, err := c.GetVersion()
	if err != nil {
		return
	}
	purl = packageurl.PackageURL{Type: "rpm", Name: name, Version: version}

	return
}

func (c rpm_collector) GetPath() (string, error) {
	return c.path, nil
}

func (c *rpm_collector) findPackage() {
	// Default distribution
	c.dist = "fedora"

	rpmCmd := exec.Command("rpm", "-q", "--whatprovides", c.path)
	out, err := rpmCmd.Output()
	if err == nil {
		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3.1 %s\n", out)
		libname := strings.TrimSpace(string(out))

		rpmCmd = exec.Command("rpm", "-q", "-i", libname)
		out, err := rpmCmd.Output()
		if err == nil {
			r, _ := regexp.Compile("Name *: ([^\\n]+)")
			matches := r.FindStringSubmatch(string(out))
			if matches != nil {
				c.name = strings.TrimSpace(matches[1])
			} else {
				logger.Error(err.Error())
				logger.Error(string(out))
				return
			}

			r, _ = regexp.Compile("Version *: ([^\\n]+)")
			matches = r.FindStringSubmatch(string(out))
			if matches != nil {
				c.version = strings.TrimSpace(matches[1])
			}
		} else {
			logger.Error(err.Error())
			logger.Error(string(out))
		}
	}
}

// func getDebianPackage(file string) (project types.Projects, err error) {
// 	project = types.Projects{}
//
// 	dpkgCmd := exec.Command("dpkg", "-S", file)
// 	out,err := dpkgCmd.Output()
// 	if (err == nil) {
// 		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3.1 %s\n", out)
// 		buf := string(out)
// 		tokens := strings.Split(buf, ":")
// 		libname := tokens[0]
//
// 		dpkgCmd := exec.Command("dpkg", "-s", libname)
// 		out,err := dpkgCmd.Output()
// 		if (err == nil) {
// 			r, _ := regexp.Compile("Version: ([^\\n]+)")
// 			matches := r.FindStringSubmatch(string(out))
// 			if matches != nil {
// 				project.Name = "pkg:dpkg/ubuntu/" + libname
// 				project.Version = doParseAptVersionIntoPurl(libname, matches[1])
// 				// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3.2: %s %s\n", project.Name, project.Version)
// 				return project,nil
// 			}
// 		}
// 	}
// 	return project, errors.New("Dpkg: Cannot find package")
// }
