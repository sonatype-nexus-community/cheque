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

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
)

func TestCreateEmptyObject(t *testing.T) {
	conf := setup(t)
	conf.CreateOrReadConfigFile()
	iqExists, err := os.Stat(conf.getIQConfig())
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(iqExists.Name(), types.IQServerConfigFileName) {
		t.Errorf("File not created properly, expected %s but got %s", types.IQServerConfigFileName, iqExists.Name())
	}

	ossiExists, err := os.Stat(conf.getOssiConfig())
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(ossiExists.Name(), types.OssIndexConfigFileName) {
		t.Errorf("File not created properly, expected %s but got %s", types.OssIndexConfigFileName, ossiExists.Name())
	}
	teardown(conf)
}

func TestReadsOssiConfigProperly(t *testing.T) {
	conf := setup(t)
	writeDataToConfig(filepath.Join(conf.options.Directory, types.OssIndexDirName), types.OssIndexConfigFileName,
		"Username: \"something\"\nToken: \"something1\"")
	conf.CreateOrReadConfigFile()

	if conf.OSSIndexConfig.Username != "something" {
		t.Errorf("username wasn't in config, expected %s but got %s", "something", conf.OSSIndexConfig.Username)
	}

	if conf.OSSIndexConfig.Token != "something1" {
		t.Errorf("Token wasn't in config, expected %s but got %s", "something1", conf.OSSIndexConfig.Token)
	}

	teardown(conf)
}

func TestReadsIQConfigProperly(t *testing.T) {
	conf := setup(t)
	writeDataToConfig(filepath.Join(conf.options.Directory, types.IQServerDirName), types.IQServerConfigFileName,
		"Username: \"something\"\nToken: \"somethingtoken\"\nServer: \"somethingserver\"")
	conf.CreateOrReadConfigFile()

	if conf.IQConfig.Username != "something" {
		t.Errorf("Username wasn't in config, expected %s but got %s", "something", conf.IQConfig.Username)
	}

	if conf.IQConfig.Token != "somethingtoken" {
		t.Errorf("Token wasn't in config, expected %s but got %s", "somethingtoken", conf.IQConfig.Token)
	}

	if conf.IQConfig.Server != "somethingserver" {
		t.Errorf("Server wasn't in config, expected %s but got %s", "somethingserver", conf.IQConfig.Server)
	}

	teardown(conf)
}

func TestReadsChequeConfigProperly(t *testing.T) {
	conf := setup(t)

	writeDataToConfig(filepath.Join(conf.options.Directory, ChequeConfigDirectory), ChequeConfigFile,
		"Create-Conan-Files: true\nUse-IQ: true\nIQ-Build-Stage: build\nIQ-App-Prefix: cheque\nIQ-App-Allow-List:\n  - cheque\nIQ-Max-Retries: 30")
	conf.CreateOrReadConfigFile()

	if !*conf.ChequeConfig.CreateConanFiles {
		t.Errorf("Create-Conan-Files wasn't in config, expected %s but got %s", "true", strconv.FormatBool(*conf.ChequeConfig.CreateConanFiles))
	}

	if !*conf.ChequeConfig.UseIQ {
		t.Errorf("Use-IQ wasn't in config, expected %s but got %s", "true", strconv.FormatBool(*conf.ChequeConfig.UseIQ))
	}

	if conf.ChequeConfig.IQBuildStage != "build" {
		t.Errorf("IQ-Build-Stage wasn't in config, expected  %s but got %s", "build", conf.ChequeConfig.IQBuildStage)
	}

	if conf.ChequeConfig.IQAppNamePrefix != "cheque" {
		t.Errorf("IQ-App-Prefix wasn't in config, expected  %s but got %s", "cheque", conf.ChequeConfig.IQAppNamePrefix)
	}

	if *conf.ChequeConfig.IQMaxRetries != 30 {
		t.Errorf("IQ-Max-Retries wasn't in config, expected  %s but got %d", "30", conf.ChequeConfig.IQMaxRetries)
	}

	if len(conf.ChequeConfig.IQAppAllowList) != 1 || conf.ChequeConfig.IQAppAllowList[0] != "cheque"  {
		t.Errorf("IQ-App-Allow-List wasn't in config, expected  %s but got %s", "[cheque]", conf.ChequeConfig.IQAppAllowList)
	}

	teardown(conf)
}

func TestLocalFileOverrides(t *testing.T) {
	conf := setup(t)

	writeDataToConfig(filepath.Join(conf.options.Directory, ChequeConfigDirectory), ChequeConfigFile,
		"Create-Conan-Files: true\nUse-IQ: true\nIQ-Build-Stage: build\nIQ-App-Prefix: cheque\nIQ-App-Allow-List:\n  - cheque\nIQ-Max-Retries: 30")

	writeDataToConfig(conf.options.WorkingDirectory, LocalChequeConfigFile,
		"Create-Conan-Files: false\nUse-IQ: false\nIQ-Build-Stage: stage\nIQ-App-Prefix: whatwhat\nIQ-App-Allow-List:\n  - ohnoyoudidnt\nIQ-Max-Retries: 120")

	conf.CreateOrReadConfigFile()

	if *conf.ChequeConfig.CreateConanFiles {
		t.Errorf("Create-Conan-Files wasn't in config, expected %s but got %s", "false", strconv.FormatBool(*conf.ChequeConfig.CreateConanFiles))
	}

	if *conf.ChequeConfig.UseIQ {
		t.Errorf("Use-IQ wasn't in config, expected %s but got %s", "false", strconv.FormatBool(*conf.ChequeConfig.UseIQ))
	}

	if conf.ChequeConfig.IQBuildStage != "stage" {
		t.Errorf("IQ-Build-Stage wasn't in config, expected  %s but got %s", "stage", conf.ChequeConfig.IQBuildStage)
	}

	if conf.ChequeConfig.IQAppNamePrefix != "whatwhat" {
		t.Errorf("IQ-App-Prefix wasn't in config, expected  %s but got %s", "whatwhat", conf.ChequeConfig.IQAppNamePrefix)
	}

	if *conf.ChequeConfig.IQMaxRetries != 120 {
		t.Errorf("IQ-Max-Retries wasn't in config, expected  %s but got %d", "120", conf.ChequeConfig.IQMaxRetries)
	}

	if len(conf.ChequeConfig.IQAppAllowList) != 1 || conf.ChequeConfig.IQAppAllowList[0] != "ohnoyoudidnt"  {
		t.Errorf("IQ-App-Allow-List wasn't in config, expected  %s but got %s", "[ohnoyoudidnt]", conf.ChequeConfig.IQAppAllowList)
	}

	teardown(conf)
}

func writeDataToConfig(directory string, filename string, data string) {
	b:= []byte(data)
	os.MkdirAll(directory, 755)
	ioutil.WriteFile(filepath.Join(directory, filename), b,0644)
}

func teardown(config *Config) {
	_ = os.RemoveAll(config.options.Directory)
}

func setup(t *testing.T) *Config {
	logLady, _ := test.NewNullLogger()
	options := Options{}
	tempDir, err := ioutil.TempDir("", "testconfig")
	if err != nil {
		t.Error(err)
	}
	options.Directory = tempDir
	options.WorkingDirectory = tempDir
	conf := New(logLady, options)

	return conf
}
