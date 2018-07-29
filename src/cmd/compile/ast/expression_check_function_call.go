package ast

import (
	"fmt"
)

func (e *Expression) checkFunctionCallExpression(block *Block, errs *[]error) []*Type {
	call := e.Data.(*ExpressionFunctionCall)
	on, es := call.Expression.checkSingleValueContextExpression(block)
	if esNotEmpty(es) {
		*errs = append(*errs, es...)
	}
	if on == nil {
		checkRightValuesValid(checkExpressions(block, call.Args, errs), errs)
		return nil
	}
	if on.Type == VariableTypeClass { // cast type
		typeConversion := &ExpressionTypeConversion{}
		typeConversion.Type = &Type{}
		typeConversion.Type.Type = VariableTypeObject
		typeConversion.Type.Class = on.Class
		typeConversion.Type.Pos = e.Pos
		if len(call.Args) != 1 {
			*errs = append(*errs, fmt.Errorf("%s cast type expect 1 argument",
				errMsgPrefix(e.Pos)))
			return nil
		}
		e.Type = ExpressionTypeCheckCast
		typeConversion.Expression = call.Args[0]
		e.Data = typeConversion
		ret := e.checkTypeConversionExpression(block, errs)
		if ret == nil {
			return nil
		}
		return []*Type{ret}
	}
	if on.Type == VariableTypeTypeAlias {
		typeConversion := &ExpressionTypeConversion{}
		typeConversion.Type = on.AliasType
		if len(call.Args) != 1 {
			*errs = append(*errs, fmt.Errorf("%s cast type expect 1 argument",
				errMsgPrefix(e.Pos)))
			return nil
		}
		e.Type = ExpressionTypeCheckCast
		typeConversion.Expression = call.Args[0]
		e.Data = typeConversion
		ret := e.checkTypeConversionExpression(block, errs)
		if ret == nil {
			return nil
		}
		return []*Type{ret}
	}
	if on.Type != VariableTypeFunction {
		*errs = append(*errs, fmt.Errorf("%s '%s' is not a function , but '%s'",
			errMsgPrefix(e.Pos),
			call.Expression.OpName(), on.TypeString()))
		return nil
	}
	if call.Expression.Type == ExpressionTypeFunctionLiteral {
		on.Function = call.Expression.Data.(*Function)
		/*
			fn() {
			}()
			no name function is statement too
		*/
		call.Expression.IsStatementExpression = true
	}
	if on.Function != nil {
		call.Function = on.Function
		if on.Function.IsBuildIn {
			return e.checkBuildInFunctionCall(block, errs, on.Function, call)
		} else {
			return e.checkFunctionCall(block, errs, on.Function, call)
		}
	}
	return e.checkFunctionPointerCall(block, errs, on.FunctionType, call)
}

func (e *Expression) checkFunctionPointerCall(block *Block, errs *[]error, ft *FunctionType, call *ExpressionFunctionCall) []*Type {
	callArgsTypes := checkRightValuesValid(checkExpressions(block, call.Args, errs), errs)
	ret := ft.getReturnTypes(e.Pos)
	var es []error
	_, call.VArgs, es = ft.fitCallArgs(e.Pos, &call.Args, callArgsTypes, nil)
	if esNotEmpty(es) {
		*errs = append(*errs, es...)
	}
	return ret
}

func (e *Expression) checkFunctionCall(block *Block, errs *[]error, f *Function, call *ExpressionFunctionCall) []*Type {
	callArgsTypes := checkRightValuesValid(checkExpressions(block, call.Args, errs), errs)
	var tf *Function
	if f.TemplateFunction != nil {
		length := len(*errs)
		//rewrite
		tf = e.checkTemplateFunctionCall(block, errs, callArgsTypes, f)
		if len(*errs) != length { // if no
			return nil
		}
	} else { // not template function
		if len(call.ParameterTypes) > 0 {
			*errs = append(*errs, fmt.Errorf("%s function is not a template function",
				errMsgPrefix(e.Pos)))
		}
	}
	var ret []*Type
	{
		f := f // override f
		if f.TemplateFunction != nil {
			f = tf
		}
		ret = f.Type.getReturnTypes(e.Pos)
		var es []error
		_, call.VArgs, es = f.Type.fitCallArgs(e.Pos, &call.Args, callArgsTypes, f)
		if esNotEmpty(es) {
			*errs = append(*errs, es...)
		}
	}
	return ret
}

func (e *Expression) checkTemplateFunctionCall(block *Block, errs *[]error,
	argTypes []*Type, f *Function) (ret *Function) {
	call := e.Data.(*ExpressionFunctionCall)
	parameterTypes := make(map[string]*Type)
	for k, v := range f.Type.ParameterList {
		if v == nil || v.Type == nil || len(v.Type.getParameterType()) == 0 {
			continue
		}
		if k >= len(argTypes) || argTypes[k] == nil {
			*errs = append(*errs, fmt.Errorf("%s missing typed parameter,index at %d",
				errMsgPrefix(e.Pos), k))
			return
		}
		if err := v.Type.canBeBindWithType(parameterTypes, argTypes[k]); err != nil {
			*errs = append(*errs, fmt.Errorf("%s %v",
				errMsgPrefix(argTypes[k].Pos), err))
			return
		}
	}
	tps := call.ParameterTypes
	for k, v := range f.Type.ReturnList {
		if v == nil || v.Type == nil || len(v.Type.getParameterType()) == 0 {
			continue
		}
		if len(tps) == 0 || tps[0] == nil {
			//trying already have
			if err := v.Type.canBeBindWithParameterTypes(parameterTypes); err == nil {
				//very good no error
				continue
			}
			*errs = append(*errs, fmt.Errorf("%s missing typed return value,index at %d",
				errMsgPrefix(e.Pos), k))
			return
		}
		if err := v.Type.canBeBindWithType(parameterTypes, tps[0]); err != nil {
			*errs = append(*errs, fmt.Errorf("%s %v",
				errMsgPrefix(tps[0].Pos), err))
			return nil
		}
		tps = tps[1:]
	}
	call.TemplateFunctionCallPair = f.TemplateFunction.insert(parameterTypes)
	if call.TemplateFunctionCallPair.Function == nil { // not called before,make the binds
		cloneFunction, es := f.clone()
		if esNotEmpty(es) {
			*errs = append(*errs, es...)
			return nil
		}
		cloneFunction.TemplateFunction = nil
		call.TemplateFunctionCallPair.Function = cloneFunction
		cloneFunction.parameterTypes = parameterTypes
		for _, v := range cloneFunction.Type.ParameterList {
			if len(v.Type.getParameterType()) > 0 {
				v.Type.bindWithParameterTypes(parameterTypes)
			}
		}
		for _, v := range cloneFunction.Type.ReturnList {
			if len(v.Type.getParameterType()) > 0 {
				v.Type.bindWithParameterTypes(parameterTypes)
			}
		}
		//check this function
		cloneFunction.Block.inherit(&PackageBeenCompile.Block)
		if cloneFunction.Block.Functions == nil {
			cloneFunction.Block.Functions = make(map[string]*Function)
		}
		cloneFunction.Block.Functions[cloneFunction.Name] = cloneFunction
		cloneFunction.Block.InheritedAttribute.Function = cloneFunction
		cloneFunction.checkParametersAndReturns(errs)
		cloneFunction.checkBlock(errs)
	}
	ret = call.TemplateFunctionCallPair.Function
	// when all ok ,ret is not a template function any more
	if len(tps) > 0 {
		*errs = append(*errs, fmt.Errorf("%s to many parameter type  to call template function",
			errMsgPrefix(e.Pos)))
	}
	return ret
}
