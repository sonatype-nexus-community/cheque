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