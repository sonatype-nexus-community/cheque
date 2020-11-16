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
	"os"

	"github.com/spf13/afero"
)

func findLibFile(libpaths []string, prefix string, lib string, suffix string) (match string, err error) {
	for _, libpath := range libpaths {
		globPattern := libpath + "/" + prefix + lib + suffix
		if suffix != "" {
			globPattern = globPattern + "*"
		}
		matches, globErr := afero.Glob(AppFs, globPattern)
		if globErr == nil {
			if len(matches) > 0 {
				for _, v := range matches {
					_, err = AppFs.Stat(v)
					if os.IsNotExist((err)) {
						continue
					}
					err = nil
					match = v
					return
				}
			}
		}
	}
	return
}
