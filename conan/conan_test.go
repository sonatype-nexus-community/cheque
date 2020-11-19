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

package conan

import (
    "github.com/package-url/packageurl-go"
    "io/ioutil"
    "os"
    "testing"
)

func TestConanFileGenerates(t *testing.T) {
    options := setup(t)
    generator := New(*options)
    purls := make([]packageurl.PackageURL, 0)
    purls = append(purls, *packageurl.NewPackageURL("rpm", "", "name", "1.0.0", nil,""))
    generator.CheckOrCreateConanFile(purls)

    _, err := os.Stat(generator.filepath)
    if err != nil {
        t.Error(err)
    }

    expected := "[requires]\nlibname/1.0.0\n"

    contentMatch, contents := validateFileContents(generator.filepath, expected)
    if !contentMatch {
        t.Errorf("expected %s but got %s", expected, contents)
    }
    teardown(options.Directory)
}

func TestConanFileGeneratesWithoutDuplicates(t *testing.T) {
    options := setup(t)
    generator := New(*options)
    purls := make([]packageurl.PackageURL, 0)
    purls = append(purls, *packageurl.NewPackageURL("rpm", "", "name", "1.0.0", nil,""))
    purls = append(purls, *packageurl.NewPackageURL("rpm", "", "libname", "1.0.0", nil,""))
    generator.CheckOrCreateConanFile(purls)

    _, err := os.Stat(generator.filepath)
    if err != nil {
        t.Error(err)
    }

    expected := "[requires]\nlibname/1.0.0\n"

    contentMatch, contents := validateFileContents(generator.filepath, expected)
    if !contentMatch {
        t.Errorf("expected %s but got %s", expected, contents)
    }
    teardown(options.Directory)
}

func validateFileContents(file string, contents string) (bool, string) {
    readFileBytes, _ := ioutil.ReadFile(file)
    s := string(readFileBytes)
    return s == contents, s
}

func teardown(directory string) {
    _ = os.RemoveAll(directory)
}

func setup(t *testing.T) *Options {
    options := &Options{}
    tempDir, err := ioutil.TempDir("", "testconfig")
    if err != nil {
        t.Error(err)
    }
    options.Directory = tempDir
    return options
}
