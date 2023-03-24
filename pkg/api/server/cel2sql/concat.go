package cel2sql

import (
	"fmt"

	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// mayBeTranslatedIntoStringConcatExpression returns a boolean whether the
// expression is a string concatenation. If it is, returns both arguments.
func (i *interpreter) mayBeTranslatedIntoStringConcatExpression(expr *exprpb.Expr) (bool, *exprpb.Expr, *exprpb.Expr) {
	callExpr, ok := expr.ExprKind.(*exprpb.Expr_CallExpr)
	if !ok {
		return false, nil, nil
	}
	if function := callExpr.CallExpr.GetFunction(); !isAddOperator(function) {
		return false, nil, nil
	}
	arg1 := callExpr.CallExpr.Args[0]
	arg2 := callExpr.CallExpr.Args[1]
	if i.isString(arg1) || i.isString(arg2) {
		return true, arg1, arg2
	}
	return false, nil, nil
}

func (i interpreter) allStringConcatArgs(expr *exprpb.Expr) []*exprpb.Expr {
	args := []*exprpb.Expr{}
	switch node := expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		if isAddOperator(node.CallExpr.Function) {
			arg1 := node.CallExpr.Args[0]
			arg2 := node.CallExpr.Args[1]
			args = append(args, i.allStringConcatArgs(arg1)...)
			args = append(args, i.allStringConcatArgs(arg2)...)
		}
	default:
		args = append(args, expr)
	}
	return args
}

func (i *interpreter) translateIntoStringConcatExpression(expr *exprpb.Expr) error {
	args := i.allStringConcatArgs(expr)
	fmt.Fprintf(&i.query, "CONCAT(")
	for j, arg := range args {
		err := i.interpretExpr(arg)
		if err != nil {
			return err
		}
		if j != len(args)-1 {
			fmt.Fprintf(&i.query, ", ")
		}
	}
	fmt.Fprintf(&i.query, ")")
	return nil
}
