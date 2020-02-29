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
	"testing"

	"github.com/spf13/afero"
)

const UBUNTU = "ubuntu"

const OSX = "darwin"

// UNIX TESTS
func TestUnixGetLibraryPathBasic(t *testing.T) {
	SetupTestUnixFileSystem(UBUNTU)
	_, err := GetLibraryPath(nil, "libpng.so")
	if err == nil {
		t.Error(err)
	}
}

func TestUnixGetLibraryPath(t *testing.T) {
	SetupTestUnixFileSystem(UBUNTU)
	path, err := GetLibraryPath([]string{"/lib"}, "libpng.so")
	if err != nil {
		t.Error(err)
	}
	if path == "" {
		t.Errorf("Error, path is %s", path)
	}
}

func TestUnixGetLibraryPathWithoutLibOrExtension(t *testing.T) {
	SetupTestUnixFileSystem(UBUNTU)
	path, err := GetLibraryPath([]string{"/lib"}, "png")
	if err != nil {
		t.Error(err)
	}
	if path == "" {
		t.Errorf("Error, path is %s", path)
	}
}

func TestUnixGetLibraryName(t *testing.T) {
	Goose = UBUNTU
	name, err := GetLibraryName("libpng.2.0.5.so")
	if err != nil {
		t.Error(err)
	}
	if name != "libpng" {
		t.Errorf("Error, name is %s", name)
	}
}

func TestUnixGetLibraryNameWithVersionLast(t *testing.T) {
	Goose = UBUNTU
	name, err := GetLibraryName("libpng.so.2.0.5")
	if err != nil {
		t.Error(err)
	}
	if name != "libpng" {
		t.Errorf("Error, name is %s", name)
	}
}

func TestUnixGetLibraryVersion(t *testing.T) {
	Goose = UBUNTU
	version, err := GetLibraryVersion("libpng-2.0.5.so")
	if err != nil {
		t.Error(err)
	}
	if version != "2.0.5" {
		t.Errorf("Error, version is %s", version)
	}
}

func TestUnixGetLibraryVersionWithVersionLast(t *testing.T) {
	Goose = UBUNTU
	version, err := GetLibraryVersion("libpng.so.2.0.5")
	if err != nil {
		t.Error(err)
	}
	if version != "2.0.5" {
		t.Errorf("Error, version is %s", version)
	}
}

// DARWIN TESTS

func TestDarwinGetLibraryPathBasic(t *testing.T) {
	SetupTestOSXFileSystem(OSX)
	_, err := GetLibraryPath(nil, "libpng.dylib")
	if err == nil {
		t.Error(err)
	}
}

func TestDarwinGetLibraryPath(t *testing.T) {
	SetupTestOSXFileSystem(OSX)
	path, err := GetLibraryPath([]string{"/lib"}, "libpng.dylib")
	if err != nil {
		t.Error(err)
	}
	if path == "" {
		t.Errorf("Error, path is %s", path)
	}
}

func TestDarwinGetLibraryPathWithoutLibOrExtension(t *testing.T) {
	SetupTestOSXFileSystem(OSX)
	path, err := GetLibraryPath([]string{"/lib"}, "png")
	if err != nil {
		t.Error(err)
	}
	if path == "" {
		t.Errorf("Error, path is %s", path)
	}
}

func TestDarwinGetLibraryName(t *testing.T) {
	Goose = OSX
	name, err := GetLibraryName("libpng.2.0.5.dylib")
	if err != nil {
		t.Error(err)
	}
	if name != "libpng" {
		t.Errorf("Error, name is %s", name)
	}
}

func TestDarwinGetLibraryVersion(t *testing.T) {
	Goose = OSX
	version, err := GetLibraryVersion("libpng.2.0.5.dylib")
	if err != nil {
		t.Error(err)
	}
	if version != "2.0.5" {
		t.Errorf("Error, version is %s", version)
	}
}

// FILESYSTEM HELPERS

func SetupTestUnixFileSystem(operating string) {
	Goose = operating
	AppFs = afero.NewMemMapFs()

	AppFs.MkdirAll("/lib/", 0755)
	afero.WriteFile(AppFs, "/lib/libpng.so", []byte("file b"), 0644)
	afero.WriteFile(AppFs, "/lib/libtiff.a", []byte("file c"), 0644)
	afero.WriteFile(AppFs, "/lib/libsnuh.so.1.2.3", []byte("file d"), 0644)
	afero.WriteFile(AppFs, "/lib/libbuh.1.2.3.so", []byte("file e"), 0644)

	AppFs.MkdirAll("/usrdefined/path", 0755)
	afero.WriteFile(AppFs, "/usrdefined/path/libbob.so.1.2.3", []byte("file b"), 0644)
	afero.WriteFile(AppFs, "/usrdefined/path/libken.a", []byte("file c"), 0644)
}

func SetupTestOSXFileSystem(operating string) {
	Goose = operating
	AppFs = afero.NewMemMapFs()

	AppFs.MkdirAll("/lib/", 0755)
	afero.WriteFile(AppFs, "/lib/libpng.dylib", []byte("file b"), 0644)

	AppFs.MkdirAll("/usrdefined/path", 0755)
	afero.WriteFile(AppFs, "/usrdefined/path/libbob.dylib", []byte("file b"), 0644)
}
