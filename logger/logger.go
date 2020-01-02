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
package logger

import (
  "github.com/sonatype-nexus-community/cheque/config"
  "fmt"
  "os"
)

func Info(msg string) {
  if config.GetVerbose() {
    fmt.Fprintf(os.Stdout, "Cheque INFO: %s\n", msg)
  }
}

func Debug(msg string) {
  if config.GetVerbose() {
    fmt.Fprintf(os.Stdout, "Cheque DEBUG: %s\n", msg)
  }
}

func Warning(msg string) {
  fmt.Fprintf(os.Stdout, "Cheque WARNING: %s\n", msg)
}

func Error(msg string) {
  fmt.Fprintf(os.Stdout, "Cheque ERROR: %s\n", msg)
}

func Fatal(msg string) {
  fmt.Fprintf(os.Stdout, "Cheque FATAL: %s\n", msg)
  os.Exit(99)
}
