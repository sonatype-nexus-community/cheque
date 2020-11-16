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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/logger"
)

/** Identify the coordinate using file path information
 */
type pkgConfigCollector struct {
	path      string
	pkgconfig string
	name      string
	version   string
}

func (c pkgConfigCollector) SetExternalCommand(e ExternalCommand) {
	// NO OP, no external command
}

func (c pkgConfigCollector) IsValid() bool {
	return true
}

func (c pkgConfigCollector) GetName() (string, error) {
	if c.pkgconfig == "" {
		c.parsePkgConfig()
	}

	if c.pkgconfig == "" {
		return "", errors.New("pkgconfig_collector: No pkgconfig file for " + c.path)
	}

	if c.name == "" {
		return "", errors.New("pkgconfig_collector: No pkgconfig name found for " + c.pkgconfig)
	}

	return c.name, nil
}

func (c pkgConfigCollector) GetVersion() (string, error) {
	if c.pkgconfig == "" {
		c.parsePkgConfig()
	}

	if c.version == "" {
		return "", errors.New("pkgconfig_collector: No pkgconfig version found")
	}

	return c.version, nil
}

func (c pkgConfigCollector) GetPurlObject() (purl packageurl.PackageURL, err error) {
	name, err := c.GetName()
	if err != nil {
		return purl, err
	}
	version, err := c.GetVersion()
	if err != nil {
		return purl, err
	}
	purl, err = packageurl.FromString(fmt.Sprintf("pkg:cpp/%s@%s", name, version))
	return purl, err
}

func (c pkgConfigCollector) GetPath() (string, error) {
	return c.path, nil
}

func (c *pkgConfigCollector) parsePkgConfig() {
	dpath := filepath.Dir(c.path)
	base := filepath.Base(c.path)
	extension := filepath.Ext(base)
	base = base[0 : len(base)-len(extension)]
	path := dpath + "/pkgconfig/" + base + ".pc"

	if _, err := AppFs.Stat(path); os.IsNotExist(err) {
		path = dpath + "/" + base + ".pc"
		if _, err := AppFs.Stat(path); os.IsNotExist(err) {
			c.pkgconfig = "unknown"
			return
		}
	}

	c.pkgconfig = path

	file, err := AppFs.Open(path)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name:") {
			c.name = strings.TrimSpace(line[5:])
		} else if strings.HasPrefix(line, "Version:") {
			c.version = strings.TrimSpace(line[8:])
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(err.Error())
	}
}
