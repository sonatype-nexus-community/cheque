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
package parse

import (
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/cheque/audit"
	"github.com/sonatype-nexus-community/cheque/oslibs"
  "os"
	"io/ioutil"
	"regexp"
	"strings"
)

func parseMakefile(path string) (deps types.ProjectList, err error) {
	libPaths := []string {"/usr/lib/", "/usr/local/lib/", "/usr/lib/x86_64-linux-gnu/"}
  var libs []string
  var files []string

	file, err := os.Open(path)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	s := string(b)

	{
		// look for -l libs
		r, _ := regexp.Compile("[[:space:]]-l[a-zA-Z0-9_]+")
		matches := r.FindAllString(s, -1)

		for _,lib := range matches {
			lib = strings.TrimSpace(lib)[2:]
			libs = append(libs, lib)
		}
	}

	{
		// look for libs in path
		r, _ := regexp.Compile(oslibs.GetLibraryPathRegexPattern())
		matches := r.FindAllString(s, -1)
		if (len(matches) > 0) {
			for _,lib := range matches {
				files = append(files, lib)
			}
		}
	}

	return audit.CreateBom(libPaths, libs, files)
}
