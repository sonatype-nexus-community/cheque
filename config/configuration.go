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
	"github.com/sonatype-nexus-community/cheque/logger"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

type OSSIConfig struct {
	Username string `yaml:Username`
	Password string `yaml:Password`
}

type IQConfig struct {
	Server   string `yaml:"Server"`
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

type Config struct {
	logger *logrus.Logger
	OSSIndexConfig OSSIConfig
	IQConfig IQConfig
}

func New(logger *logrus.Logger) *Config {
	return &Config{logger: logger}
}

func (c *Config) CreateOrReadConfigFile() {

	c.createDirectory(types.IQServerDirName)
	c.createDirectory(types.OssIndexDirName)

	if !fileExists(c.getIQConfig()) {
		c.writeDefaultIQConfig()
	}
	if !fileExists(c.getOssiConfig()) {
		c.writeDefaultOssiConfig()
	}

	c.readConfig()
}

func (c Config) createDirectory(directory string) {
	home, _ := os.UserHomeDir()
	fileDirectory := filepath.Join(home, directory)
	err := os.MkdirAll(fileDirectory, os.ModePerm)
	if err != nil {
		logger.Error("Couldn't create config directory: " + fileDirectory)
	}
}

//Gets the default location for the config file
func (c Config) getIQConfig() string {
	home, _ := os.UserHomeDir()
	filePath := filepath.Join(home,types.IQServerDirName, types.IQServerConfigFileName)
	return filePath
}

func (c Config) getOssiConfig() string {
	home, _ := os.UserHomeDir()
	filePath := filepath.Join(home,types.OssIndexDirName, types.OssIndexConfigFileName)
	return filePath
}

func (c *Config) readConfig() {
	iqBytes, err := ioutil.ReadFile(c.getIQConfig())
	if err != nil {
		c.logger.Error(err)
	}
	iqConfig := IQConfig{}
	yaml.Unmarshal(iqBytes, &iqConfig)

	ossiBytes, err := ioutil.ReadFile(c.getOssiConfig())
	if err != nil {
		c.logger.Error(err)
	}
	ossiConfig := OSSIConfig{}
	yaml.Unmarshal(ossiBytes, &ossiConfig)

	c.OSSIndexConfig = ossiConfig
	c.IQConfig = iqConfig
}


func (c Config) writeDefaultOssiConfig() {
	ossiConfig, _ := yaml.Marshal(OSSIConfig{})
	err := ioutil.WriteFile(c.getOssiConfig(), ossiConfig, 0644)
	if err != nil {
		c.logger.WithField("configFile", c.getOssiConfig()).Error("Could not create OSSIndeConfig.")
	}
}

func (c Config) writeDefaultIQConfig() {
	iqConfig, _ := yaml.Marshal(IQConfig{})
	err := ioutil.WriteFile(c.getIQConfig(), iqConfig, 0644)
	if err != nil {
		c.logger.WithField("configFile", c.getIQConfig()).Error("Could not create IQConfig.")
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
