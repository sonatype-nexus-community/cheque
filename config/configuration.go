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
	CreateConanFiles bool     `yaml:"Create-Conan-Files"`
	UseIQ            bool     `yaml:"Use-IQ"`
	IQMaxRetries     int      `yaml:"IQ-Max-Retries"`
	IQBuildStage     string   `yaml:"IQ-Build-Stage"`
	IQAppNamePrefix  string   `yaml:"IQ-App-Prefix"`
	IQAppAllowList   []string `yaml:"IQ-App-Allow-List"`
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
		c.writeDefaultConfig(ChequeConfigDirectory, ChequeConfig{
			CreateConanFiles: false,
			UseIQ:            false,
			IQMaxRetries:     30,
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

	c.OSSIndexConfig = ossiConfig
	c.IQConfig = iqConfig
	c.ChequeConfig = chequeConfig
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
