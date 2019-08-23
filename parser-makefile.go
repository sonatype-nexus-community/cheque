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
)

func ParseMakefile(path string) (deps types.ProjectList, err error) {
	file, err := os.Open(path)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	s := string(b)

	{
		// look for -l libs
		r, _ := regexp.Compile("-l[a-zA-Z0-9_]+")
		matches := r.FindAllString(s, -1)

		for _,lib := range matches {
			lib = lib[2:]

			version, err := GetLibraryVersion(lib)
			customerrors.Check(err, "Error finding file/version")

			if (version != "") {
				// Add the simple name
				project := types.Projects{}
				project.Name = lib;
				project.Version = version
				deps.Projects = append(deps.Projects, project)

				// Also add with "lib" prepended, since that is how it may be represented in some ecosystems
				libProject := types.Projects{}
				libProject.Name = "lib" + lib;
				libProject.Version = project.Version
				deps.Projects = append(deps.Projects, libProject)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Cannot find '%s' library... skipping\n", lib)
			}
		}
	}

	{
		// look for libs in path
		r, _ := regexp.Compile("[a-zA-Z0-9_/\\.]+\\.dylib")
		matches := r.FindAllString(s, -1)

		for _,lib := range matches {
      rn, _ := regexp.Compile("([a-zA-Z0-9_]+)\\.[0-9\\.]+\\.dylib")
      nameMatch := rn.FindStringSubmatch(lib)

			version, err := GetLibraryVersion(lib)
			customerrors.Check(err, "Error finding file/version")

			if (version != "") {
				project := types.Projects{}
				project.Name = nameMatch[1];
				project.Version = version
				deps.Projects = append(deps.Projects, project)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Cannot find '%s' library... skipping\n", lib)
			}
		}
	}

	return deps, nil
}
