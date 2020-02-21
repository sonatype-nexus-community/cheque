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
	"fmt"
	"testing"
)

type FakeLDDCommand struct {
}

func (f FakeLDDCommand) ExecCommand(args ...string) ([]byte, error) {
	byteResponse := []byte("linux-vdso.so.1 => (0x00007fff07fe0000)\nlibc.so.6 => /lib64/libc.so.6 (0x00007f4b3e3c8000)\n/lib64/ld-linux-x86-64.so.2 (0x00007f4b3e9ab000)")
	return byteResponse, nil
}

func (f FakeLDDCommand) IsValid() bool {
	return true
}

func TestCreateBom(t *testing.T) {
	SetupTestUnixFileSystem(UBUNTU)
	LDDCommand = FakeLDDCommand{}
	deps, err := CreateBom([]string{"/usrdefined/path"},
		[]string{"bob", "ken"},
		[]string{"/lib/libpng.so", "/lib/libtiff.a"})

	fmt.Print(deps)
	if err == nil {
		t.Error("What Error")
	}
	if len(deps.Projects) > 0 {
		t.Error("What")
	}
}
