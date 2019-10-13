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
	"fmt"
	"os"
	"flag"

	// Required to get OS
	"runtime"
)


func GetWindowsLibraryId(name string) (version string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", runtime.GOOS)
	flag.PrintDefaults()
	os.Exit(2)
	return "", nil
}
