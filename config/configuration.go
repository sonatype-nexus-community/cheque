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

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"gopkg.in/yaml.v3"
)

var (
	ChequeConfigDirectory = ".cheque"
	ChequeConfigFile = "config"
	LocalChequeConfigFile = ".cheque-config"
)

type OSSIConfig struct {
	Username string `yaml:"Username"`
	Token string `yaml:"Token"`
}

type IQConfig struct {
	Server   string `yaml:"Server"`
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

type ChequeConfig struct {
	CreateConanFiles *bool     `yaml:"Create-Conan-Files",omitempty`
	UseIQ            *bool     `yaml:"Use-IQ",omitempty`
	IQMaxRetries     *int      `yaml:"IQ-Max-Retries",omitempty`
	IQBuildStage     string   `yaml:"IQ-Build-Stage"`
	IQAppNamePrefix  string   `yaml:"IQ-App-Prefix"`
	IQAppAllowList   []string `yaml:"IQ-App-Allow-List",omitempty`
}

type Config struct {
	logger         *logrus.Logger
	options        Options
	OSSIndexConfig OSSIConfig
	IQConfig       IQConfig
	ChequeConfig   ChequeConfig
}

type Options struct {
	Directory string
}

func New(logger *logrus.Logger, options Options) *Config {
	if options.Directory == "" {
		home, _ := os.UserHomeDir()
		options.Directory = home
	}

	return &Config{logger: logger, options: options}
}

func (c *Config) CreateOrReadConfigFile() {
	if !fileExists(c.getIQConfig()) {
		c.writeDefaultConfig(types.IQServerDirName, IQConfig{}, c.getIQConfig())
	}
	if !fileExists(c.getOssiConfig()) {
		c.writeDefaultConfig(types.OssIndexDirName, OSSIConfig{}, c.getOssiConfig())
	}
	if !fileExists(c.getChequeConfig()) {

		falseBoolean := false
		retryDefault := 30
		c.writeDefaultConfig(ChequeConfigDirectory, ChequeConfig{
			CreateConanFiles: &falseBoolean,
			UseIQ:            &falseBoolean,
			IQMaxRetries:     &retryDefault,
			IQBuildStage: "build",
			IQAppNamePrefix: "cheque-",
		}, c.getChequeConfig())
	}

	c.readConfig()
}

func (c Config) createDirectory(directory string) {
	fileDirectory := filepath.Join(c.options.Directory, directory)
	err := os.MkdirAll(fileDirectory, os.ModePerm)
	if err != nil {
		c.logger.WithField("err", err).Error("Couldn't create config directory: " + fileDirectory)
	}
}

//Gets the default location for the config file
func (c Config) getIQConfig() string {
	return filepath.Join(c.options.Directory, types.IQServerDirName, types.IQServerConfigFileName)
}

func (c Config) getChequeConfig() string {
	return filepath.Join(c.options.Directory, ChequeConfigDirectory, ChequeConfigFile)
}

func (c Config) getWorkingDirectoryChequeConfig() string {
	getwd, _ := os.Getwd()
	return filepath.Join(getwd, LocalChequeConfigFile)
}

func (c Config) getOssiConfig() string {
	return filepath.Join(c.options.Directory, types.OssIndexDirName, types.OssIndexConfigFileName)
}

func (c *Config) readConfig() {
	iqBytes, err := ioutil.ReadFile(c.getIQConfig())
	if err != nil {
		c.logger.WithField("err", err).Error(err)
	}
	iqConfig := IQConfig{}
	_ = yaml.Unmarshal(iqBytes, &iqConfig)

	ossiBytes, err := ioutil.ReadFile(c.getOssiConfig())
	if err != nil {
		c.logger.WithField("err", err).Error(err)
	}
	ossiConfig := OSSIConfig{}
	_ = yaml.Unmarshal(ossiBytes, &ossiConfig)

	chequeBytes, err := ioutil.ReadFile(c.getChequeConfig())
	if err != nil {
		c.logger.WithField("err", err).Error(err)
	}
	chequeConfig := ChequeConfig{}
	_ = yaml.Unmarshal(chequeBytes, &chequeConfig)

	chequeConfig = c.overrideWithLocalConfig(chequeConfig)

	c.OSSIndexConfig = ossiConfig
	c.IQConfig = iqConfig
	c.ChequeConfig = chequeConfig
}

func (c Config) overrideWithLocalConfig(config ChequeConfig) ChequeConfig {
	if fileExists(c.getWorkingDirectoryChequeConfig()) {
		localChequeConfigBytes, err := ioutil.ReadFile(c.getWorkingDirectoryChequeConfig())
		if err != nil {
			c.logger.WithField("err", err).Error(err)
		}
		localConfig := ChequeConfig{}
		_ = yaml.Unmarshal(localChequeConfigBytes, &localConfig)

		if localConfig.UseIQ != nil {
			config.UseIQ = localConfig.UseIQ
		}

		if localConfig.CreateConanFiles != nil {
			config.CreateConanFiles = localConfig.CreateConanFiles
		}

		if localConfig.IQAppNamePrefix != "" {
			config.IQAppNamePrefix = localConfig.IQAppNamePrefix
		}

		if localConfig.IQBuildStage != "" {
			config.IQBuildStage = localConfig.IQBuildStage
		}

		if localConfig.IQMaxRetries != nil {
			config.IQMaxRetries = localConfig.IQMaxRetries
		}

		if len(localConfig.IQAppAllowList) > 0 {
			config.IQAppAllowList = localConfig.IQAppAllowList
		}

		return config
	}
	return config
}

func (c Config) writeDefaultConfig(directoryName string, config interface{}, pathToConfig string) {
	c.createDirectory(directoryName)
	myConfig, _ := yaml.Marshal(config)
	err := ioutil.WriteFile(pathToConfig, myConfig, 0644)
	if err != nil {
		c.logger.WithFields(
			logrus.Fields{
				"configFile": pathToConfig,
				"err":        err,
			}).Error("Could not create OSSIndexConfig.")
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
