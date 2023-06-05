// Copyright 2023 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cel2sql

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func newTestView() *View {
	return &View{
		TableName: "testtable",
		Constants: map[string]Constant{},
		Fields:    map[string]Field{},
	}
}
func TestConversionErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{
			name: "non-boolean expression",
			in:   "parent",
			want: errors.New("expected boolean expression, but got string"),
		},
	}

	view := newTestView()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ConvertView(view, test.in)
			if err == nil {
				t.Fatalf("Want the %q error, but the interpreter returned the following result instead: %q", test.want.Error(), got)
			}

			if diff := cmp.Diff(test.want.Error(), err.Error()); diff != "" {
				t.Fatalf("Mismatch in the error returned by the Convert function (-want +got):\n%s", diff)
			}
		})
	}
}
