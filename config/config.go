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
package config

import (
  "github.com/sonatype-nexus-community/nancy/buildversion"
	"os"
  "fmt"
  "path/filepath"
  "strings"
)

var bom = false
var isBom = false
var isMakefile = false
var isLog = false
var noColorPtr = false
var version = false
var path string
var workingDir string
var cmd string

func init() {
  cmd = filepath.Base(os.Args[0])

  args := os.Args[1:]

  if len(args) < 1 {
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
      case "parse-bom": isBom = true
      case "parse-makefile": isMakefile = true
      case "parse-log": isLog = true
      case "noColor": noColorPtr = true
      case "version": version = true
      case "working-directory": i++; workingDir = args[i]
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
  fmt.Fprintf(os.Stderr, "Usage: \nauditcpp [options] </path/to/Makefile>\n\nOptions:\n")
  fmt.Fprintf(os.Stderr, "Usage of cheque:\n")
  fmt.Fprintf(os.Stderr, "  -export-bom\n")
  fmt.Fprintf(os.Stderr, "    	generate a Bill Of Materials only\n")
  fmt.Fprintf(os.Stderr, "  -noColor\n")
  fmt.Fprintf(os.Stderr, "    	indicate output should not be colorized\n")
  fmt.Fprintf(os.Stderr, "  -parse-bom\n")
  fmt.Fprintf(os.Stderr, "    	The input file is a Bill Of Materials\n")
  fmt.Fprintf(os.Stderr, "  -parse-log\n")
  fmt.Fprintf(os.Stderr, "    	The input file is a Make log\n")
  fmt.Fprintf(os.Stderr, "  -parse-makefile\n")
  fmt.Fprintf(os.Stderr, "    	The input file is a Makefile\n")
  fmt.Fprintf(os.Stderr, "  -version\n")
  fmt.Fprintf(os.Stderr, "    	prints current auditcpp version\n")
  fmt.Fprintf(os.Stderr, "  -working-directory string\n")
  fmt.Fprintf(os.Stderr, "    	Resolve file paths relative to the specified directory (default '.')\n")

  os.Exit(2)
}

func GetCommand() (s string) {
  return cmd
}

func IsMakefile() (b bool) {
  return isMakefile
}

func IsLog() (b bool) {
  return isLog
}

func IsBom() (b bool) {
  return isBom
}

func GetBom() (b bool) {
  return bom
}
