package jvm

import (
	"github.com/756445638/lucy/src/cmd/compile/ast"
	"github.com/756445638/lucy/src/cmd/compile/common"
	"github.com/756445638/lucy/src/cmd/compile/jvm/cg"
)

func (m *MakeExpression) mkBuildinFunctionCall(class *cg.ClassHighLevel, code *cg.AttributeCode, e *ast.Expression, context *Context) (maxstack uint16) {
	call := e.Data.(*ast.ExpressionFunctionCall)
	switch call.Func.Name {
	case common.BUILD_IN_FUNCTION_PRINT:
		return m.mkBuildinPrint(class, code, call, context)
	case common.BUILD_IN_FUNCTION_PANIC:
		return m.mkBuildinPanic(class, code, call, context)
	case common.BUILD_IN_FUNCTION_CATCH:
		return m.mkBuildinRecover(class, code, e, context)
	default:
		panic("unhandle buildin function:" + call.Func.Name)
	}
}
