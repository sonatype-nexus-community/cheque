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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/package-url/packageurl-go"

	"path/filepath"
)

/** Identify the coordinate using file path information
 */
type archiveCollector struct {
	path    string
	symlink string
}

func (c archiveCollector) SetExternalCommand(e ExternalCommand) {
	// NO-OP, no external command
}

func (c archiveCollector) IsValid() bool {
	return true
}

func (c archiveCollector) GetName() (string, error) {
	symlink, err := c.getSymlink()
	if err != nil {
		return symlink, err
	}

	return GetLibraryName(symlink)
}

func (c archiveCollector) GetVersion() (string, error) {
	symlink, err := c.getSymlink()
	if err != nil {
		return symlink, err
	}

	return GetLibraryVersion(symlink)
}

func (c *archiveCollector) getSymlink() (string, error) {
	if c.symlink == "" {
		symlink, err := filepath.EvalSymlinks(c.path)
		if err != nil {
			// Ignore the error in this case. If we cannot follow the symlink, try and
			// figure things out from the given path. This is particularly important
			// for testing, since afero does not support symlinks yet.
			return c.path, nil
		}
		c.symlink = symlink
	}
	return c.symlink, nil
}

func (c archiveCollector) GetPurlObject() (purl packageurl.PackageURL, err error) {
	name, err := getArchiveName(c.path)
	if err != nil {
		return purl, err
	}
	name = filepath.Base(name)
	version, err := getArchiveVersion(c.path)
	if err != nil {
		return purl, err
	}

	purl, err = packageurl.FromString(fmt.Sprintf("pkg:cpp/%s@%s", name, version))
	if err != nil {
		return purl, err
	}
	return
}

func (c archiveCollector) GetPath() (string, error) {
	if c.symlink != "" {
		return c.symlink, nil
	}
	return c.path, nil
}

/** Given a file path, extract a library name from the path.
 */
func getArchiveName(name string) (path string, err error) {
	path, _, err = getArchiveNameAndVersion(name)
	return path, err
}

/** Given a file path, extract a library name from the path.
 */
func getArchiveVersion(name string) (version string, err error) {
	_, version, err = getArchiveNameAndVersion(name)
	return version, err
}

func getArchiveNameAndVersion(path string) (name string, version string, err error) {

	// Extract a name
	fname := filepath.Base(path)
	fname = replaceArchiveExtension(fname)
	fname = removeSrcSuffix(fname)

	r, _ := regexp.Compile("^(.*?)[_\\.\\-]([0-9_\\.\\-]+[0-9_\\.\\-a-z]+)")
	matches := r.FindStringSubmatch(fname)
	if matches == nil {
		return "", "", errors.New("getUnixLibraryNameAndVersion: cannot get name/version from " + path + " (" + fname + ")")
	}
	repl := strings.NewReplacer("-", ".", "-", ".")
	return removeSrcSuffix(matches[1]), repl.Replace(matches[2]), nil
}

func removeSrcSuffix(name string) (s string) {
	r, _ := regexp.Compile("^(.*?)[_\\.\\-]src$")
	matches := r.FindStringSubmatch(name)
	if matches != nil {
		fmt.Printf("%s .. %s\n", name, matches[1])
		return matches[1]
	}
	fmt.Printf("%s -- %s\n", name, name)
	return name
}

func replaceArchiveExtension(name string) (s string) {
	result := name
	for ARCHIVE_REGEX.MatchString(result) {
		matches := ARCHIVE_REGEX.FindStringSubmatch(result)
		result = matches[1]
	}

	return result
}
