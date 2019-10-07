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
package main

import (
	"github.com/sonatype-nexus-community/nancy/types"
  "github.com/sonatype-nexus-community/nancy/customerrors"
  "fmt"
  "os"
	"io/ioutil"
	"regexp"
	"strings"
)

func ParseMakefile(path string) (deps types.ProjectList, err error) {
	file, err := os.Open(path)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	s := string(b)

	// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 1: %s\n", path)

	{
		// look for -l libs
		r, _ := regexp.Compile("[[:space:]]-l[a-zA-Z0-9_]+")
		matches := r.FindAllString(s, -1)

		for _,lib := range matches {
			lib = strings.TrimSpace(lib)[2:]
	// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 1.1: %s\n", lib)

			project, err := GetLibraryId(lib)
			customerrors.Check(err, "Error finding file/version")

			if (project.Version != "") {
				// Add the simple name
					if (project.Name == "") {
						project.Name = "lib" + lib
					}
				deps.Projects = append(deps.Projects, project)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Cannot find '%s' library... skipping\n", lib)
			}
		}
	}
	// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 2: %s\n", path)

	{
		// look for libs in path
		r, _ := regexp.Compile(GetLibraryPathRegexPattern())
		matches := r.FindAllString(s, -1)
		if (len(matches) > 0) {
			// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 2.1: %s\n", matches[0])

			for _,lib := range matches {
				// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 2.2: %s\n", lib)
	      rn, _ := regexp.Compile(GetLibraryFileRegexPattern())
	      nameMatch := rn.FindStringSubmatch(lib)
				// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 2.1: %s\n", nameMatch)

				project, err := GetLibraryId(lib)
				customerrors.Check(err, "Error finding file/version")

				if (project.Version != "") {
					if (project.Name == "") {
						project.Name = nameMatch[1];
					}
					deps.Projects = append(deps.Projects, project)
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "Cannot find '%s' library... skipping\n", lib)
				}
			}
		}
	}
	// _, _ = fmt.Fprintf(os.Stderr, "ParseMakefile 3: %s\n", path)

	return deps, nil
}
