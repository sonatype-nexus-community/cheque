// Copyright 2020 Sonatype Inc.
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
package linker

import "testing"

func TestDoLink(t *testing.T) {
	args := []string{
		"-l/usr/path/lib.so",
		"-l",
		"/usr/path/different.so",
		"-L/usr/path/",
		"-L",
		"/usr/differentpath",
		"-o/usr/path/output",
		"-o",
		"/usr/differentoutput/output",
	}

  // FIXME: By default we are returned the number of vulnerable packages. We
  // might want to add a way to get the total number of packages as well.
  var expected = 0;
	countFromSesameStreet := DoLink(args)

	if countFromSesameStreet != expected {
		t.Errorf("Error: Expected %d but got %d", expected, countFromSesameStreet)
	}
}
