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
	"fmt"
	"github.com/sonatype-nexus-community/cheque/config"
	"strings"

	"github.com/sonatype-nexus-community/cheque/audit"
	"github.com/sonatype-nexus-community/cheque/bom"
	"github.com/sonatype-nexus-community/cheque/logger"
)

var TYPESTOCHECK = map[string]string{
	".dylib": "OSX DLL",
	".so":    "Linux DLL",
	".so.":   "Linux DLL",
	".a":     "Static Lib",
	".a.":    "Static Lib",
}

type Linker struct {
	ossiConfig config.OSSIConfig
}

type Results struct {
	Count int
	LibPaths []string
	Libs []string
	Files []string
}

func New(config config.OSSIConfig) *Linker {
	return &Linker{
        ossiConfig: config,
	}
}

func (l Linker) DoLink(args []string) (results *Results) {
	libPaths := []string{}
	libs := make(map[string]bool)
	files := make(map[string]bool)

  for i := 0; i < len(args); i++ {
    arg := args[i]
    if strings.HasPrefix(arg, "-l") {
      if len(arg) > 2 {
        logger.Info("lib: " + arg)
        libs[arg[2:]] = true
      } else {
        i++
        arg := args[i]
        logger.Info("lib: " + arg)
				libs[arg] = true
      }
      continue
    }

    // Additional library path
    if strings.HasPrefix(arg, "-L") {
      if len(arg) > 2 {
        // logger.Info("LibPath: " + arg)
				libPaths = append(libPaths, arg[2:])
      } else {
        i++
        arg := args[i]
        // logger.Info("LibPath: " + arg)
				libPaths = append(libPaths, arg)
      }
      continue
    }

		if strings.HasPrefix(arg, "-o") {
			if len(arg) > 2 {
				logger.Info("output: " + arg)
			} else {
				i++
				arg := args[i]
				logger.Info("output: " + arg)
			}
			continue
		}

		// Ignore some arguments and their options
		if strings.HasPrefix(arg, "-install_name") {
			if (len(arg) > 14) {
				i++
			}
			continue
		}


    if strings.HasPrefix(arg, "-") {
      // Ignore any other arguments
      continue;
    }

    // -----------------------------------------
    // If we get here, it is a file of some sort
    // -----------------------------------------
		for k, v := range TYPESTOCHECK {
			files[arg] = checkSuffix(arg, k, v)
			if files[arg] {
				break
			}
		}
	}

	if len(libs) > 0 || len(files) > 0 {

		audit := audit.New(l.ossiConfig)
		libPaths := iterateAndAppendToLibPathsSlice(libPaths)
		libs := iterateAndAppendToSlice(libs)
		files := iterateAndAppendToSlice(files)
		count := audit.ProcessPaths(
			libPaths,
			libs,
			files)
		return &Results {
			LibPaths: libPaths,
			Libs: libs,
			Files: files,
			Count: count,
		}
	}

	return new(Results)
}

func iterateAndAppendToLibPathsSlice(libPaths []string) (libPathsSlice []string) {
	for _, v := range libPaths {
		libPathsSlice = append(libPathsSlice, v)
	}
	libPathsSlice = append(libPathsSlice, bom.GetLibPaths()...)
	return
}

func iterateAndAppendToSlice(iterator map[string]bool) (slice []string) {
	for k := range iterator {
		slice = append(slice, k)
	}
	return
}

func checkSuffix(arg string, extension string, loggerInfo string) bool {
	if strings.HasSuffix(extension, ".") {
		if strings.Contains(arg, extension) {
			logger.Info(fmt.Sprintf("%s: %s", loggerInfo, arg))
			return true
		}
		return false
	}
	if strings.HasSuffix(arg, extension) {
		logger.Info(fmt.Sprintf("%s: %s", loggerInfo, arg))
		return true
	}
	return false
}
