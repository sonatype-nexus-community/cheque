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
    "strings"
)

var (
    newline = "\n"
)

type conanPurlInfo struct {
    name string
    version string
}

func CheckOrCreateConanFile(binary string, purls []packageurl.PackageURL) {
    _, err := os.Stat("conanfile.txt")
    if err != nil {
        duplessPurls := checkForDuplicates(purls)
        writeConanFile(binary, duplessPurls)
    }
}

func checkForDuplicates(purls []packageurl.PackageURL) []conanPurlInfo {
    duplessMap := make(map[conanPurlInfo]bool)

    //Add our lib prefix libraries
    for _, purl := range purls[:] {
        info := conanPurlInfo{
            name:    purl.Name,
            version: purl.Version,
        }
        if strings.HasPrefix(purl.Name, "lib") {
            duplessMap[info] = true
        }
    }

    //Check to see if our libless prefix's exist already in lib form.
    for _, purl := range purls[:] {
        info := conanPurlInfo{
            name:    "lib" + purl.Name,
            version: purl.Version,
        }
        if !strings.HasPrefix(purl.Name, "lib") && !duplessMap[info] {
            duplessMap[info] = true
        }
    }

    keys := make([]conanPurlInfo, 0, len(duplessMap))
    for k := range duplessMap {
        keys = append(keys, k)
    }

    return keys
}

func writeConanFile(binary string, purls []conanPurlInfo) {
    var data strings.Builder
    data.WriteString("[requires]")
    data.WriteString(newline)
    for _, purl := range purls[:] {
        data.WriteString(purl.name)
        data.WriteString("/")
        data.WriteString(purl.version)
        data.WriteString(newline)
    }
    ioutil.WriteFile("conanfile." + binary + ".cheque", []byte(data.String()), 0655)
}