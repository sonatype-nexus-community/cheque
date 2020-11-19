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
	"github.com/sonatype-nexus-community/cheque/audit"
	"github.com/sonatype-nexus-community/cheque/conan"
	"github.com/sonatype-nexus-community/cheque/config"
	"github.com/sonatype-nexus-community/cheque/context"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/go-sona-types/cyclonedx"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	"os"
	"os/exec"
)

func main() {
	args := []string{}

	//Will check for config and create if necessary
	options := config.Options{}
	myConfig := config.New(logger.GetLogger(), options)
	myConfig.CreateOrReadConfigFile()

	// Remove cheque custom arguments
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-Werror=cheque":
		default:
			args = append(args, arg)
		}
	}

	myLinker := linker.New(myConfig.OSSIndexConfig)
	results := myLinker.DoLink(args)
	if results.Count > 0 {
		if context.ExitWithError() {
			fmt.Fprintf(os.Stderr, "Error: Vulnerable dependencies found: %v\n", results.Count)
			os.Exit(results.Count)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Vulnerable dependencies found: %v\n", results.Count)
		}
	}

	generateConanFiles(*myConfig, results)
	generateCycloneDx(*myConfig, results)


	switch context.GetCommand() {
	case "cheque":
		break
	default:
		var cmdPath = fmt.Sprint("/usr/bin/", context.GetCommand())

		_, err := os.Stat(cmdPath)
		if err != nil {
			logger.Fatal("Cannot find official command: " + cmdPath)
		} else {
			externalCmd := exec.Command(cmdPath, args...)

			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr

			if err := externalCmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					logger.Fatal(fmt.Sprintf("There was an issue running the command %s, and the issue is %v", context.GetCommand(), os.Stderr))
					os.Exit(exitError.ExitCode())
				}
			}
		}
		break
	}

	os.Exit(0)
}

func generateCycloneDx(config config.Config, lResults *linker.Results) {
	if config.ChequeConfig.UseIQ {
		dx := cyclonedx.New(logger.GetLogger(), cyclonedx.Options{})
		sbom := dx.FromCoordinates(lResults.Coordinates)
		iqOptions := iq.Options{
			User:        config.IQConfig.Username,
			Token:       config.IQConfig.Token,
			Application: "cheque",
			Server:      config.IQConfig.Server,
			Stage: "build",
		}
		server, err := iq.New(logger.GetLogger(), iqOptions)
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("error creating connection to IQ")
			return
		}
		result, iqerr := server.AuditWithSbom(sbom)
		if iqerr != nil {
			logger.GetLogger().WithField("err", iqerr).Error("error submitting bom")
		}
		logger.GetLogger().WithField("result", result).Info("Completed submittion of bom to IQ")
	}
}

func generateConanFiles(myConfig config.Config, results *linker.Results) {
	if myConfig.ChequeConfig.CreateConanFiles {
		myAudit := audit.New(myConfig.OSSIndexConfig)
		purls := myAudit.GetPurls(results.LibPaths, results.Libs, results.Files)
		options := conan.Options{
			BinaryName: context.GetBinaryName(),
		}
		generator := conan.New(logger.GetLogger(), options)
		err := generator.CheckOrCreateConanFile(purls)
		if err != nil {
			logger.GetLogger().WithField("err", err).Warnf("Something went wrong writing conan files")
		}
	}
}
