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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sonatype-nexus-community/cheque/context"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/logger"
)

func main() {
	args := []string{}

	//Will check for config and create if necessary
	//options := config.Options{}
	//config := config.New(logger.GetLogger(), options)
	//config.CreateOrReadConfigFile()

	// Remove cheque custom arguments
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-Werror=cheque":
		default:
			args = append(args, arg)
		}
	}

	count := linker.DoLink(args)
	if count > 0 {
		if context.ExitWithError() {
			fmt.Fprintf(os.Stderr, "Error: Vulnerable dependencies found: %v\n", count)
			os.Exit(count)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Vulnerable dependencies found: %v\n", count)
		}
	}

	switch context.GetCommand() {
	case "cheque":
		break
	default:
		cmdPath := getWrappedCommand()

		if cmdPath == "" {
			logger.Fatal("Cannot find official command: " + cmdPath)
		} else {
			externalCmd := exec.Command(cmdPath, args...)

			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr

			if err := externalCmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					logger.Fatal(fmt.Sprintf("There was an issue running the command %s, and the issue is %v", context.GetCommand(), os.Stderr))
					os.Exit(exitError.ExitCode())
				}
			}
		}
		break
	}

	os.Exit(0)
}

// TODO: This is unix system specific. When we get around to supporting windows we will want to
// make this more platform agnostic.
func getWrappedCommand() (cmdPath string) {
	// Set this variable to the path of the cheque wrapper link
	var chequePath = ""

	var paths = strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		cmdPath, _ = filepath.Abs(fmt.Sprint(path, "/", context.GetCommand()))
		// If we don't know chequePath yet, then the first one found in path should be cheque
		if chequePath == "" {
			chequePath = cmdPath
			cmdPath = "" // Don't report this one as the wrapped binary
			continue
		}
		_, err := os.Stat(cmdPath)
		if err == nil && chequePath != cmdPath {
			break // Found the real binary
		}
		cmdPath = "" // Don't report this one as the wrapped binary
	}
	return
}
