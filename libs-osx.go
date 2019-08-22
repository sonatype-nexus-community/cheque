package main

import (
	"fmt"
	"os"
	"strings"
  "regexp"
  "path/filepath"

	// Required to run external commands
	"os/exec"
)


func GetOsxLibraryVersion(name string) (version string, err error) {
  file, err := FindOsxLibFile(name)

  if (err == nil) {
    if (file == "") {
      return "", nil
    }

    return GetOsxSymlinkVersion(file)

    // return GetOtoolVersion("lib" + name, file)
    return "", nil
  }

  // Try to fallback to pulling a version out of the filename
  if strings.HasSuffix(name, ".dylib") {
    // Extract a version
    r, err := regexp.Compile("\\.([0-9\\.]+)\\.dylib")
    if err != nil {
      return "", err
    }
    matches := r.FindStringSubmatch(name)
    if matches == nil {
      return "", nil
    }

    return matches[1], nil
  }

  return "", nil
}

/** Run otool and get version strings from it
 *
 * 	/usr/lib/libpam.2.dylib (compatibility version 3.0.0, current version 3.0.0)
 */
func GetOtoolVersion(name string, file string) (version string, err error) {
	outbytes, err := exec.Command("otool", "-L", file).Output()
	if err != nil {
		return "", err
	}
	out := string(outbytes)

	lines := strings.Split(out, "\n")

	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "SNUH %s\n", line)
	}

	return "", nil
}

func FindOsxLibFile(name string) (match string, err error) {
	if strings.HasSuffix(name, ".dylib") {
    if _, err := os.Stat(name); os.IsNotExist(err) {
      return "", err
    }
		return name,nil
	} else {
		return FindLibFile("lib", name, ".dylib")
	}
}

func GetOsxSymlinkVersion(file string) (version string, err error) {
	path,err = filepath.EvalSymlinks(file)

	if err != nil {
		return "", err
	}

	// Extract a version
	r, err := regexp.Compile("\\.([0-9\\.]+)\\.dylib")
	if err != nil {
		return "", err
	}
	matches := r.FindStringSubmatch(path)
	if matches == nil {
		return "", nil
	}

	return matches[1], nil
}
