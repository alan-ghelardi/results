package cel2sql

import (
	"fmt"

	"gorm.io/gorm/schema"
)

// translateToJSONAccessors converts the provided field path to a Postgres JSON
// property selection directive. This allows us to yield appropriate SQL
// expressions to navigate through the record.data field, for instance.
func (i *interpreter) translateToJSONAccessors(fieldPath []any) {
	firstField := fieldPath[0]
	lastField := fieldPath[len(fieldPath)-1]

	fmt.Fprintf(&i.query, "(%s->", firstField)
	if len(fieldPath) > 2 {
		for _, field := range fieldPath[1 : len(fieldPath)-1] {
			fmt.Fprintf(&i.query, "%s->", jsonKey(field))
		}
	}
	fmt.Fprintf(&i.query, ">%s)", jsonKey(lastField))
}

func jsonKey(value any) string {
	if s, ok := value.(string); ok {
		return fmt.Sprintf("'%s'", s)
	}
	return fmt.Sprint(value)
}

// translateIntoRecordSummaryColum
func (i *interpreter) translateIntoRecordSummaryColum(fieldPath []any) {
	namer := &schema.NamingStrategy{}
	fmt.Fprintf(&i.query, "recordsummary_%s", namer.ColumnName("", fmt.Sprint(fieldPath[1])))
}
