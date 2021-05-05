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
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sonatype-nexus-community/cheque/audit"
	"github.com/sonatype-nexus-community/cheque/conan"
	"github.com/sonatype-nexus-community/cheque/config"
	"github.com/sonatype-nexus-community/cheque/context"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/scanner"
	"github.com/sonatype-nexus-community/go-sona-types/cyclonedx"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
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
		case "--version":
			// Bail early
			runWrappedCommand(os.Args[1:])
			os.Exit(0)
		case "-Werror=cheque":
		default:
			args = append(args, arg)
		}
	}

	var results *linker.Results

	// If we are running in "compiler mode" run the cheque linker
	if !context.GetChequeScan() {
		myLinker := linker.New(myConfig.OSSIndexConfig)
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
		myScanner := scanner.New(myConfig.OSSIndexConfig)
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
		generateConanFiles(*myConfig, results)
		auditWithIQ(*myConfig, results)
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

func auditWithIQ(config config.Config, lResults *linker.Results) {
	if *config.ChequeConfig.UseIQ {
		dx := cyclonedx.Default(logger.GetLogger())
		sbom := dx.FromCoordinates(lResults.Coordinates)

		binaryName := context.GetBinaryName()
		if config.ChequeConfig.IQAppNamePrefix != "" {
			binaryName = config.ChequeConfig.IQAppNamePrefix + binaryName
		}

		runIQ := len(config.ChequeConfig.IQAppAllowList) == 0

		// Check to see if we're using allow list, if so, then check that
		// it's in there.
		if !runIQ {
			allowListArr := config.ChequeConfig.IQAppAllowList
			for _, app := range allowListArr[:] {
				if app == context.GetBinaryName() {
					runIQ = true
				}
			}
		}

		if runIQ {
			sendBomToIQ(config, binaryName, sbom)
		} else {
			logger.GetLogger().Info("Skipping sending bom to IQ due to app allow list.")
		}
	}
}

func sendBomToIQ(config config.Config, binaryName string, sbom string) {
	iqOptions := iq.Options{
		User:        config.IQConfig.Username,
		Token:       config.IQConfig.Token,
		Application: binaryName,
		Server:      config.IQConfig.Server,
		Stage:       config.ChequeConfig.IQBuildStage,
		MaxRetries:  *config.ChequeConfig.IQMaxRetries,
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
	logger.GetLogger().WithField("result", result).Info("Completed submission of bom to IQ")

	// print summary
	showPolicyActionMessage(result, os.Stdout)
}

func showPolicyActionMessage(res iq.StatusURLResult, writer io.Writer) {
	_, _ = fmt.Fprintln(writer)
	switch res.PolicyAction {
	case iq.PolicyActionFailure:
		_, _ = fmt.Fprintln(writer, "There are policy violations to clean up")
		_, _ = fmt.Fprintln(writer, "Report URL: ", res.AbsoluteReportHTMLURL)
	case iq.PolicyActionWarning:
		_, _ = fmt.Fprintln(writer, "There are policy warnings to investigate")
		_, _ = fmt.Fprintln(writer, "Report URL: ", res.AbsoluteReportHTMLURL)
	default:
		_, _ = fmt.Fprintln(writer, "No policy violations reported for this audit")
		_, _ = fmt.Fprintln(writer, "Report URL: ", res.AbsoluteReportHTMLURL)
	}
}

func generateConanFiles(myConfig config.Config, results *linker.Results) {
	if *myConfig.ChequeConfig.CreateConanFiles {
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
