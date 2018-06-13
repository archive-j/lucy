package ast

import (
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/common"
)

func init() {
	registerBuildinFunctions()
}

func registerBuildinFunctions() {
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_PRINT] = &Function{
		buildInFunctionChecker: func(ft *Function, e *ExpressionFunctionCall, block *Block, errs *[]error, args []*VariableType, pos *Pos) {
			if len(e.TypedParameters) > 0 {
				*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
					errMsgPrefix(pos)))
			}
			meta := &BuildInFunctionPrintfMeta{}
			e.BuildInFunctionMeta = meta
			if len(args) == 0 || args[0] == nil {
				return // not error
			}
			if args[0].Typ == VARIABLE_TYPE_OBJECT {
				have, _ := args[0].Class.haveSuper("java/io/PrintStream")
				if have {
					_, err := e.Args[0].mustBeOneValueContext(e.Args[0].Values)
					if err != nil {
						*errs = append(*errs, err)
					} else {
						meta.Stream = e.Args[0]
						e.Args = e.Args[1:]
					}
				}
			}
		},
		IsBuildIn: true,
		Name:      common.BUILD_IN_FUNCTION_PRINT,
	}
	catchBuildFunction := &Function{}
	catchBuildFunction.IsBuildIn = true
	catchBuildFunction.Name = common.BUILD_IN_FUNCTION_CATCH
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_CATCH] = catchBuildFunction
	{
		catchBuildFunction.Typ.ReturnList = make([]*VariableDefinition, 1)
		catchBuildFunction.Typ.ReturnList[0] = &VariableDefinition{}
		catchBuildFunction.Typ.ReturnList[0].Name = "retrunValue"
		catchBuildFunction.Typ.ReturnList[0].Typ = &VariableType{}
		catchBuildFunction.Typ.ReturnList[0].Typ.Typ = VARIABLE_TYPE_OBJECT
		catchBuildFunction.Typ.ReturnList[0].Typ.Class = &Class{}
		catchBuildFunction.Typ.ReturnList[0].Typ.Class.Name = DEFAULT_EXCEPTION_CLASS
		catchBuildFunction.Typ.ReturnList[0].Typ.Class.NotImportedYet = true
		//class is going to make value by checker
	}
	catchBuildFunction.buildInFunctionChecker = func(ft *Function, e *ExpressionFunctionCall, block *Block, errs *[]error, args []*VariableType, pos *Pos) {
		if len(e.TypedParameters) > 0 {
			*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
				errMsgPrefix(pos)))
		}
		if block.InheritedAttribute.Defer == nil ||
			block.InheritedAttribute.Defer.allowCatch == false {
			*errs = append(*errs, fmt.Errorf("%s buildin function '%s' only allow in defer block",
				errMsgPrefix(pos), common.BUILD_IN_FUNCTION_CATCH))
			return
		}
		if len(args) > 1 {
			*errs = append(*errs, fmt.Errorf("%s build function '%s' expect at most 1 argument",
				errMsgPrefix(pos), common.BUILD_IN_FUNCTION_CATCH))
			return
		}
		if len(args) == 0 {
			// make default exception class
			// load java/lang/Exception this is default exception level to catch
			if block.InheritedAttribute.Defer.ExceptionClass == nil {
				c, err := ResourceLoader.LoadName(DEFAULT_EXCEPTION_CLASS)
				if err != nil {
					*errs = append(*errs, fmt.Errorf("%s load exception class failed,err:%v",
						errMsgPrefix(pos), err))
					return
				}
				ft.Typ.ReturnList[0].Typ.Class = c.(*Class)
				err = block.InheritedAttribute.Defer.registerExceptionClass(c.(*Class))
				if err != nil {
					*errs = append(*errs, fmt.Errorf("%s %v", errMsgPrefix(pos), err))
				}
			} else {
				ft.Typ.ReturnList[0].Typ.Class = block.InheritedAttribute.Defer.ExceptionClass

			}
			return
		}
		if args[0] == nil {
			return
		}
		if args[0].Typ != VARIABLE_TYPE_OBJECT {
			*errs = append(*errs, fmt.Errorf("%s build function '%s' expect a object ref argument",
				errMsgPrefix(pos), common.BUILD_IN_FUNCTION_CATCH))
			return
		}
		if has, _ := args[0].Class.haveSuper(JAVA_THROWABLE_CLASS); has == false {
			*errs = append(*errs, fmt.Errorf("%s '%s' does not have super-class '%s'",
				errMsgPrefix(pos), args[0].Class.Name, JAVA_THROWABLE_CLASS))
			return
		}
		err := block.InheritedAttribute.Defer.registerExceptionClass(args[0].Class)
		if err != nil {
			*errs = append(*errs, fmt.Errorf("%s %v", errMsgPrefix(pos), err))
		}
	}
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_PANIC] = &Function{
		buildInFunctionChecker: func(ft *Function, e *ExpressionFunctionCall,
			block *Block, errs *[]error, args []*VariableType, pos *Pos) {
			if len(e.TypedParameters) > 0 {
				*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
					errMsgPrefix(pos)))
			}
			if len(args) != 1 {
				*errs = append(*errs, fmt.Errorf("%s buildin function 'panic' expect one argument",
					errMsgPrefix(pos)))
				return
			}
			if len(args) == 0 || args[0] == nil {
				return
			}
			if args[0].Typ != VARIABLE_TYPE_OBJECT {
				*errs = append(*errs, fmt.Errorf("%s cannot use '%s' for panic",
					errMsgPrefix(pos), args[0].TypeString()))
				return
			}
			if have, _ := args[0].Class.haveSuper(JAVA_THROWABLE_CLASS); have == false {
				*errs = append(*errs, fmt.Errorf("%s cannot use '%s' for panic",
					errMsgPrefix(pos), args[0].TypeString()))
				return
			}
		},
		IsBuildIn: true,
		Name:      common.BUILD_IN_FUNCTION_PANIC,
	}
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_MONITORENTER] = &Function{
		buildInFunctionChecker: monitorChecker,
		IsBuildIn:              true,
		Name:                   common.BUILD_IN_FUNCTION_MONITORENTER,
	}
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_MONITOREXIT] = &Function{
		buildInFunctionChecker: monitorChecker,
		IsBuildIn:              true,
		Name:                   common.BUILD_IN_FUNCTION_MONITOREXIT,
	}
	// len
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_LEN] = &Function{
		buildInFunctionChecker: func(f *Function, e *ExpressionFunctionCall, block *Block, errs *[]error, args []*VariableType, pos *Pos) {
			if len(e.TypedParameters) > 0 {
				*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
					errMsgPrefix(pos)))
			}
			if len(args) != 1 {
				*errs = append(*errs, fmt.Errorf("%s expect one argument", errMsgPrefix(pos)))
				return
			}
			if args[0] == nil {
				return
			}
			if args[0].Typ != VARIABLE_TYPE_ARRAY && args[0].Typ != VARIABLE_TYPE_JAVA_ARRAY &&
				args[0].Typ != VARIABLE_TYPE_MAP && args[0].Typ != VARIABLE_TYPE_STRING {
				*errs = append(*errs, fmt.Errorf("%s len expect 'array' or 'map' or 'string' argument",
					errMsgPrefix(pos)))
				return
			}
		},
		IsBuildIn: true,
		Name:      common.BUILD_IN_FUNCTION_LEN,
	}
	lenFunction := buildInFunctionsMap[common.BUILD_IN_FUNCTION_LEN]
	lenFunction.Typ.ReturnList = make(ReturnList, 1)
	lenFunction.Typ.ReturnList[0] = &VariableDefinition{}
	lenFunction.Typ.ReturnList[0].Typ = &VariableType{}
	lenFunction.Typ.ReturnList[0].Typ.Typ = VARIABLE_TYPE_INT
	// sprintf
	sprintfBuildFunction := &Function{}
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_SPRINTF] = sprintfBuildFunction
	sprintfBuildFunction.Name = common.BUILD_IN_FUNCTION_SPRINTF
	sprintfBuildFunction.IsBuildIn = true
	{
		sprintfBuildFunction.Typ.ReturnList = make([]*VariableDefinition, 1)
		sprintfBuildFunction.Typ.ReturnList[0] = &VariableDefinition{}
		sprintfBuildFunction.Typ.ReturnList[0].Name = "retrunValue"
		sprintfBuildFunction.Typ.ReturnList[0].Typ = &VariableType{}
		sprintfBuildFunction.Typ.ReturnList[0].Typ.Typ = VARIABLE_TYPE_STRING
	}
	sprintfBuildFunction.buildInFunctionChecker = func(ft *Function, e *ExpressionFunctionCall, block *Block, errs *[]error,
		args []*VariableType, pos *Pos) {
		if len(e.TypedParameters) > 0 {
			*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
				errMsgPrefix(pos)))
		}
		if len(args) == 0 {
			err := fmt.Errorf("%s '%s' expect one argument at lease",
				errMsgPrefix(pos), common.BUILD_IN_FUNCTION_SPRINTF)
			*errs = append(*errs, err)
			return
		}
		if args[0] == nil {
			return
		}
		if args[0].Typ != VARIABLE_TYPE_STRING {
			err := fmt.Errorf("%s '%s' first argument must be string",
				errMsgPrefix(pos), common.BUILD_IN_FUNCTION_SPRINTF)
			*errs = append(*errs, err)
			return
		}
		_, err := e.Args[0].mustBeOneValueContext(e.Args[0].Values)
		if err != nil {
			*errs = append(*errs, err)
			return
		}
		meta := &BuildInFunctionSprintfMeta{}
		e.BuildInFunctionMeta = meta
		meta.Format = e.Args[0]
		meta.ArgsLength = len(args) - 1
		e.Args = e.Args[1:]
	}
	// printf
	buildInFunctionsMap[common.BUILD_IN_FUNCTION_PRINTF] = &Function{
		buildInFunctionChecker: func(ft *Function, e *ExpressionFunctionCall, block *Block, errs *[]error,
			args []*VariableType, pos *Pos) {
			if len(e.TypedParameters) > 0 {
				*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
					errMsgPrefix(pos)))
			}
			meta := &BuildInFunctionPrintfMeta{}
			e.BuildInFunctionMeta = meta
			if len(args) == 0 {
				err := fmt.Errorf("%s '%s' expect one argument at least",
					errMsgPrefix(pos), common.BUILD_IN_FUNCTION_PRINTF)
				*errs = append(*errs, err)
				return
			}
			if args[0] == nil {
				return
			}
			if args[0].Typ == VARIABLE_TYPE_OBJECT {
				have, _ := args[0].Class.haveSuper("java/io/PrintStream")
				if have {
					_, err := e.Args[0].mustBeOneValueContext(e.Args[0].Values)
					if err != nil {
						*errs = append(*errs, err)
						return
					} else {
						meta.Stream = e.Args[0]
						e.Args = e.Args[1:]
						args = args[1:]
					}
				}
			}
			if len(args) == 0 {
				err := fmt.Errorf("%s missing format argument",
					errMsgPrefix(pos))
				*errs = append(*errs, err)
				return
			}
			if args[0] == nil {
				return
			}
			if args[0].Typ != VARIABLE_TYPE_STRING {
				err := fmt.Errorf("%s format must be string",
					errMsgPrefix(pos))
				*errs = append(*errs, err)
				return
			}
			_, err := e.Args[0].mustBeOneValueContext(e.Args[0].Values)
			if err != nil {
				*errs = append(*errs, err)
				return
			}
			meta.Format = e.Args[0]
			e.Args = e.Args[1:]
			meta.ArgsLength = len(args)
		},
		IsBuildIn: true,
		Name:      common.BUILD_IN_FUNCTION_PRINTF,
	}
}

func monitorChecker(f *Function, e *ExpressionFunctionCall, block *Block, errs *[]error,
	args []*VariableType, pos *Pos) {
	if len(e.TypedParameters) > 0 {
		*errs = append(*errs, fmt.Errorf("%s buildin function expect no typed parameter",
			errMsgPrefix(pos)))
	}
	if len(args) != 1 {
		*errs = append(*errs, fmt.Errorf("%s only expect one argument", errMsgPrefix(pos)))
		return
	}
	if args[0] == nil {
		return
	}
	if args[0].IsPointer() == false || args[0].Typ == VARIABLE_TYPE_STRING {
		*errs = append(*errs, fmt.Errorf("%s '%s' is not valid type to call",
			errMsgPrefix(pos), args[0].TypeString()))
		return
	}
}
