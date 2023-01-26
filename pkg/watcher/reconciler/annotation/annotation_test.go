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

package annotation

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestPatch(t *testing.T) {
	const (
		fakeResult = "foo/results/bar"
		fakeRecord = "foo/results/bar/records/baz"
	)

	tests := []struct {
		name string
		in   func() metav1.Object
		want mergePatch
	}{{
		name: "create a patch containing only the result and record identifiers since the object is a PipelineRun",
		in: func() metav1.Object {
			return &pipelinev1beta1.PipelineRun{}
		},
		want: mergePatch{
			Metadata: metadata{
				Annotations: map[string]string{
					Result: fakeResult,
					Record: fakeRecord,
				},
			},
		},
	},
		{
			name: "create a patch containing only the result and record identifiers since the TaskRun isn't owned by a PipelineRun",
			in: func() metav1.Object {
				return &pipelinev1beta1.TaskRun{}
			},
			want: mergePatch{
				Metadata: metadata{
					Annotations: map[string]string{
						Result: fakeResult,
						Record: fakeRecord,
					},
				},
			},
		},
		{
			name: "create a patch containing only the result and record identifiers since the TaskRun isn't done yet",
			in: func() metav1.Object {
				return &pipelinev1beta1.TaskRun{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{{
							UID: types.UID("UID"),
						},
						},
					},
				}
			},
			want: mergePatch{
				Metadata: metadata{
					Annotations: map[string]string{
						Result: fakeResult,
						Record: fakeRecord,
					},
				},
			},
		},
		{
			name: "mark the TaskRun as ready for deletion since it's owned by a PipelineRun and is done",
			in: func() metav1.Object {
				taskRun := &pipelinev1beta1.TaskRun{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{{
							UID: types.UID("UID"),
						},
						},
					},
				}
				taskRun.Status.MarkResourceFailed(pipelinev1beta1.TaskRunReasonFailed, errors.New("Failed"))
				return taskRun
			},
			want: mergePatch{
				Metadata: metadata{
					Annotations: map[string]string{
						Result:                fakeResult,
						Record:                fakeRecord,
						ChildReadyForDeletion: "true",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := test.in()
			patch, err := Patch(object, fakeResult, fakeRecord)
			if err != nil {
				t.Fatal(err)
			}

			var got mergePatch
			if err := json.Unmarshal(patch, &got); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsPatched(t *testing.T) {
	const (
		fakeResult = "foo/results/bar"
		fakeRecord = "foo/results/bar/records/baz"
	)

	tests := []struct {
		name string
		in   func() metav1.Object
		want bool
	}{{
		name: "result and record identifiers are missing in the PipelineRun",
		in: func() metav1.Object {
			return &pipelinev1beta1.PipelineRun{}
		},
		want: false,
	},
		{
			name: "the record identifier is missing in the PipelineRun",
			in: func() metav1.Object {
				return &pipelinev1beta1.PipelineRun{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							Result: fakeResult,
						},
					},
				}
			},
			want: false,
		},
		{
			name: "the PipelineRun contains all relevant annotations",
			in: func() metav1.Object {
				return &pipelinev1beta1.PipelineRun{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							Result: fakeResult,
							Record: fakeRecord,
						},
					},
				}
			},
			want: true,
		},
		{
			name: "the TaskRun contains all relevant annotations",
			in: func() metav1.Object {
				taskRun := &pipelinev1beta1.TaskRun{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							Result:                fakeResult,
							Record:                fakeRecord,
							ChildReadyForDeletion: "true",
						},
						OwnerReferences: []metav1.OwnerReference{{
							UID: types.UID("UID"),
						},
						},
					},
				}
				taskRun.Status.MarkResourceFailed(pipelinev1beta1.TaskRunReasonFailed, errors.New("Failed"))
				return taskRun
			},
			want: true,
		},
		{
			name: "the TaskRun should be marked as ready to be deleted",
			in: func() metav1.Object {
				taskRun := &pipelinev1beta1.TaskRun{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							Result: fakeResult,
							Record: fakeRecord,
						},
						OwnerReferences: []metav1.OwnerReference{{
							UID: types.UID("UID"),
						},
						},
					},
				}
				taskRun.Status.MarkResourceFailed(pipelinev1beta1.TaskRunReasonFailed, errors.New("Failed"))
				return taskRun
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := test.in()
			got := IsPatched(object)
			if test.want != got {
				t.Errorf("Want %t, but got %t", test.want, got)
			}
		})
	}
}
