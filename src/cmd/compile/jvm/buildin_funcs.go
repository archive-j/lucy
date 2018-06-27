package jvm

import (
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/common"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

func (makeExpression *MakeExpression) mkBuildInFunctionCall(class *cg.ClassHighLevel, code *cg.AttributeCode,
	e *ast.Expression, context *Context, state *StackMapState) (maxStack uint16) {
	call := e.Data.(*ast.ExpressionFunctionCall)
	switch call.Function.Name {
	case common.BuildInFunctionPrint:
		return makeExpression.mkBuildInPrint(class, code, e, context, state)
	case common.BuildInFunctionPanic:
		return makeExpression.mkBuildInPanic(class, code, e, context, state)
	case common.BuildInFunctionCatch:
		return makeExpression.mkBuildInCatch(class, code, e, context)
	case common.BuildInFunctionMonitorEnter, common.BuildInFunctionMonitorExit:
		maxStack, _ = makeExpression.build(class, code, call.Args[0], context, state)
		if call.Function.Name == common.BuildInFunctionMonitorEnter {
			code.Codes[code.CodeLength] = cg.OP_monitorenter
		} else { // monitor enter on exit
			code.Codes[code.CodeLength] = cg.OP_monitorexit
		}
		code.CodeLength++
	case common.BuildInFunctionPrintf:
		return makeExpression.mkBuildInPrintf(class, code, e, context, state)
	case common.BuildInFunctionSprintf:
		return makeExpression.mkBuildInSprintf(class, code, e, context, state)
	case common.BuildInFunctionLen:
		return makeExpression.mkBuildInLen(class, code, e, context, state)
	default:
		panic("unKnow buildIn function:" + call.Function.Name)
	}
	return
}
