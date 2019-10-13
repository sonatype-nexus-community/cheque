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
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/cheque/parse"
	"flag"
	"fmt"
	"os"
)

var bom *bool
var isBom *bool
var isMakefile *bool
var isLog *bool
var noColorPtr *bool
var path string
var workingDir string

func main() {
	args := os.Args[1:]

	bom = flag.Bool("export-bom", false, "generate a Bill Of Materials only")
	isBom = flag.Bool("parse-bom", false, "The input file is a Bill Of Materials")
	isMakefile = flag.Bool("parse-makefile", false, "The input file is a Makefile")
	isLog = flag.Bool("parse-log", false, "The input file is a Make log")
	noColorPtr = flag.Bool("noColor", false, "indicate output should not be colorized")
	version := flag.Bool("version", false, "prints current auditcpp version")
	flag.StringVar(&workingDir, "working-directory", ".", "Resolve file paths relative to the specified directory")


	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: \nauditcpp [options] </path/to/Makefile>\n\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Parse flags from the command line output
	flag.Parse()

	if *version {
		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		os.Exit(0)
	}

	if *isBom || *isMakefile || *isLog {
		path = args[len(args)-1]
		// Currently only checks Makefile, can eventually check directory, cmake, etc...
		doCheckExistenceAndParse()
		os.Exit(0)
	}

	// Otherwise act like a compiler
	doLink(args);
}
