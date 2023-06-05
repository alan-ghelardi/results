package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/results/pkg/api/server/cel2sql"
	"github.com/tektoncd/results/pkg/api/server/test"
)

func TestConvertRecordExpressions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "simple expression",
			in:   `name == "foo"`,
			want: "results.name = 'foo'",
		},
		{
			name: "select expression",
			in:   `data.metadata.namespace == "default"`,
			want: "(results.data->'metadata'->>'namespace') = 'default'",
		},
		{
			name: "type coercion with a dyn expression in the left hand side",
			in:   `data.status.completionTime > timestamp("2022/10/30T21:45:00.000Z")`,
			want: "(results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE > '2022/10/30T21:45:00.000Z'::TIMESTAMP WITH TIME ZONE",
		},
		{
			name: "type coercion with a dyn expression in the right hand side",
			in:   `timestamp("2022/10/30T21:45:00.000Z") < data.status.completionTime`,
			want: "'2022/10/30T21:45:00.000Z'::TIMESTAMP WITH TIME ZONE < (results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE",
		},
		{
			name: "in operator",
			in:   `data.metadata.namespace in ["foo", "bar"]`,
			want: "(results.data->'metadata'->>'namespace') IN ('foo', 'bar')",
		},
		{
			name: "index operator",
			in:   `data.metadata.labels["foo"] == "bar"`,
			want: "(results.data->'metadata'->'labels'->>'foo') = 'bar'",
		},
		{
			name: "concatenate strings",
			in:   `name + "bar" == "foobar"`,
			want: "CONCAT(results.name, 'bar') = 'foobar'",
		},
		{
			name: "multiple concatenate strings",
			in:   `name + "bar" + "baz" == "foobarbaz"`,
			want: "CONCAT(results.name, 'bar', 'baz') = 'foobarbaz'",
		},
		{
			name: "contains string function",
			in:   `data.metadata.name.contains("foo")`,
			want: "POSITION('foo' IN (results.data->'metadata'->>'name')) <> 0",
		},
		{
			name: "endsWith string function",
			in:   `data.metadata.name.endsWith("bar")`,
			want: "(results.data->'metadata'->>'name') LIKE '%' || 'bar'",
		},
		{
			name: "getDate function",
			in:   `data.status.completionTime.getDate() == 2`,
			want: "EXTRACT(DAY FROM (results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE) = 2",
		},
		{
			name: "getDayOfMonth function",
			in:   `data.status.completionTime.getDayOfMonth() == 2`,
			want: "(EXTRACT(DAY FROM (results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE) - 1) = 2",
		},
		{
			name: "getDayOfWeek function",
			in:   `data.status.completionTime.getDayOfWeek() > 0`,
			want: "EXTRACT(DOW FROM (results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE) > 0",
		},
		{
			name: "getDayOfYear function",
			in:   `data.status.completionTime.getDayOfYear() > 15`,
			want: "(EXTRACT(DOY FROM (results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE) - 1) > 15",
		},
		{
			name: "getFullYear function",
			in:   `data.status.completionTime.getFullYear() >= 2022`,
			want: "EXTRACT(YEAR FROM (results.data->'status'->>'completionTime')::TIMESTAMP WITH TIME ZONE) >= 2022",
		},
		{
			name: "matches function",
			in:   `data.metadata.name.matches("^foo.*$")`,
			want: "(results.data->'metadata'->>'name') ~ '^foo.*$'",
		},
		{
			name: "startsWith string function",
			in:   `data.metadata.name.startsWith("bar")`,
			want: "(results.data->'metadata'->>'name') LIKE 'bar' || '%'",
		},
		{
			name: "data_type field",
			in:   `data_type == PIPELINE_RUN`,
			want: "results.type = 'tekton.dev/v1beta1.PipelineRun'",
		},
		{
			name: "index operator with numeric argument in JSON arrays",
			in:   `data_type == "tekton.dev/v1beta1.TaskRun" && data.status.conditions[0].status == "True"`,
			want: "results.type = 'tekton.dev/v1beta1.TaskRun' AND (results.data->'status'->'conditions'->0->>'status') = 'True'",
		},
		{
			name: "index operator as first operation in JSON object",
			in:   `data_type == "tekton.dev/v1beta1.TaskRun" && data["status"].conditions[0].status == "True"`,
			want: "results.type = 'tekton.dev/v1beta1.TaskRun' AND (results.data->'status'->'conditions'->0->>'status') = 'True'",
		},
		{
			name: "index operator with string argument in JSON object",
			in:   `data_type == "tekton.dev/v1beta1.TaskRun" && data.status["conditions"][0].status == "True"`,
			want: "results.type = 'tekton.dev/v1beta1.TaskRun' AND (results.data->'status'->'conditions'->0->>'status') = 'True'",
		},
	}

	db := test.NewDB(t)
	view, err := NewRecordsView(db.Table("results"))
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := cel2sql.ConvertView(view, test.in)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("want: %+v\n", test.want)
			t.Logf("got:  %+v\n", got)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
