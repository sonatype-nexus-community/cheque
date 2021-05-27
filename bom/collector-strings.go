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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/package-url/packageurl-go"
)

var TYPESTOCHECK = map[string]string{
	".dylib": "OSX DLL",
	".so":    "Linux DLL",
	".so.":   "Linux DLL",
	".a":     "Static Lib",
	".a.":    "Static Lib",
	".dll":   "Windows DLL",
}

/** Identify library names by trying to find the full name with version in the "strings"
 * data of the library binary.
 */
type stringsCollector struct {
	path           string
	libraryStrings []string
	name           string
	version        string
}

func (c stringsCollector) SetExternalCommand(e ExternalCommand) {
	// NO OP, no external command
}

func (c stringsCollector) IsValid() bool {
	return true
}

func (c stringsCollector) GetName() (string, error) {
	names, err := c.getPossibleNames()
	if err == nil {
		return GetLibraryName(names)
	}
	return "", errors.New("No possible name matches found")
}

func (c stringsCollector) GetVersion() (string, error) {
	names, err := c.getPossibleNames()
	if err == nil {
		return GetLibraryVersion(names)
	}
	return "", errors.New("No possible name matches found")
}

/** Looking for a name and version that are high quality
 */
func (c stringsCollector) isValidNameAndVersion(name string) bool {
	myname, err := GetLibraryName(name)
	if err != nil || myname == "" {
		return false
	}
	myversion, err := GetLibraryVersion(name)
	if err != nil || myversion == "" {
		return false
	}

	// Make sure the version is of a reasonable length
	tokens := strings.Split(myversion, ".")
	if len(tokens) > 2 {
		return true
	}

	return false
}

func (c stringsCollector) getPossibleNames() (string, error) {
	// Only run on libraries
	found := false
	for k := range TYPESTOCHECK {
		if strings.HasSuffix(c.path, k) {
			found = true
		}
		// Special case for postfixed versions (libpng.so.1.2.3)
		if strings.HasPrefix(k, ".") && strings.Contains(c.path, k) {
			found = true
		}
	}

	if len(c.libraryStrings) == 0 {
		c.libraryStrings = getStrings(c.path)
	}

	if !found {
		return "", errors.New("Not a library")
	}

	// Remove path
	fname := filepath.Base(c.path)
	// Remove extension (and version number)
	tokens := strings.Split(fname, ".")
	fname = tokens[0]
	// Remove any trailing numbers
	fname = strings.TrimRight(fname, "1234567890")

	nameAndVersionPattern, _ := regexp.Compile("(" + fname + "[^ \\t]+)")
	nameSeparatorVersionPattern, _ := regexp.Compile("(" + fname + ") +\\w+ +((\\d+)\\.(\\d+)\\.(\\d+))")

	for i := 0; i < len(c.libraryStrings); i++ {
		if strings.Contains(c.libraryStrings[i], fname) {

			// try and find a good match for the full name with version
			matches := nameAndVersionPattern.FindStringSubmatch(c.libraryStrings[i])
			if matches != nil {
				if c.isValidNameAndVersion(matches[1]) {
					return matches[1], nil
				}
			}

			// Try and find a match with the name and version, but separated by other text
			matches = nameSeparatorVersionPattern.FindStringSubmatch(c.libraryStrings[i])
			if matches != nil {
				name := matches[1] + "." + matches[2] + ".so"
				if c.isValidNameAndVersion(name) {
					return name, nil
				}
			}

		}
	}
	return "", errors.New("No possible name matches found")
}

func (c stringsCollector) GetPurlObject() (purl packageurl.PackageURL, err error) {
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

func (c stringsCollector) GetPath() (string, error) {
	return c.path, nil
}

func (c stringsCollector) loadLibraryStrings() []string {
	return getStrings(c.path)
}

// The following code is derived from here: https://github.com/robpike/strings/blob/master/strings.go

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Strings is a more capable, UTF-8 aware version of the standard strings utility.
//
// Flags(=default) are:
//
//	-ascii(=false)    restrict strings to ASCII
//	-min(=6)          minimum length of UTF-8 strings printed, in runes
//	-max(=256)        maximum length of UTF-8 strings printed, in runes
//	-offset(=false)   show file name and offset of start of each string
//

var (
	min    = flag.Int("min", 6, "minimum length of UTF-8 strings printed, in runes")
	max    = flag.Int("max", 256, "maximum length of UTF-8 strings printed, in runes")
	ascii  = flag.Bool("ascii", false, "restrict strings to ASCII")
	offset = flag.Bool("offset", false, "show file name and offset of start of each string")
)

func getStrings(path string) (results []string) {
	log.SetFlags(0)
	log.SetPrefix("strings: ")

	fd, err := os.Open(path)
	if err != nil {
		log.Print(err)
	} else {
		results = do(path, fd)
		fd.Close()
	}
	return results
}

func do(name string, file *os.File) (results []string) {
	in := bufio.NewReader(file)
	str := make([]rune, 0, *max)
	filePos := int64(0)
	print := func() {
		if len(str) >= *min {
			s := string(str)
			results = append(results, s)
		}
		str = str[0:0]
	}
	for {
		var (
			r   rune
			wid int
			err error
		)
		// One string per loop.
		for ; ; filePos += int64(wid) {
			r, wid, err = in.ReadRune()
			if err != nil {
				if err != io.EOF {
					log.Print(err)
				}
				return
			}
			if !strconv.IsPrint(r) || *ascii && r >= 0xFF {
				print()
				continue
			}
			// It's printable. Keep it.
			if len(str) >= *max {
				print()
			}
			str = append(str, r)
		}
	}
}
