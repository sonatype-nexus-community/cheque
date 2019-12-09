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
package oslibs

import (
  "os"

	// Required to search file system
	"path/filepath"
)

func findLibFile(libpaths []string, prefix string, lib string, suffix string) (match string, err error) {

  for _, libpath := range libpaths {
    globPattern := libpath + "/" + prefix + lib + suffix + "*";
    // fmt.Fprintf(os.Stderr, "findLibFile 1: %s\n", globPattern)
    // fmt.Fprintf(os.Stderr, "FindLibFile 1 %s\n", globPattern)
   	matches, err := filepath.Glob(globPattern)

  	if err == nil {
      // fmt.Fprintf(os.Stderr, "FindLibFile 2 %s\n", globPattern)
    	if len(matches) > 0 {
        for _,match := range matches {
          // fmt.Fprintf(os.Stderr, "FindLibFile 3 %s\n", match)

          if _, err := os.Stat(match); os.IsNotExist(err) {
            // Do nothing if file does not exist
          } else {
            //  fmt.Fprintf(os.Stderr, "FindLibFile 4 %s\n", match)
      	     return match, nil
          }
        }
    	}
    }
  }
	return "", nil
}
