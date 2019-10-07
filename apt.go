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
package main

import (
	"fmt"
	// "log"
	"regexp"
	// "strings"

)

func doParseAptVersionIntoPurl(name string, version string) (newVersion string) {
	// exclude prefix delimited by :, and drop suffixes after .
	re, err := regexp.Compile(`^([0-9]+:)?(([0-9]+)\.([0-9]+)(\.([0-9]+))?)`)
	if err != nil {
		fmt.Println(err)
	}
	newSlice := re.FindStringSubmatch(version)
	if newSlice != nil {
		newVersion = newSlice[2]
	} else {
		// first approach failed, second attempt:
		// use prefix up to the first alphabetic character
		reNumericPrefix, err := regexp.Compile(`([^a-zA-Z]+)?`)
		if err != nil {
			fmt.Println(err)
		}
		numberPrefix := reNumericPrefix.FindStringSubmatch(version)
		if numberPrefix != nil {
			newVersion = numberPrefix[1]
		} else {
			// yikes, nothing we recognize. fallback to taking the string as is.
			fmt.Printf("package name: %s, using fallback value for version: %s\n", name, version)
			newVersion = version
		}
	}
	return
}
