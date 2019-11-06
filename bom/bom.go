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
  "github.com/sonatype-nexus-community/nancy/types"
	"regexp"
  "strings"
  // "errors"
  // "fmt"
  // "os"
)

/**
 * - Create Bom through a variety of mechanisms.
 * - Export Bom to file
 * - Import Bom from file
 */

func CreateBom(_ []string, libs []string, files []string) (deps types.ProjectList, err error) {
 // Library names
 for _,lib := range libs {
   path, err := oslibs.GetLibraryPath(lib)
   if (err != nil) {
     logger.Error("Error finding path to library '" + lib + "'")
     continue
   }

   project, err := getDllCoordinate(path)
   if (err != nil) {
     logger.Error(err.Error())
     continue
   }

   // Minor repair to names to make them consistent
   if (!strings.HasPrefix(project.Name, "lib")) {
     project.Name = "lib" + project.Name
   }
   deps.Projects = append(deps.Projects, project)



   // // logger.Info("CreateBom 1: " + lib)
   // project, err := oslibs.GetLibraryId(lib)
   // customerrors.Check(err, "Error finding file/version")
   //
   // if (project.Version != "") {
   //   // Add the simple name
   //     if (project.Name == "") {
   //       project.Name = "lib" + lib
   //     }
   //   deps.Projects = append(deps.Projects, project)
   // } else {
   //   logger.Warning("Cannot find " + lib + " library... skipping")
   // }
 }

 // Paths to libraries
 for _,lib := range files {
   // logger.Info("CreateBom 2: " + lib)

   rn, _ := regexp.Compile(oslibs.GetLibraryFileRegexPattern())
   nameMatch := rn.FindStringSubmatch(lib)

   if nameMatch != nil {
     // This is a dynamic library (DLL)
     project, err := getDllCoordinate(lib)
     if (err != nil) {
       logger.Error(err.Error())
       continue
     }

     // Minor repair to names to make them consistent
     if (!strings.HasPrefix(project.Name, "lib")) {
       project.Name = "lib" + project.Name
     }
     deps.Projects = append(deps.Projects, project)

     // // logger.Info("CreateBom 2: " + fmt.Sprintf("%v", nameMatch))
     //
     // project, err := oslibs.GetLibraryId(lib)
     // customerrors.Check(err, "Error finding file/version")
     //
     // if (project.Version != "") {
     //   if (project.Name == "") {
     //     project.Name = nameMatch[1];
     //   }
     //   deps.Projects = append(deps.Projects, project)
     // } else {
     //   logger.Warning("Cannot find " + lib + " library... skipping")
     // }
   } else {
     project, err := getArchiveCoordinate(lib)
     if (err != nil) {
       logger.Error(err.Error())
       continue
     }

     // Minor repair to names to make them consistent
     if (!strings.HasPrefix(project.Name, "lib")) {
       project.Name = "lib" + project.Name
     }
     deps.Projects = append(deps.Projects, project)

     // // This is a static library (archive)
     // rn, _ := regexp.Compile(oslibs.GetArchiveFileRegexPattern())
     // nameMatch := rn.FindStringSubmatch(lib)
     //
     // // logger.Info("CreateBom 3: " + fmt.Sprintf("%v", nameMatch))
     //
     // project, err := oslibs.GetArchiveId(lib)
     // customerrors.Check(err, "Error finding file/version")
     //
     // if (project.Version != "") {
     //   if (project.Name == "") {
     //     project.Name = nameMatch[1];
     //   }
     //   deps.Projects = append(deps.Projects, project)
     // } else {
     //   logger.Warning("Cannot find " + lib + " archive... skipping")
     // }
   }
 }
 return deps,nil
}

func getDllCoordinate(path string) (project types.Projects, err error) {
  project = types.Projects{}
  // Check each collector in turn to see which gives us a good result.
  pc := path_collector{path: path}
  _, err = pc.GetPurl()
  if (err == nil) {
    project.Name,_ = pc.GetName();
    project.Version,_ = pc.GetVersion();
  }
  return project, err
}

func getArchiveCoordinate(path string) (project types.Projects, err error) {
  project = types.Projects{}
  // Check each collector in turn to see which gives us a good result.
  pc := pkgconfig_collector{path: path}
  _, err = pc.GetPurl()
  if (err == nil) {
    project.Name,_ = pc.GetName();
    project.Version,_ = pc.GetVersion();
  }
  return project, err
}
