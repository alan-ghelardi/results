package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/results/pkg/api/server/cel2sql"
	"github.com/tektoncd/results/pkg/api/server/test"
)

func TestConvertResultExpressions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "Result.Parent field",
			in:   `parent.endsWith("bar")`,
			want: "records.parent LIKE '%' || 'bar'",
		},
		{
			name: "Result.Uid field",
			in:   `uid == "foo"`,
			want: "records.id = 'foo'",
		},
		{
			name: "Result.Annotations field",
			in:   `annotations["repo"] == "tektoncd/results"`,
			want: `records.annotations @> '{"repo":"tektoncd/results"}'::jsonb`,
		},
		{
			name: "Result.Annotations field",
			in:   `"tektoncd/results" == annotations["repo"]`,
			want: `records.annotations @> '{"repo":"tektoncd/results"}'::jsonb`,
		},
		{
			name: "other operators involving the Result.Annotations field",
			in:   `annotations["repo"].startsWith("tektoncd")`,
			want: "records.annotations->>'repo' LIKE 'tektoncd' || '%'",
		},
		{
			name: "Result.CreateTime field",
			in:   `create_time > timestamp("2022/10/30T21:45:00.000Z")`,
			want: "records.created_time > '2022/10/30T21:45:00.000Z'::TIMESTAMP WITH TIME ZONE",
		},
		{
			name: "Result.UpdateTime field",
			in:   `update_time > timestamp("2022/10/30T21:45:00.000Z")`,
			want: "records.updated_time > '2022/10/30T21:45:00.000Z'::TIMESTAMP WITH TIME ZONE",
		},
		{
			name: "Result.Summary.Record field",
			in:   `summary.record == "foo/results/bar/records/baz"`,
			want: "records.recordsummary_record = 'foo/results/bar/records/baz'",
		},
		{
			name: "Result.Summary.StartTime field",
			in:   `summary.start_time > timestamp("2022/10/30T21:45:00.000Z")`,
			want: "records.recordsummary_start_time > '2022/10/30T21:45:00.000Z'::TIMESTAMP WITH TIME ZONE",
		},
		{
			name: "comparison with the PIPELINE_RUN const value",
			in:   `summary.type == PIPELINE_RUN`,
			want: "records.recordsummary_type = 'tekton.dev/v1beta1.PipelineRun'",
		},
		{
			name: "comparison with the TASK_RUN const value",
			in:   `summary.type == TASK_RUN`,
			want: "records.recordsummary_type = 'tekton.dev/v1beta1.TaskRun'",
		},
		{
			name: "RecordSummary_Status constants",
			in:   `summary.status == CANCELLED || summary.status == TIMEOUT`,
			want: "records.recordsummary_status = 4 OR records.recordsummary_status = 3",
		},
		{
			name: "Result.Summary.Annotations",
			in:   `summary.annotations["branch"] == "main"`,
			want: `records.recordsummary_annotations @> '{"branch":"main"}'::jsonb`,
		},
		{
			name: "Result.Summary.Annotations",
			in:   `"main" == summary.annotations["branch"]`,
			want: `records.recordsummary_annotations @> '{"branch":"main"}'::jsonb`,
		},
		{
			name: "more complex expression",
			in:   `summary.annotations["actor"] == "john-doe" && summary.annotations["branch"] == "feat/amazing" && summary.status == SUCCESS`,
			want: `records.recordsummary_annotations @> '{"actor":"john-doe"}'::jsonb AND records.recordsummary_annotations @> '{"branch":"feat/amazing"}'::jsonb AND records.recordsummary_status = 1`,
		},
	}

	db := test.NewDB(t)
	env, err := NewResultsView(db.Table("records"))
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := cel2sql.ConvertView(env, test.in)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
