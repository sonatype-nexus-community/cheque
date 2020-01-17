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
package bom

import (
  "github.com/sonatype-nexus-community/cheque/oslibs"
  "github.com/sonatype-nexus-community/cheque/logger"
  "path/filepath"
)

/** Identify the coordinate using file path information
 */
type path_collector struct {
    path string
    symlink string
}

func (c path_collector) GetName() (string, error) {
  symlink, err := c.getSymlink()
  if (err != nil) {
    return symlink, err
  }

  return oslibs.GetLibraryName(symlink)
}

func (c path_collector) GetVersion() (string, error) {
  symlink, err := c.getSymlink()
  if (err != nil) {
    return symlink, err
  }

  return oslibs.GetLibraryVersion(symlink)
}

func (c *path_collector) getSymlink() (string, error) {
  if (c.symlink == "") {
    symlink, err := filepath.EvalSymlinks(c.path)
    if (err != nil) {
      return "", err
    }
    c.symlink = symlink
  }
  return c.symlink, nil
}

func (c path_collector) GetPurl() (string, error) {
  name, err := c.GetName()
  if (err != nil) {
    return c.path, err
  }
  version, err := c.GetVersion()
  if (err != nil) {
    return name, err
  }
  return "pkg:cpp/" + name + "@" + version, nil
}

func (c path_collector) GetPath() (string, error) {
  if (c.symlink != "") {
    return c.symlink, nil
  }
  return c.path, nil
}
