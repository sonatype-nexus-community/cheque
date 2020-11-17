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
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ChequeConfig struct {
	ossiConfig OSSIConfig
	iqConfig IQConfig
}

type OSSIConfig struct {
	Username string `yaml:Username`
	Password string `yaml:Password`
}

type IQConfig struct {
	Server   string `yaml:"Server"`
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

func CreateOrReadConfigFile(logger *logrus.Logger) ChequeConfig {

	createDirectory(logger, types.IQServerDirName)
	createDirectory(logger, types.OssIndexDirName)

	if fileExists(getIQConfig()) {
		//read in config
	} else {
		writeDefaultIQConfig(logger)
	}

	if fileExists(getOssiConfig()) {
		//read in config
	} else {
		writeDefaultOssiConfig(logger)
	}

	return readConfig(logger)
}

func createDirectory(logger *logrus.Logger, directory string) {
	home, _ := os.UserHomeDir()
	fileDirectory := filepath.Join(home, directory)
	err := os.MkdirAll(fileDirectory, os.ModePerm)
	if err != nil {
		logger.Error("Couldn't create config directory: " + fileDirectory)
	}
}

//Gets the default location for the config file
func getIQConfig() string {
	home, _ := os.UserHomeDir()
	filePath := filepath.Join(home,types.IQServerDirName, types.IQServerConfigFileName)
	return filePath
}

func getOssiConfig() string {
	home, _ := os.UserHomeDir()
	filePath := filepath.Join(home,types.OssIndexDirName, types.OssIndexConfigFileName)
	return filePath
}

func writeDefaultOssiConfig(logger *logrus.Logger) {
	ossiConfig, _ := yaml.Marshal(OSSIConfig{})
	err := ioutil.WriteFile(getOssiConfig(), ossiConfig, 0644)
	if err != nil {
		logger.Error(err)
	}
}

func readConfig(logger *logrus.Logger) ChequeConfig{
	iqBytes, err := ioutil.ReadFile(getIQConfig())
	if err != nil {
		logger.Error(err)
	}
	iqConfig := IQConfig{}
	yaml.Unmarshal(iqBytes, &iqConfig)

	ossiBytes, err := ioutil.ReadFile(getOssiConfig())
	if err != nil {
		logger.Error(err)
	}
	ossiConfig := OSSIConfig{}
	yaml.Unmarshal(ossiBytes, &ossiConfig)

	return ChequeConfig{
		ossiConfig: ossiConfig,
		iqConfig: iqConfig,
	}
}

func writeDefaultIQConfig(logger *logrus.Logger) {
	iqConfig, _ := yaml.Marshal(IQConfig{})
	err := ioutil.WriteFile(getIQConfig(), iqConfig, 0644)
	if err != nil {
		logger.Error(err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
