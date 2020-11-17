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
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
)

func TestCreateEmptyObject(t *testing.T) {
	conf := setup()

	conf.CreateOrReadConfigFile()

	exists, err := os.Stat(conf.getIQConfig())

	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(exists.Name(), types.IQServerConfigFileName) {
		t.Errorf("File not created properly, expected %s but got %s", types.IQServerConfigFileName, exists.Name())
	}
}

func setup() *Config {
	logLady, _ := test.NewNullLogger()

	options := Options{}

	tempDir, err := ioutil.TempDir("fakefortest")
	options.Directory = tempDir

	conf := New(logLady)

	return conf
}
