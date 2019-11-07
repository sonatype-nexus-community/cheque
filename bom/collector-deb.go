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
  "github.com/sonatype-nexus-community/cheque/logger"
  // "path/filepath"
  "errors"
  // "os"
  // "bufio"
  "strings"
  "regexp"
  "os/exec"
)

/** Identify the coordinate using file path information
 */
type deb_collector struct {
    path string
    pkgconfig string
    name string
    version string
    dist string
}

func (c deb_collector) GetName() (string, error) {
  if (c.dist == "") {
    c.findPackage()
  }
  if (c.name != "") {
    return c.name, nil
  }
  return "", errors.New("deb_collector: Cannot get name for " + c.path)
}

func (c deb_collector) GetVersion() (string, error) {
  if (c.dist == "") {
    c.findPackage()
  }
  if (c.version != "") {
    return c.version, nil
  }
  return "", errors.New("deb_collector: Cannot get version for " + c.path)
}


func (c deb_collector) GetPurl() (string, error) {
  name, err := c.GetName()
  if (err != nil) {
    return c.path, err
  }
  version, err := c.GetVersion()
  if (err != nil) {
    return name, err
  }
  return "pkg:deb/" + c.dist + "/" + name + "@" + version, nil
}

func (c deb_collector) GetPath() (string, error) {
  return c.path, nil
}

func (c *deb_collector) findPackage() {
  // Default distribution
  c.dist = "ubuntu"


  	dpkgCmd := exec.Command("dpkg", "-S", c.path)
  	out,err := dpkgCmd.Output()
  	if (err == nil) {
  		// fmt.Fprintf(os.Stderr, "GetUnixLibraryVersion 3.1 %s\n", out)
  		buf := string(out)
  		tokens := strings.Split(buf, ":")
  		libname := tokens[0]

  		dpkgCmd := exec.Command("dpkg", "-s", libname)
  		out,err := dpkgCmd.Output()
  		if (err == nil) {
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
