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
	"strings"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/types"
)

const DPKGRESPONSE = `Package: debtest                                                                                                                                         
Status: install ok installed                                                                                                                           
Priority: required                                                                                                                                     
Section: libs                                                                                                                                          
Installed-Size: 11877                                                                                                                                  
Maintainer: Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>                                                                                  
Architecture: amd64                                                                                                                                    
Multi-Arch: same                                                                                                                                       
Source: glibc                                                                                                                                          
Version: 1.2.3                                                                                                                                 
Replaces: libc6-amd64                                                                                                                                  
Depends: libgcc1                                                                                                                                       
Suggests: glibc-doc, debconf | debconf-2.0, locales                                                                                                    
Breaks: hurd (<< 1:0.5.git20140203-1), libtirpc1 (<< 0.2.3), locales (<< 2.27), locales-all (<< 2.27), nscd (<< 2.27)                                  
Conflicts: openrc (<< 0.27-2~)`

type FakeLDDCommand struct {
}

func (f FakeLDDCommand) ExecCommand(args ...string) ([]byte, error) {
	byteResponse := []byte("linux-vdso.so.1 => (0x00007fff07fe0000)\nlibc.so.6 => /lib64/libc.so.6 (0x00007f4b3e3c8000)\n/lib64/ld-linux-x86-64.so.2 (0x00007f4b3e9ab000)")
	return byteResponse, nil
}

func (f FakeLDDCommand) IsValid() bool {
	return true
}

type FakeRPMCommand struct {
}

func (r FakeRPMCommand) ExecCommand(args ...string) ([]byte, error) {
	if strings.Contains(args[2], "rpmtest") {
		if args[1] == "--whatprovides" {
			return []byte("rpmtest-1.2.3-260.175.amzn1.x86_64"), nil
		}
		return []byte("Name        : rpmtest\nVersion     : 1.2.3"), nil
	}
	return []byte(fmt.Sprintf("file %s is not owned by any package", args[2])), fmt.Errorf("file %s is not owned by any package", args[2])
}

func (r FakeRPMCommand) IsValid() bool {
	return true
}

type FakeDEBCommand struct {
}

func (d FakeDEBCommand) ExecCommand(args ...string) ([]byte, error) {
	if strings.Contains(args[1], "debtest") {
		if args[0] == "-S" {
			return []byte("debtest:amd64: /lib64/ld-linux-x86-64.so.2"), nil
		}
		return []byte(DPKGRESPONSE), nil
	}
	return []byte(fmt.Sprintf("file %s is not owned by any package", args[2])), fmt.Errorf("file %s is not owned by any package", args[2])
}

func (d FakeDEBCommand) IsValid() bool {
	return true
}

func TestUnixCreateBom(t *testing.T) {
	SetupTestUnixFileSystem(UBUNTU)
	LDDCommand = FakeLDDCommand{}
	RPMExtCmd = FakeRPMCommand{}
	deps, err := CreateBom([]string{"/usrdefined/path"},
		[]string{"bob", "ken", "pkgtest", "rpmtest", "debtest"},
		[]string{"/lib/libpng.so", "/lib/libtiff.a", "/lib/libsnuh.so.1.2.3", "/lib/libbuh.4.5.6.so"})

	fmt.Print(deps)
	if err != nil {
		t.Error(err)
	}
	// Path based results
	assertResultContains(t, deps, "pkg:cpp/libbob@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/bob@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/libsnuh@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/snuh@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/libbuh@4.5.6")
	assertResultContains(t, deps, "pkg:cpp/buh@4.5.6")

	// pkgconfig based results
	assertResultContains(t, deps, "pkg:cpp/libken@2.3.4")
	assertResultContains(t, deps, "pkg:cpp/ken@2.3.4")
	assertResultContains(t, deps, "pkg:cpp/pkgtest@3.4.5")
	assertResultContains(t, deps, "pkg:cpp/pkgtest@3.4.5")

	// OS based results
	assertResultContains(t, deps, "pkg:rpm/rpmtest@1.2.3")
	assertResultContains(t, deps, "pkg:deb/ubuntu/debtest@1.2.3")

	// Should not get more than 12 results
	if len(deps.Projects) != 12 {
		t.Error(fmt.Sprintf("Expecting twelve (12) package in BOM, found %v", len(deps.Projects)))
	}
}

func TestOsxCreateBom(t *testing.T) {
	SetupTestUnixFileSystem(OSX)
	LDDCommand = FakeLDDCommand{}
	deps, err := CreateBom([]string{"/usrdefined/path"},
		[]string{"bob", "ken", "pkgtest"},
		[]string{"/lib/libpng.dylib", "/lib/libsnuh.1.2.3.dylib", "/lib/libbuh.4.5.6.dylib"})

	fmt.Print(deps)
	if err != nil {
		t.Error(err)
	}
	// Path based results
	assertResultContains(t, deps, "pkg:cpp/libbob@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/bob@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/libsnuh@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/snuh@1.2.3")
	assertResultContains(t, deps, "pkg:cpp/libbuh@4.5.6")
	assertResultContains(t, deps, "pkg:cpp/buh@4.5.6")

	// pkgconfig based results
	assertResultContains(t, deps, "pkg:cpp/libken@2.3.4")
	assertResultContains(t, deps, "pkg:cpp/ken@2.3.4")
	assertResultContains(t, deps, "pkg:cpp/pkgtest@3.4.5")
	assertResultContains(t, deps, "pkg:cpp/pkgtest@3.4.5")

	// Should not get more than 12 results
	if len(deps.Projects) != 12 {
		t.Error(fmt.Sprintf("Expecting twelve (10) package in BOM, found %v", len(deps.Projects)))
	}
}

func assertResultContains(t *testing.T, deps types.ProjectList, pstring string) {
	purl, err := packageurl.FromString(pstring)
	if err != nil {
		t.Error(err)
		return
	}

	for _, p := range deps.Projects {
		if p.Type == purl.Type &&
			p.Namespace == purl.Namespace &&
			p.Name == purl.Name &&
			p.Version == purl.Version {
			return
		}
	}

	t.Error("Missing expected PURL: " + purl.String())
}
