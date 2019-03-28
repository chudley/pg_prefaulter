// Copyright © 2019 Joyent, Inc.
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

package pg_test

import (
	"testing"

	"github.com/joyent/pg_prefaulter/pg"
	"github.com/kylelemons/godebug/pretty"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		version   string
		expected  string
	}{
		{
			version:  "10.1",
			expected: "10",
		},
		{
			version:  "9.6",
			expected: "9.6",
		},
		{
			version:  "11",
			expected: "11",
		},
		{
			version:  "123.1.1",
			expected: "123",
		},
		{
			version:  "8.2.1",
			expected: "8.2",
		},
		{
			version:  "9.2",
			expected: "9.2",
		},
		{
			version:  "9.6.2",
			expected: "9.6",
		},
	}

	for i, test := range tests {
		translations, err := pg.Translate(test.version)
		if err != nil {
			t.Fatalf("%d failed: %q", i, test.version)
		}

		if diff := pretty.Compare(translations.Major, test.expected); diff != "" {
			t.Fatalf("version diff: (-got +want)\n%s", diff)
		}

	}
}
