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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sonatype-nexus-community/cheque/audit"
	"github.com/sonatype-nexus-community/cheque/bom"
	"github.com/sonatype-nexus-community/cheque/conan"
	"github.com/sonatype-nexus-community/cheque/config"
	"github.com/sonatype-nexus-community/cheque/context"
	"github.com/sonatype-nexus-community/cheque/iq"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/scanner"
	"github.com/sonatype-nexus-community/go-sona-types/cyclonedx"
)

func main() {
	args := []string{}

	//Will check for config and create if necessary
	options := config.Options{}
	myConfig := config.New(logger.GetLogger(), options)
	myConfig.CreateOrReadConfigFile()
	myConfig.CreateOrReadCacheFile()

	// Remove cheque custom arguments
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version":
			// Bail early
			runWrappedCommand(os.Args[1:])
			os.Exit(0)
		case "-Werror=cheque":
		default:
			args = append(args, arg)
		}
	}

	if myConfig.ChequeConfig.ShouldUseIQ() {
		doIqRun(*myConfig, args)
	} else {
		doOssiRun(*myConfig, args)
	}
}

/**
 * Run an analysis using IQ. This will avoid any communications with OSS Index
 */
func doIqRun(myConfig config.Config, args []string) {
	var results *linker.Results

	// If we are running in "compiler mode" run the cheque linker
	if !context.GetChequeScan() {
		myLinker := linker.New(myConfig.OSSIndexConfig, myConfig.ConanPackages)
		results = myLinker.GetArguments(args)
		if results.Count > 0 {
			if context.ExitWithError() {
				fmt.Fprintf(os.Stderr, "Error: Vulnerable dependencies found: %v\n", results.Count)
				os.Exit(results.Count)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Vulnerable dependencies found: %v\n", results.Count)
			}
		}
	}

	// If we are running in "scan mode" run the cheque scanner
	if context.GetChequeScan() {
		myScanner := scanner.New(myConfig.OSSIndexConfig, myConfig.ConanPackages)
		results = myScanner.GetArguments(context.GetChequeScanPath(), args)
		if results.Count > 0 {
			if context.ExitWithError() {
				fmt.Fprintf(os.Stderr, "Error: Vulnerable dependencies found: %v\n", results.Count)
				os.Exit(results.Count)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Vulnerable dependencies found: %v\n", results.Count)
			}
		}
	}

	if results != nil && (len(results.Libs) > 0 || len(results.Files) > 0) {
		var projectList, _ = bom.CreateBomFromRoot(results.LibPaths, results.Libs, results.Files, context.GetChequeScanPath())

		i := iq.New(myConfig)
		results.Coordinates = i.ExtractCoordinates(projectList)

		generateConanFiles(myConfig, results)
		generateSbom(myConfig, results)
		i.AuditWithIQ(results)
	}

	// If we are running in "compiler mode" run the real linker
	if !context.GetChequeScan() {
		switch context.GetCommand() {
		case "cheque":
			break
		default:
			runWrappedCommand(args)
			break
		}
	}
	os.Exit(0)
}

/**
 * Run an analysis against OSS Index.
 */
func doOssiRun(myConfig config.Config, args []string) {
	var results *linker.Results

	// If we are running in "compiler mode" run the cheque linker
	if !context.GetChequeScan() {
		myLinker := linker.New(myConfig.OSSIndexConfig, myConfig.ConanPackages)
		results = myLinker.DoLink(args)
		if results.Count > 0 {
			if context.ExitWithError() {
				fmt.Fprintf(os.Stderr, "Error: Vulnerable dependencies found: %v\n", results.Count)
				os.Exit(results.Count)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Vulnerable dependencies found: %v\n", results.Count)
			}
		}
	}

	// If we are running in "scan mode" run the cheque scanner
	if context.GetChequeScan() {
		myScanner := scanner.New(myConfig.OSSIndexConfig, myConfig.ConanPackages)
		results = myScanner.DoScan(context.GetChequeScanPath(), args)
		if results.Count > 0 {
			if context.ExitWithError() {
				fmt.Fprintf(os.Stderr, "Error: Vulnerable dependencies found: %v\n", results.Count)
				os.Exit(results.Count)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Vulnerable dependencies found: %v\n", results.Count)
			}
		}
	}

	if results != nil && (len(results.Libs) > 0 || len(results.Files) > 0) {
		generateConanFiles(myConfig, results)
		generateSbom(myConfig, results)
	}

	// If we are running in "compiler mode" run the real linker
	if !context.GetChequeScan() {
		switch context.GetCommand() {
		case "cheque":
			break
		default:
			runWrappedCommand(args)
			break
		}
	}
	os.Exit(0)
}

func runWrappedCommand(args []string) {
	var cmdPath = getWrappedCommand()

	_, err := os.Stat(cmdPath)
	if err != nil {
		logger.Fatal("Cannot find official command: " + cmdPath)
	} else {
		externalCmd := exec.Command(cmdPath, args...)

		externalCmd.Stdout = os.Stdout
		externalCmd.Stderr = os.Stderr

		if err := externalCmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				logger.Fatal(fmt.Sprintf("There was an issue running the command %s, and the issue is %v", cmdPath, os.Stderr))
				os.Exit(exitError.ExitCode())
			}
		}
	}
}

func getWrappedCommand() (cmdPath string) {
	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		cmdPath = filepath.Join(path, context.GetCommand())

		// Ignore if this is a symlink to cheque
		realPath, err := filepath.EvalSymlinks(cmdPath)
		if err == nil && filepath.Base(realPath) == "cheque" {
			continue
		}
		_, err = os.Stat(cmdPath)
		if err == nil {
			break // Found the real binary
		}
		cmdPath = "" // Don't report this one as the wrapped binary
	}
	return
}

func generateSbom(myConfig config.Config, results *linker.Results) {
	if context.GetSbom() || myConfig.ChequeConfig.ShouldCreateSbom() {
		dx := cyclonedx.Default(logger.GetLogger())
		sbom := dx.FromCoordinates(results.Coordinates)
		fname := context.GetBinaryName()
		if fname == "a.out" {
			fname = "cheque"
			if context.GetChequeScan() {
				fname = context.GetChequeScanPath() + "/" + fname
			}
		}
		fname = fname + "-bom.xml"
		f, err := os.Create(fname)
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("error exporting sbom")
			return
		}
		defer f.Close()
		_, err = f.WriteString(sbom)
		if err != nil {
			logger.GetLogger().WithField("err", err).Error("error exporting sbom")
			return
		}
		f.Sync()
		fmt.Printf("exported SBOM to %s\n", fname)
	}
}

func generateConanFiles(myConfig config.Config, results *linker.Results) {
	if myConfig.ChequeConfig.ShouldCreateConanFiles() {
		myAudit := audit.New(myConfig.OSSIndexConfig, myConfig.ConanPackages)
		purls, _ := myAudit.GetPurls(results.LibPaths, results.Libs, results.Files)
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
