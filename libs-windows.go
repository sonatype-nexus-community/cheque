package main

import (
	"fmt"
	"os"
	"flag"
	
	// Required to get OS
	"runtime"
)


func GetWindowsLibraryVersion(name string) (version string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Unsupported OS: %s\n", runtime.GOOS)
	flag.PrintDefaults()
	os.Exit(2)
	return "", nil
}
