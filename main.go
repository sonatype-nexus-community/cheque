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
	"github.com/sonatype-nexus-community/cheque/config"
	"github.com/sonatype-nexus-community/cheque/parse"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/oslibs"
	"os"
	"os/exec"
	"github.com/golang/glog"
	"fmt"
)

func main() {
	args := os.Args[1:]

	if config.IsBom() || config.IsMakefile() || config.IsLog() {
		path := args[len(args)-1]
		// Currently only checks Makefile, can eventually check directory, cmake, etc...
		parse.DoCheckExistenceAndParse(path)
		os.Exit(0)
	}

	// Otherwise act like a compiler
	count := linker.DoLink(args);

	switch(config.GetCommand()) {
	case "cheque":
		break;
	default:
		cmd := oslibs.GetCommandPath(config.GetCommand())
		if cmd == "" {
			glog.Fatal("Cannot find official command: " + config.GetCommand())
		} else {
			// Run external command
			externalCmd := exec.Command(cmd, args...)
			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr
			err := externalCmd.Run()

			if err != nil {
				// FIXME: Return actual error code from command
				fmt.Fprintf(os.Stderr, "Error running %s: %v\n", config.GetCommand(), err)
			}
		}
		break;
	}

	if count > 0 {
		os.Exit(count)
  }
}
