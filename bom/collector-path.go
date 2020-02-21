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
	"fmt"

	"github.com/package-url/packageurl-go"

	"path/filepath"
)

/** Identify the coordinate using file path information
 */
type pathCollector struct {
	path    string
	symlink string
}

func (c pathCollector) SetExternalCommand(e ExternalCommand) {
	// NO-OP, no external command
}

func (c pathCollector) IsValid() bool {
	return true
}

func (c pathCollector) GetName() (string, error) {
	symlink, err := c.getSymlink()
	if err != nil {
		return symlink, err
	}

	return GetLibraryName(symlink)
}

func (c pathCollector) GetVersion() (string, error) {
	symlink, err := c.getSymlink()
	if err != nil {
		return symlink, err
	}

	return GetLibraryVersion(symlink)
}

func (c *pathCollector) getSymlink() (string, error) {
	if c.symlink == "" {
		symlink, err := filepath.EvalSymlinks(c.path)
		if err != nil {
			return "", err
		}
		c.symlink = symlink
	}
	return c.symlink, nil
}

func (c pathCollector) GetPurl() (string, error) {
	name, err := c.GetName()
	if err != nil {
		return c.path, err
	}
	version, err := c.GetVersion()
	if err != nil {
		return name, err
	}
	return "pkg:cpp/" + name + "@" + version, nil
}

func (c pathCollector) GetPurlObject() (purl packageurl.PackageURL, err error) {
	name, err := c.GetName()
	if err != nil {
		return purl, err
	}
	version, err := c.GetVersion()
	if err != nil {
		return purl, err
	}
	// If it says lib remove it, if it doesn't, add it so we end up with libname and name purls
	purl, err = packageurl.FromString(fmt.Sprintf("pkg:cpp/%s@%s", name, version))
	if err != nil {
		return purl, err
	}
	return
}

func (c pathCollector) GetPath() (string, error) {
	if c.symlink != "" {
		return c.symlink, nil
	}
	return c.path, nil
}
