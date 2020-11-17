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
package context

import (
  "fmt"
  "github.com/sonatype-nexus-community/nancy/buildversion"
  "os"
  "path/filepath"
  "strings"
)

var bom = false
var noColorPtr = false
var version = false
var path string
var cmd string
var verbose = false

var exitWithError = false


func init() {
  cmd = filepath.Base(os.Args[0])

  args := os.Args[1:]

  if len(args) < 1 {
    fmt.Fprintf(os.Stderr, "Error: No input files\n\n")
    usage()
    os.Exit(1)
  }

  for i := 0; i < len(args); i++ {
    arg := args[i]
    if arg == "--" {
      break;
    }
    for strings.HasPrefix(arg, "-") {
      arg = arg[1:]
    }
    switch(arg) {
      case "help": usage()
      case "export-bom": bom = true
      case "Werror=cheque": exitWithError = true
      case "noColor": noColorPtr = true
      case "version": version = true
      case "v": verbose = true
    }
  }

  if version {
    fmt.Println(buildversion.BuildVersion)
    _, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
    _, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
    os.Exit(0)
  }
}

func usage() {
  fmt.Fprintf(os.Stderr, "Usage: cheque [options] <filename> ...\n\n")
  fmt.Fprintf(os.Stderr, "When you invoke cheque, it identifies static and dynamic library dependencies\n")
  fmt.Fprintf(os.Stderr, "and identifies known vulnerabilities using the OSS Index vulnerability database.\n\n")
  fmt.Fprintf(os.Stderr, "Cheque can be used as a wrapper around the compiler/linker by making symbolic\n")
  fmt.Fprintf(os.Stderr, "links to cheque with the compiler name, and ensuring they are in the front of\n")
  fmt.Fprintf(os.Stderr, "your PATH. Cheque will run, and also execute the compiler/linker appropriately.\n")
  fmt.Fprintf(os.Stderr, "This allows cheque to be embedded in most builds.\n\n")
  fmt.Fprintf(os.Stderr, "Option summary: (Many cheque options match those of the underlying compiler/linker)\n")
  fmt.Fprintf(os.Stderr, "  -L<dir>\n")
  fmt.Fprintf(os.Stderr, "    	Add the specified directory to the front of the library search path\n")
  fmt.Fprintf(os.Stderr, "  -l<library>\n")
  fmt.Fprintf(os.Stderr, "    	Specify the name of a DLL required for compiling/linking\n")
  fmt.Fprintf(os.Stderr, "  -Werror=cheque\n")
  fmt.Fprintf(os.Stderr, "    	Treat cheque warnings as errors\n")

  // fmt.Fprintf(os.Stderr, "  -export-bom\n")
  // fmt.Fprintf(os.Stderr, "    	generate a Bill Of Materials only\n")
  // fmt.Fprintf(os.Stderr, "  -noColor\n")
  // fmt.Fprintf(os.Stderr, "    	indicate output should not be colorized\n")
  fmt.Fprintf(os.Stderr, "  -version\n")
  fmt.Fprintf(os.Stderr, "    	prints current cheque version\n")

  os.Exit(2)
}

func GetCommand() (s string) {
  return cmd
}

func GetBom() (b bool) {
  return bom
}

func ExitWithError() (b bool) {
  return exitWithError
}

func GetVerbose() (b bool) {
  return verbose
}
