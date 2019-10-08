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
	"strings"
  "github.com/golang/glog"
)

func doLink(args []string) {
  libPaths := []string {"/usr/lib/", "/usr/local/lib/", "/usr/lib/x86_64-linux-gnu/"}
  var libs []string
  var files []string

  for i := 0; i < len(args); i++ {
    arg := args[i]
    if (strings.HasPrefix(arg, "-l")) {
      if (len(arg) > 2) {
        glog.Info("lib: " + arg)
        libs = append(libs, arg[2:])
      } else {
        i++
        arg := args[i]
        glog.Info("lib: " + arg)
        libs = append(libs, arg)
      }
      continue;
    }

    // Additional library path
    if (strings.HasPrefix(arg, "-L")) {
      if (len(arg) > 2) {
        glog.Info("LibPath: " + arg)
        libPaths = append(libPaths, arg[2:])
      } else {
        i++
        arg := args[i]
        glog.Info("LibPath: " + arg)
        libPaths = append(libPaths, arg)
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
      files = append(files, arg)
      continue;
    }
    if (strings.HasSuffix(arg, ".so")) {
      glog.Info("Linux DLL: " + arg)
      files = append(files, arg)
      continue;
    }
    if (strings.Contains(arg, ".so.")) {
      glog.Info("Linux DLL: " + arg)
      files = append(files, arg)
      continue;
    }
    if (strings.HasSuffix(arg, ".a")) {
      glog.Info("Static lib: " + arg)
      files = append(files, arg)
      continue;
    }
    if (strings.Contains(arg, ".a.")) {
      glog.Info("Static lib: " + arg)
      files = append(files, arg)
      continue;
    }
  }

  processPaths(libPaths, libs, files)
}
