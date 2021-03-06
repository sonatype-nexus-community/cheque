// Copyright 2020 Sonatype Inc.
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
package bom

import (
	"os/exec"
)

type DebExternalCommand struct {
}

func (d DebExternalCommand) IsValid() bool {
	_, err := exec.LookPath("dpkg")
	if err != nil {
		return false
	}
	return true
}

func (d DebExternalCommand) ExecCommand(args ...string) ([]byte, error) {
	return exec.Command("dpkg", args...).CombinedOutput()
}
