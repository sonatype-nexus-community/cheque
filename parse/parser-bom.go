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
  // "fmt"
  "os"
	// "io/ioutil"
	// "regexp"
	"strings"
	"bufio"
	"log"
)

func ParseBom(path string) (deps types.ProjectList, err error) {
	file, err := os.Open(path)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		purl := scanner.Text()
		tokens := strings.Split(purl, ":")
		tokens = strings.Split(tokens[1], "/")
		tokens = strings.Split(tokens[len(tokens) - 1], "@")

		project := types.Projects{}
		project.Name = tokens[0];
		project.Version = tokens[1]
		deps.Projects = append(deps.Projects, project)
	}

	if err := scanner.Err(); err != nil {
	    log.Fatal(err)
	}


	return deps, nil
}
