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
	"fmt"
	"github.com/sonatype-nexus-community/cheque/packages"
	"github.com/sonatype-nexus-community/cheque/audit"
	"os"
	"strings"
)

func doCheckExistenceAndParse() {
	dep := packages.Make{}
	dep.MakefilePath = path
	if dep.CheckExistenceOfManifest() {
		if (*isBom) {
			dep.ProjectList, _ = ParseBom(path)
		} else {
			dep.ProjectList, _ = ParseMakefile(path)
		}
		if (*bom) {
			for _,dep := range dep.ProjectList.Projects {
				if strings.HasPrefix(dep.Name, "pkg:") {
					_, _ = fmt.Printf("%s@%s\n", dep.Name, dep.Version)
				} else {
					_, _ = fmt.Printf("pkg:cpp/%s@%s\n", dep.Name, dep.Version)
				}
			}

			os.Exit(0)
		}

    AuditBom(dep.ProjectList)
	}
}
