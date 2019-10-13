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
package linker

import (
	"github.com/sonatype-nexus-community/cheque/audit"
	"strings"
  "github.com/golang/glog"
)

func DoLink(args []string) (count int) {
	libPaths := make(map[string]bool)
	libPaths["/usr/lib/"] = true
	libPaths["/usr/local/lib/"] = true
	libPaths["/usr/lib/x86_64-linux-gnu/"] = true
	libs := make(map[string]bool)
	files := make(map[string]bool)

  for i := 0; i < len(args); i++ {
    arg := args[i]
    if (strings.HasPrefix(arg, "-l")) {
      if (len(arg) > 2) {
        glog.Info("lib: " + arg)
        libs[arg[2:]] = true
      } else {
        i++
        arg := args[i]
        glog.Info("lib: " + arg)
				libs[arg] = true
      }
      continue;
    }

    // Additional library path
    if (strings.HasPrefix(arg, "-L")) {
      if (len(arg) > 2) {
        glog.Info("LibPath: " + arg)
				libPaths[arg[2:]] = true
      } else {
        i++
        arg := args[i]
        glog.Info("LibPath: " + arg)
				libPaths[arg] = true
      }
      continue;
    }

    if (strings.HasPrefix(arg, "-")) {
      // Ignore any other arguments
      continue;
    }

    // -----------------------------------------
    // If we get here, it is a file of some sort
    // -----------------------------------------
    if (strings.HasSuffix(arg, ".dylib")) {
      glog.Info("OSX DLL: " + arg)
			files[arg] = true
      continue;
    }
    if (strings.HasSuffix(arg, ".so")) {
      glog.Info("Linux DLL: " + arg)
			files[arg] = true
      continue;
    }
    if (strings.Contains(arg, ".so.")) {
      glog.Info("Linux DLL: " + arg)
			files[arg] = true
      continue;
    }
    if (strings.HasSuffix(arg, ".a")) {
      glog.Info("Static lib: " + arg)
			files[arg] = true
      continue;
    }
    if (strings.Contains(arg, ".a.")) {
      glog.Info("Static lib: " + arg)
			files[arg] = true
      continue;
    }
  }

	if len(libs) > 0 || len(files) > 0 {
		libPathsSlice := []string{}
		for key, _ := range libPaths {
				libPathsSlice = append(libPathsSlice, key)
		}
		libsSlice := []string{}
		for key, _ := range libs {
				libsSlice = append(libsSlice, key)
		}
		filesSlice := []string{}
		for key, _ := range files {
				filesSlice = append(filesSlice, key)
		}

	  return audit.ProcessPaths(libPathsSlice, libsSlice, filesSlice)
	}

	return 0
}