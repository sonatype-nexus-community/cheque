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
package iq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"net/http"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/cheque/config"
	"github.com/sonatype-nexus-community/cheque/context"
	"github.com/sonatype-nexus-community/cheque/linker"
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/cheque/types"
	"github.com/sonatype-nexus-community/go-sona-types/cyclonedx"
	"github.com/sonatype-nexus-community/go-sona-types/iq"

	ossiTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
)

const getComponentVersions = "/api/v2/components/versions/"

// --------------------------------------------------------
// Cloned from go-sona-types iq.go

// ServerError is a custom error type that can be used to differentiate between
// regular errors and errors specific to handling IQ Server
type ServerError struct {
	Err     error
	Message string
}

func (i *ServerError) Error() string {
	if i.Err != nil {
		return fmt.Sprintf("An error occurred: %s, err: %s", i.Message, i.Err.Error())
	}
	return fmt.Sprintf("An error occurred: %s", i.Message)
}

type ServerErrorMissingLicense struct {
}

func (i *ServerErrorMissingLicense) Error() string {
	return "error accessing nexus iq server: No valid product license installed"
}

// Cloned from go-sona-types iq.go
// --------------------------------------------------------

type Iq struct {
	config    config.Config
	iqOptions iq.Options
}

func New(config config.Config) *Iq {
	return &Iq{
		config: config,
		iqOptions: iq.Options{
			User:       config.IQConfig.Username,
			Token:      config.IQConfig.Token,
			Server:     config.IQConfig.Server,
			Stage:      config.ChequeConfig.IQBuildStage,
			MaxRetries: *config.ChequeConfig.IQMaxRetries,
		},
	}
}

func (i Iq) ExtractCoordinates(deps types.ProjectList) []ossiTypes.Coordinate {
	results := make([]ossiTypes.Coordinate, 0)

	logger.Info("Identifying coordinates in IQ")

	for count := 0; count < len(deps.Projects); count++ {
		purl := deps.Projects[count]
		// path := deps.FileLookup[purl.ToString()]
		exists, _ := i.componentExists(purl)
		if exists {
			results = append(results, ossiTypes.Coordinate{
				Coordinates: "pkg:/conan/" + purl.Name + "@" + purl.Version,
				Reference:   "",
			})
		}
		fmt.Print(".")
	}
	fmt.Printf("\n")

	return results
}

func (i Iq) componentExists(purl packageurl.PackageURL) (bool, error) {

	var body = []byte(fmt.Sprintf(
		"{ \"format\": \"%s\", \"coordinates\": { \"name\": \"%s\" } }",
		"conan",
		purl.Name))
	var url = fmt.Sprintf("%s%s", i.config.IQConfig.Server, getComponentVersions)

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return false, &ServerError{
			Err:     err,
			Message: "Request to get component versions failed",
		}
	}
	req.SetBasicAuth(i.iqOptions.User, i.iqOptions.Token)
	// req.Header.Set("User-Agent", i.agent.GetUserAgent())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, &ServerError{
			Err:     err,
			Message: "There was an error communicating with Nexus IQ Server to get your internal application ID",
		}
	}

	if resp.StatusCode == http.StatusPaymentRequired {
		logger.Error("Error accessing Nexus IQ Server due to product license")
		return false, &ServerErrorMissingLicense{}
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, &ServerError{
				Err:     err,
				Message: "There was an error retrieving the bytes of the response for getting your internal application ID from Nexus IQ Server",
			}
		}

		var response []string
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return false, &ServerError{
				Err:     err,
				Message: "failed to unmarshal response",
			}
		}

		if len(response) != 0 {
			return true, nil
		}

		return false, &ServerError{
			Err:     fmt.Errorf("Unable to find coordinate matching: %s", &purl),
			Message: "Unable to retrieve matching coordinate",
		}
	}
	logger.Error(fmt.Sprintf("[%d] Error communicating with Nexus IQ Server application endpoint", resp.StatusCode))

	return false, &ServerError{
		Err:     fmt.Errorf("Unable to communicate with Nexus IQ Server, status code returned is: %d", resp.StatusCode),
		Message: "Unable to communicate with Nexus IQ Server",
	}
}

func (i Iq) getBinaryName(config config.Config) string {
	binaryName := context.GetBinaryName()
	if config.ChequeConfig.IQAppNamePrefix != "" {
		binaryName = config.ChequeConfig.IQAppNamePrefix + binaryName
	}
	return binaryName
}

func (i Iq) AuditWithIQ(lResults *linker.Results) {
	dx := cyclonedx.Default(logger.GetLogger())
	sbom := dx.FromCoordinates(lResults.Coordinates)

	binaryName := i.getBinaryName(i.config)

	runIQ := len(i.config.ChequeConfig.IQAppAllowList) == 0

	// Check to see if we're using allow list, if so, then check that
	// it's in there.
	if !runIQ {
		allowListArr := i.config.ChequeConfig.IQAppAllowList
		for _, app := range allowListArr[:] {
			if app == context.GetBinaryName() {
				runIQ = true
			}
		}
	}

	if runIQ {
		i.sendBomToIQ(i.config, binaryName, sbom)
	} else {
		logger.GetLogger().Info("Skipping sending bom to IQ due to app allow list.")
	}
}

func (i Iq) sendBomToIQ(config config.Config, binaryName string, sbom string) {
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
	i.showPolicyActionMessage(result, os.Stdout)
}

func (i Iq) showPolicyActionMessage(res iq.StatusURLResult, writer io.Writer) {
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
