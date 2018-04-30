package ast

import (
	"fmt"
)

func errMsgPrefix(pos *Pos) string {
	return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.StartLine, pos.StartColumn)
}

//func ErrMsgPrefix(pos *Pos) string {
//	return errMsgPrefix(pos)
//}
func errsNotEmpty(es []error) bool {
	return es != nil && len(es) > 0
}

func checkEnum(enums []*Enum) []error {
	ret := make([]error, 0)
	for _, v := range enums {
		if len(v.Enums) == 0 {
			continue
		}
		err := v.check()
		if err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}

func oneAnyTypeParameterChecker(ft *Function, e *ExpressionFunctionCall,
	block *Block, errs *[]error, args []*VariableType, pos *Pos) {
	if len(args) != 1 {
		*errs = append(*errs, fmt.Errorf("%s only expect one argument", errMsgPrefix(pos)))
	}
}

func devisionByZeroErr(pos *Pos) error {
	return fmt.Errorf("%s division by zero", errMsgPrefix(pos))
}

func checkExpressions(block *Block, es []*Expression, errs *[]error) []*VariableType {
	ret := []*VariableType{}
	for _, v := range es {
		ts, e := v.check(block)
		if errsNotEmpty(e) {
			*errs = append(*errs, e...)
		}
		if ts != nil {
			ret = append(ret, ts...)
		} else {
			ret = append(ret, nil)
		}
	}
	return ret
}

func mkVoidType(pos *Pos) *VariableType {
	t := &VariableType{}
	t.Typ = VARIABLE_TYPE_VOID // means no return;
	t.Pos = pos
	return t
}

func checkRightValuesValid(ts []*VariableType, errs *[]error) (ret []*VariableType) {
	ret = []*VariableType{}
	for _, v := range ts {
		if v == nil { // avoid null pointer
			continue
		}
		if !v.RightValueValid() {
			*errs = append(*errs, fmt.Errorf("%s '%s' cannot used as right value",
				errMsgPrefix(v.Pos), v.TypeString()))
			continue
		}
		ret = append(ret, v)
	}
	return ret
}

/*
	when access from global,should check if access from package
*/
func shouldAccessFromImports(name string, from *Pos, have *Pos) (*Import, bool) {
	if have == nil { // incase buildin types
		return nil, false
	}
	// different file
	if from.Filename != have.Filename {
		i := PackageBeenCompile.getImport(from.Filename, name)
		return i, i != nil
	}
	i := PackageBeenCompile.getImport(from.Filename, name)
	if i == nil {
		return nil, false
	}
	return i, have.StartLine < from.StartLine
}

func msNotMatchError(pos *Pos, name string, ms []*ClassMethod, want []*VariableType) error {
	errmsg := fmt.Sprintf("%s method named '%s' have no suitable match:\n",
		errMsgPrefix(pos), name)
	errmsg += "\twant " + ms[0].Func.badParameterMsg(name, want) + "\n"
	for _, m := range ms {
		errmsg += "\thave " + m.Func.readableMsg(name) + "\n"
	}
	return fmt.Errorf(errmsg)
}

func searchBuildIns(name string) interface{} {
	var t interface{}
	var ok bool
	t, ok = buildinFunctionsMap[name]
	if ok {
		return t
	}
	if lucyLangBuildinPackage != nil {
		t, ok = lucyLangBuildinPackage.Block.Types[name]
		if ok {
			return t
		}
		t, ok = lucyLangBuildinPackage.Block.Classes[name]
		if ok {
			return t
		}
		t, ok = lucyLangBuildinPackage.Block.Vars[name]
		if ok {
			return t
		}
		t, ok = lucyLangBuildinPackage.Block.Consts[name]
		if ok {
			return t
		}
		t, ok = lucyLangBuildinPackage.Block.Enums[name]
		if ok {
			return t
		}
	}
	return nil
}
