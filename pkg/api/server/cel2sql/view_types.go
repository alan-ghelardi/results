package cel2sql

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FieldType string

const (
	FieldTypeString    FieldType = "string"
	FieldTypeTimestamp FieldType = "timestamp"
	FieldTypeObject    FieldType = "object"
	FieldTypeJSON      FieldType = "json"
)

var (
	CELTypeTimestamp = cel.ObjectType("google.protobuf.Timestamp")
)

type Field struct {
	CELType    *cel.Type
	SQL        string
	ObjectType any
}

type Constant struct {
	StringVal *string
	Int32Val  *int32
}

type View struct {
	TableName string
	Fields    map[string]Field
	Constants map[string]Constant
}

func (c *Constant) protoType() *exprpb.Type {
	if c.StringVal != nil {
		return decls.String
	}
	if c.Int32Val != nil {
		return decls.Int
	}
	return decls.Dyn
}

func (c *Constant) protoConstant() *exprpb.Constant {
	if c.StringVal != nil {
		return &exprpb.Constant{
			ConstantKind: &exprpb.Constant_StringValue{StringValue: *c.StringVal},
		}
	}

	if c.Int32Val != nil {
		return &exprpb.Constant{
			ConstantKind: &exprpb.Constant_Int64Value{Int64Value: int64(*c.Int32Val)},
		}
	}

	return nil
}

func (v *View) GetEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Declarations(v.celConstants()...),
		cel.Types(v.celTypes()...),
		cel.Declarations(v.celVariables()...),
	)
}

func (v *View) celConstants() []*exprpb.Decl {
	constants := make([]*exprpb.Decl, 0, len(v.Constants))
	for name, value := range v.Constants {
		constants = append(constants, decls.NewConst(name, value.protoType(), value.protoConstant()))
	}
	return constants
}

func (v *View) celTypes() []any {
	types := []any{&timestamppb.Timestamp{}}
	for _, field := range v.Fields {
		if field.ObjectType != nil {
			types = append(types, field.ObjectType)
		}
	}
	return types
}

func (v *View) celVariables() []*exprpb.Decl {
	vars := []*exprpb.Decl{}
	for name, field := range v.Fields {
		exprType, err := cel.TypeToExprType(field.CELType)
		if err != nil {
			panic("unexpected field type in view")
		}
		vars = append(vars, decls.NewVar(name, exprType))
	}
	return vars
}
