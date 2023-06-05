package view

import (
	"github.com/google/cel-go/cel"
	"github.com/tektoncd/results/pkg/api/server/cel2sql"
	resultspb "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"gorm.io/gorm"
)

var (
	typePipelineRun = "tekton.dev/v1beta1.PipelineRun"
	typeTaskRun     = "tekton.dev/v1beta1.TaskRun"

	typeConstants = map[string]cel2sql.Constant{
		"PIPELINE_RUN": {
			StringVal: &typePipelineRun,
		},
		"TASK_RUN": {
			StringVal: &typeTaskRun,
		},
	}
)

func NewResultsView(db *gorm.DB) (*cel2sql.View, error) {
	view := &cel2sql.View{
		TableName: db.Statement.Table,
		Constants: map[string]cel2sql.Constant{},
		Fields: map[string]cel2sql.Field{
			"parent": {
				CELType: cel.StringType,
				SQL:     `{{.Table}}.parent`,
			},
			"uid": {
				CELType: cel.StringType,
				SQL:     `{{.Table}}.id`,
			},
			"create_time": {
				CELType: cel2sql.CELTypeTimestamp,
				SQL:     `{{.Table}}.created_time`,
			},
			"update_time": {
				CELType: cel2sql.CELTypeTimestamp,
				SQL:     `{{.Table}}.updated_time`,
			},
			"annotations": {
				CELType: cel.MapType(cel.StringType, cel.StringType),
				SQL:     `{{.Table}}.annotations`,
			},
			"summary": {
				CELType:    cel.ObjectType("tekton.results.v1alpha2.RecordSummary"),
				ObjectType: &resultspb.RecordSummary{},
				SQL:        `{{.Table}}.recordsummary_{{.Field}}`,
			},
		},
	}
	for typeName, value := range typeConstants {
		view.Constants[typeName] = value
	}
	for name, value := range resultspb.RecordSummary_Status_value {
		v := value
		view.Constants[name] = cel2sql.Constant{
			Int32Val: &v,
		}
	}
	return view, nil
}
