package jvm

import (
	"bytes"
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
)

type Descript struct {
}

func (m *Descript) methodDescriptor(f *ast.Function) string {
	if f.Name == ast.MAIN_FUNCTION_NAME {
		return "([Ljava/lang/String;)V"
	}
	s := "("
	for _, v := range f.Typ.ParameterList {
		s += m.typeDescriptor(v.Typ)
	}
	s += ")"
	if f.NoReturnValue() {
		s += "V"
	} else if len(f.Typ.ReturnList) == 1 {
		s += m.typeDescriptor(f.Typ.ReturnList[0].Typ)
	} else {
		s += "[Ljava/lang/Object;" //always this type
	}
	return s
}

func (m *Descript) typeDescriptor(v *ast.VariableType) string {
	switch v.Typ {
	case ast.VARIABLE_TYPE_BOOL:
		return "Z"
	case ast.VARIABLE_TYPE_BYTE:
		return "B"
	case ast.VARIABLE_TYPE_SHORT:
		return "S"
	case ast.VARIABLE_TYPE_INT, ast.VARIABLE_TYPE_ENUM:
		return "I"
	case ast.VARIABLE_TYPE_LONG:
		return "J"
	case ast.VARIABLE_TYPE_FLOAT:
		return "F"
	case ast.VARIABLE_TYPE_DOUBLE:
		return "D"
	case ast.VARIABLE_TYPE_ARRAY:
		meta := ArrayMetas[v.ArrayType.Typ] // combination type
		return "L" + meta.className + ";"
	case ast.VARIABLE_TYPE_STRING:
		return "Ljava/lang/String;"
	case ast.VARIABLE_TYPE_VOID:
		return "V"
	case ast.VARIABLE_TYPE_OBJECT:
		return "L" + v.Class.Name + ";"
	case ast.VARIABLE_TYPE_MAP:
		return "L" + java_hashmap_class + ";"
	case ast.VARIABLE_TYPE_JAVA_ARRAY:
		return "[" + m.typeDescriptor(v.ArrayType)
	}
	panic("unhandle type signature")
}

func (m *Descript) ParseType(bs []byte) ([]byte, *ast.VariableType, error) {
	switch bs[0] {
	case 'V':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_VOID,
		}, nil
	case 'B':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_BYTE,
		}, nil
	case 'D':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_DOUBLE,
		}, nil
	case 'F':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_FLOAT,
		}, nil
	case 'I':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_INT,
		}, nil
	case 'J':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_LONG,
		}, nil
	case 'S', 'C':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_SHORT,
		}, nil
	case 'Z':
		bs = bs[1:]
		return bs, &ast.VariableType{
			Typ: ast.VARIABLE_TYPE_BOOL,
		}, nil
	case 'L':
		bs = bs[1:]
		index := bytes.Index(bs, []byte{';'}) // end token
		t := &ast.VariableType{}
		t.Typ = ast.VARIABLE_TYPE_OBJECT
		t.Class = &ast.Class{}
		t.Class.Name = string(bs[:index])
		bs = bs[index+1:] // skip ;
		t.Class.NotImportedYet = true
		if t.Class.Name == java_string_class {
			t.Typ = ast.VARIABLE_TYPE_STRING
		}
		return bs, t, nil
	case '[':
		bs = bs[1:]
		var t *ast.VariableType
		var err error
		bs, t, err = m.ParseType(bs)
		ret := &ast.VariableType{}
		if err == nil {
			ret.Typ = ast.VARIABLE_TYPE_JAVA_ARRAY
			ret.ArrayType = t
		}
		return bs, ret, err
	}
	return bs, nil, fmt.Errorf("unkown type:%v", string(bs))
}

func (m *Descript) ParseFunctionType(bs []byte) (ast.FunctionType, error) {
	t := ast.FunctionType{}
	if bs[0] != '(' {
		return t, fmt.Errorf("function descriptor does not start with '('")
	}
	bs = bs[1:]
	i := 1
	var err error
	for bs[0] != ')' {
		vd := &ast.VariableDefinition{}
		vd.Name = fmt.Sprintf("var_%d", i)
		bs, vd.Typ, err = m.ParseType(bs)
		if err != nil {
			return t, err
		}
		t.ParameterList = append(t.ParameterList, vd)
		i++
	}
	bs = bs[1:] // skip )
	vd := &ast.VariableDefinition{}
	vd.Name = "returnValue"
	_, vd.Typ, err = m.ParseType(bs)
	if err != nil {
		return t, err
	}
	t.ReturnList = append(t.ReturnList, vd)
	return t, nil
}
