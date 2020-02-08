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

	"github.com/sonatype-nexus-community/cheque/config"
	"github.com/sonatype-nexus-community/cheque/linker"
)

func main() {
	args := []string{}

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
		if config.ExitWithError() {
			fmt.Errorf("Error: Vulnerabilities found: %v", count)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Vulnerabilities found: %v\n", count)
		}
	}

	switch config.GetCommand() {
	case "cheque":
		break
	default:
		_, err := os.Stat(fmt.Sprint("/usr/bin/%s", config.GetCommand()))
		if err != nil {
			panic(fmt.Errorf("Cannot find official command: %s", config.GetCommand()))
		} else {
			externalCmd := exec.Command(config.GetCommand(), args...)
			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr

			if err := externalCmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					fmt.Errorf("There was an issue running the command %s, and the issue is %s", config.GetCommand(), os.Stderr)
					os.Exit(exitError.ExitCode())
				}
			}
		}
	}

	os.Exit(0)
}
