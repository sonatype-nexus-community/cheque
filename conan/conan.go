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
    "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
    "strings"
)

var (
    newline = "\n"
)

type conanPurlInfo struct {
    name string
    version string
}

type Options struct {
    Directory  string
    BinaryName string
}

type ConanGenerator struct {
    filepath string
    logger *logrus.Logger
}

func New(logger *logrus.Logger, options Options) *ConanGenerator {
    if options.Directory == "" {
        options.Directory = "."
    }

    return &ConanGenerator{
        filepath: filepath.Join(options.Directory, "conanfile." + options.BinaryName + ".cheque"),
        logger: logger,
    }
}

func (c ConanGenerator) CheckOrCreateConanFile(purls []packageurl.PackageURL) (e error) {
    var err error
    if _, err = os.Stat(c.filepath); os.IsNotExist(err) {
        duplessPurls := c.checkForDuplicates(purls)
        c.writeConanFile(duplessPurls)
        err = nil
    }
    return err
}

func (c ConanGenerator) checkForDuplicates(purls []packageurl.PackageURL) []conanPurlInfo {
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

func (c ConanGenerator) writeConanFile(purls []conanPurlInfo) {
    file, err := os.Create(c.filepath)
    if err != nil {
        c.logger.Error(err)
        return
    }
    defer file.Close()

    header := "[requires]" + newline
    n, err := file.WriteString(header)
    if err != nil {
        c.logger.Error(err)
        return
    }
    if n != len(header) {
        c.logger.Error("Unable to write data")
        return
    }
    for _, purl := range purls {
        purlInfo := purl.name + "/" + purl.version + newline
        n, err := file.WriteString(purlInfo)
        if err != nil {
            c.logger.Error(err)
            return
        }
        if n != len(purlInfo) {
            c.logger.Error("Unable to write data")
            return
        }
    }
}
