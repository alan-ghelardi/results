package view

import (
	"github.com/google/cel-go/cel"
	"github.com/tektoncd/results/pkg/api/server/cel2sql"
	"gorm.io/gorm"
)

func NewRecordsView(db *gorm.DB) (*cel2sql.View, error) {
	view := &cel2sql.View{
		TableName: db.Statement.Table,
		Constants: map[string]cel2sql.Constant{},
		Fields: map[string]cel2sql.Field{
			"parent": {
				CELType: cel.StringType,
				SQL:     `{{.Table}}.parent`,
			},
			"result_name": {
				CELType: cel.StringType,
				SQL:     `{{.Table}}.result_name`,
			},
			"name": {
				CELType: cel.StringType,
				SQL:     `{{.Table}}.name`,
			},
			"data_type": {
				CELType: cel.StringType,
				SQL:     `{{.Table}}.type`,
			},
			"data": {
				CELType: cel.AnyType,
				SQL:     `{{.Table}}.data`,
			},
		},
	}
	for typeName, value := range typeConstants {
		view.Constants[typeName] = value
	}
	return view, nil
}
