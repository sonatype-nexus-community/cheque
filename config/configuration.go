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
	"encoding/json"
	myTypes "github.com/sonatype-nexus-community/cheque/types"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"gopkg.in/yaml.v3"
)

var (
	ChequeConfigDirectory = ".cheque"
	ChequeConfigFile      = "config"
	LocalChequeConfigFile = ".cheque-config"
)

type OSSIConfig struct {
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

type IQConfig struct {
	Server   string `yaml:"Server"`
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

type ChequeConfig struct {
	CreateSbom       *bool    `yaml:"Create-SBOM",omitempty`
	CreateConanFiles *bool    `yaml:"Create-Conan-Files",omitempty`
	UseIQ            *bool    `yaml:"Use-IQ",omitempty`
	IQMaxRetries     *int     `yaml:"IQ-Max-Retries",omitempty`
	IQBuildStage     string   `yaml:"IQ-Build-Stage"`
	IQAppNamePrefix  string   `yaml:"IQ-App-Prefix"`
	IQAppAllowList   []string `yaml:"IQ-App-Allow-List",omitempty`
}

type ConanPackage struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ConanPackages struct {
	Lookup map[string]*ConanPackage
}

type Config struct {
	logger         *logrus.Logger
	options        Options
	OSSIndexConfig OSSIConfig
	IQConfig       IQConfig
	ChequeConfig   ChequeConfig
	ChequeCache    []string
	ConanPackages  ConanPackages
}

type Options struct {
	Directory        string
	WorkingDirectory string
}

func New(logger *logrus.Logger, options Options) *Config {
	if options.Directory == "" {
		home, _ := os.UserHomeDir()
		options.Directory = home
	}

	if options.WorkingDirectory == "" {
		getwd, _ := os.Getwd()
		options.WorkingDirectory = getwd
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
			CreateSbom:       &falseBoolean,
			CreateConanFiles: &falseBoolean,
			UseIQ:            &falseBoolean,
			IQMaxRetries:     &retryDefault,
			IQBuildStage:     "build",
			IQAppNamePrefix:  "cheque-",
		}, c.getChequeConfig())
	}

	c.readConfig()
}

func (c *Config) CreateOrReadCacheFile() {
	if !fileExists(c.getConanCache()) {
		c.writeConanCache(myTypes.ConanCacheDirName, c.getConanCache())
	}

	c.readCache()
}

func (c ChequeConfig) ShouldCreateSbom() bool {
	return c.CreateSbom != nil && *c.CreateSbom
}

func (c ChequeConfig) ShouldCreateConanFiles() bool {
	return c.CreateConanFiles != nil && *c.CreateConanFiles
}

func (c ChequeConfig) ShouldUseIQ() bool {
	return c.UseIQ != nil && *c.UseIQ
}

func (c Config) createDirectory(directory string) {
	fileDirectory := filepath.Join(c.options.Directory, directory)
	err := os.MkdirAll(fileDirectory, os.ModePerm)
	if err != nil {
		c.logger.WithField("err", err).Error("Couldn't create config directory: " + fileDirectory)
	}
}

//Gets the default location for the config file
func (c Config) getConanCache() string {
	return filepath.Join(c.options.Directory, myTypes.ConanCacheDirName, myTypes.ConanCacheFileName)
}

func (c Config) getIQConfig() string {
	return filepath.Join(c.options.Directory, types.IQServerDirName, types.IQServerConfigFileName)
}

func (c Config) getChequeConfig() string {
	return filepath.Join(c.options.Directory, ChequeConfigDirectory, ChequeConfigFile)
}

func (c Config) getWorkingDirectoryChequeConfig() string {

	return filepath.Join(c.options.WorkingDirectory, LocalChequeConfigFile)
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

func (c *Config) readCache() {
	conanBytes, err := ioutil.ReadFile(c.getConanCache())
	if err != nil {
		c.logger.WithField("err", err).Error(err)
	}
	c.ConanPackages.Lookup = make(map[string]*ConanPackage)
	_ = yaml.Unmarshal(conanBytes, &c.ConanPackages.Lookup)
}

func (c Config) overrideWithLocalConfig(config ChequeConfig) ChequeConfig {
	if fileExists(c.getWorkingDirectoryChequeConfig()) {
		localChequeConfigBytes, err := ioutil.ReadFile(c.getWorkingDirectoryChequeConfig())
		if err != nil {
			c.logger.WithField("err", err).Error(err)
			return config
		}
		localConfig := ChequeConfig{}
		_ = yaml.Unmarshal(localChequeConfigBytes, &localConfig)

		if localConfig.UseIQ != nil {
			config.UseIQ = localConfig.UseIQ
		}

		if localConfig.CreateConanFiles != nil {
			config.CreateConanFiles = localConfig.CreateConanFiles
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

func (c Config) writeConanCache(directoryName string, pathToConfig string) {
	c.ConanPackages.Lookup = make(map[string]*ConanPackage)

	url := "https://api.github.com/repositories/204671232/contents/recipes"

	conanClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "cheque")

	res, getErr := conanClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	packages := make([]ConanPackage, 0)

	jsonErr := json.Unmarshal(body, &packages)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// Initialize the identity database
	for _, v := range packages {
		c.ConanPackages.Lookup[v.Name] = &v
	}

	c.createDirectory(directoryName)
	myConfig, _ := yaml.Marshal(c.ConanPackages)
	err = ioutil.WriteFile(pathToConfig, myConfig, 0644)
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
