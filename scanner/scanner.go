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
package scanner

import (
	"os"
	"path/filepath"
	"strings"

	// "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	// "strings"

	"github.com/sonatype-nexus-community/cheque/config"

	"github.com/sonatype-nexus-community/cheque/audit"
	"github.com/sonatype-nexus-community/cheque/bom"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/logger"
)

var TYPESTOCHECK = map[string]string{
	".dylib": "OSX DLL",
	".so":    "Linux DLL",
	".so.":   "Linux DLL",
	".a":     "Static Lib",
	".a.":    "Static Lib",
	".dll":   "Windows DLL",
}

type Scanner struct {
	ossiConfig config.OSSIConfig
}

func New(config config.OSSIConfig) *Scanner {
	return &Scanner{
		ossiConfig: config,
	}
}

func (s Scanner) DoScan(path string, args []string) (results *linker.Results) {
	libPaths := []string{}
	mylibs := make(map[string]bool)
	files := make(map[string]bool)

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				if bom.IsArchive(path) {
					logger.Info("path " + path)

					files[path] = true
				}

				if strings.HasSuffix(path, ".pc") {
					logger.Info("pkgconfig " + path)

					files[path] = true
				}

				for k, v := range TYPESTOCHECK {
					if strings.HasSuffix(path, k) {
						logger.Info(v + " " + path)

						files[path] = true
					}
				}

			}
			return nil
		})
	if err != nil {
		logger.Error(err.Error())
	}

	if len(mylibs) > 0 || len(files) > 0 {
		audit := audit.New(s.ossiConfig)
		libPaths := iterateAndAppendToLibPathsSlice(libPaths)
		libs := iterateAndAppendToSlice(mylibs)
		files := iterateAndAppendToSlice(files)
		auditResults := audit.ProcessPaths(
			libPaths,
			libs,
			files)
		return &linker.Results{
			LibPaths:    libPaths,
			Libs:        libs,
			Files:       files,
			Count:       auditResults.Count,
			Coordinates: auditResults.Coordinates,
		}
	}

	return new(linker.Results)
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
