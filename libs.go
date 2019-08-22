package main

import (
  "os"

	// Required to get OS
	"runtime"

	// Required to search file system
	"path/filepath"
)


// Depending on the operating system, we need to find library versions in
// different ways.
//
//   OSX: otool -L <path>
//   Linux: Get file name (may be symbolically linked)
//   Windows: There is a way...
func GetLibraryVersion(name string) (version string, err error) {

	if runtime.GOOS == "windows" {
    return GetWindowsLibraryVersion(name)
  }

	if runtime.GOOS == "darwin" {
    return GetOsxLibraryVersion(name)
  }

	// Fall back to unix variant
  return GetUnixLibraryVersion(name)
}


func FindLibFile(prefix string, lib string, suffix string) (match string, err error) {

	matches, err := filepath.Glob("/usr/lib/" + prefix + lib + suffix)

	if err != nil {
	  return "", err
	}

	if len(matches) != 0 {

    if _, err := os.Stat(matches[0]); os.IsNotExist(err) {
      return "", nil
    }

	  return matches[0], nil
	}
	return "", nil
}
