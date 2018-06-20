package jvm

import (
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

const (
	_ = iota
	CLOSURE_INT_CLASS
	CLOSURE_LONG_CLASS
	CLOSURE_FLOAT_CLASS
	CLOSURE_DOUBLE_CLASS
	CLOSURE_STRING_CLASS
	CLOSURE_OBJECT_CLASS
)

type Closure struct {
	ClosureObjectMetas map[int]*ClosureObjectMeta
}

var (
	closure Closure
)

type ClosureObjectMeta struct {
	className       string
	fieldName       string
	fieldDescriptor string
}

func init() {
	closure.ClosureObjectMetas = make(map[int]*ClosureObjectMeta)
	closure.ClosureObjectMetas[CLOSURE_INT_CLASS] = &ClosureObjectMeta{
		className:       "lucy/deps/ClosureInt",
		fieldName:       "value",
		fieldDescriptor: "I",
	}
	closure.ClosureObjectMetas[CLOSURE_LONG_CLASS] = &ClosureObjectMeta{
		className:       "lucy/deps/ClosureLong",
		fieldName:       "value",
		fieldDescriptor: "J",
	}
	closure.ClosureObjectMetas[CLOSURE_FLOAT_CLASS] = &ClosureObjectMeta{
		className:       "lucy/deps/ClosureFloat",
		fieldName:       "value",
		fieldDescriptor: "F",
	}
	closure.ClosureObjectMetas[CLOSURE_DOUBLE_CLASS] = &ClosureObjectMeta{
		className:       "lucy/deps/ClosureDouble",
		fieldName:       "value",
		fieldDescriptor: "D",
	}
	closure.ClosureObjectMetas[CLOSURE_STRING_CLASS] = &ClosureObjectMeta{
		className:       "lucy/deps/ClosureString",
		fieldName:       "value",
		fieldDescriptor: "Ljava/lang/String;",
	}
	closure.ClosureObjectMetas[CLOSURE_OBJECT_CLASS] = &ClosureObjectMeta{
		className:       "lucy/deps/ClosureObject",
		fieldName:       "value",
		fieldDescriptor: "Ljava/lang/Object;",
	}
}

func (closure *Closure) getMeta(t int) (meta *ClosureObjectMeta) {
	switch t {
	case ast.VARIABLE_TYPE_BOOL:
		fallthrough
	case ast.VARIABLE_TYPE_BYTE:
		fallthrough
	case ast.VARIABLE_TYPE_SHORT:
		fallthrough
	case ast.VARIABLE_TYPE_ENUM:
		fallthrough
	case ast.VARIABLE_TYPE_INT:
		meta = closure.ClosureObjectMetas[CLOSURE_INT_CLASS]
	case ast.VARIABLE_TYPE_LONG:
		meta = closure.ClosureObjectMetas[CLOSURE_LONG_CLASS]
	case ast.VARIABLE_TYPE_FLOAT:
		meta = closure.ClosureObjectMetas[CLOSURE_FLOAT_CLASS]
	case ast.VARIABLE_TYPE_DOUBLE:
		meta = closure.ClosureObjectMetas[CLOSURE_DOUBLE_CLASS]
	case ast.VARIABLE_TYPE_STRING:
		meta = closure.ClosureObjectMetas[CLOSURE_STRING_CLASS]
	case ast.VARIABLE_TYPE_OBJECT:
		fallthrough
	case ast.VARIABLE_TYPE_ARRAY: //[]int
		fallthrough
	case ast.VARIABLE_TYPE_JAVA_ARRAY: // java array int[]
		fallthrough
	case ast.VARIABLE_TYPE_MAP:
		meta = closure.ClosureObjectMetas[CLOSURE_OBJECT_CLASS]
	}
	return
}

/*
	create a closure var, inited and leave on stack
*/
func (closure *Closure) createClosureVar(class *cg.ClassHighLevel,
	code *cg.AttributeCode, v *ast.Type) (maxStack uint16) {
	maxStack = 2
	var meta *ClosureObjectMeta
	switch v.Type {
	case ast.VARIABLE_TYPE_BOOL:
		fallthrough
	case ast.VARIABLE_TYPE_BYTE:
		fallthrough
	case ast.VARIABLE_TYPE_SHORT:
		fallthrough
	case ast.VARIABLE_TYPE_ENUM:
		fallthrough
	case ast.VARIABLE_TYPE_INT:
		meta = closure.ClosureObjectMetas[CLOSURE_INT_CLASS]
	case ast.VARIABLE_TYPE_LONG:
		meta = closure.ClosureObjectMetas[CLOSURE_LONG_CLASS]
	case ast.VARIABLE_TYPE_FLOAT:
		meta = closure.ClosureObjectMetas[CLOSURE_FLOAT_CLASS]
	case ast.VARIABLE_TYPE_DOUBLE:
		meta = closure.ClosureObjectMetas[CLOSURE_DOUBLE_CLASS]
	case ast.VARIABLE_TYPE_STRING:
		meta = closure.ClosureObjectMetas[CLOSURE_STRING_CLASS]
	case ast.VARIABLE_TYPE_OBJECT:
		fallthrough
	case ast.VARIABLE_TYPE_ARRAY: //[]int
		fallthrough
	case ast.VARIABLE_TYPE_JAVA_ARRAY: // java array int[]
		fallthrough
	case ast.VARIABLE_TYPE_MAP:
		meta = closure.ClosureObjectMetas[CLOSURE_OBJECT_CLASS]
	}
	code.Codes[code.CodeLength] = cg.OP_new
	class.InsertClassConst(meta.className, code.Codes[code.CodeLength+1:code.CodeLength+3])
	code.Codes[code.CodeLength+3] = cg.OP_dup
	code.CodeLength += 4
	code.Codes[code.CodeLength] = cg.OP_invokespecial
	class.InsertMethodRefConst(cg.CONSTANT_Methodref_info_high_level{
		Class:      meta.className,
		Method:     special_method_init,
		Descriptor: "()V",
	}, code.Codes[code.CodeLength+1:code.CodeLength+3])
	code.CodeLength += 3
	return
}

func (closure *Closure) storeLocalClosureVar(class *cg.ClassHighLevel, code *cg.AttributeCode, v *ast.Variable) {
	var meta *ClosureObjectMeta
	switch v.Type.Type {
	case ast.VARIABLE_TYPE_BOOL:
		fallthrough
	case ast.VARIABLE_TYPE_BYTE:
		fallthrough
	case ast.VARIABLE_TYPE_SHORT:
		fallthrough
	case ast.VARIABLE_TYPE_ENUM:
		fallthrough
	case ast.VARIABLE_TYPE_INT:
		meta = closure.ClosureObjectMetas[CLOSURE_INT_CLASS]
	case ast.VARIABLE_TYPE_LONG:
		meta = closure.ClosureObjectMetas[CLOSURE_LONG_CLASS]
	case ast.VARIABLE_TYPE_FLOAT:
		meta = closure.ClosureObjectMetas[CLOSURE_FLOAT_CLASS]
	case ast.VARIABLE_TYPE_DOUBLE:
		meta = closure.ClosureObjectMetas[CLOSURE_DOUBLE_CLASS]
	case ast.VARIABLE_TYPE_STRING:
		meta = closure.ClosureObjectMetas[CLOSURE_STRING_CLASS]
	case ast.VARIABLE_TYPE_OBJECT:
		fallthrough
	case ast.VARIABLE_TYPE_MAP:
		fallthrough
	case ast.VARIABLE_TYPE_ARRAY:
		fallthrough
	case ast.VARIABLE_TYPE_JAVA_ARRAY:
		meta = closure.ClosureObjectMetas[CLOSURE_OBJECT_CLASS]
	}
	code.Codes[code.CodeLength] = cg.OP_putfield
	class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
		Class:      meta.className,
		Field:      meta.fieldName,
		Descriptor: meta.fieldDescriptor,
	}, code.Codes[code.CodeLength+1:code.CodeLength+3])
	code.CodeLength += 3
}

/*
	create a closure var on stack
*/
func (closure *Closure) loadLocalClosureVar(class *cg.ClassHighLevel, code *cg.AttributeCode,
	v *ast.Variable) (maxStack uint16) {
	copyOPs(code, loadLocalVariableOps(ast.VARIABLE_TYPE_OBJECT, v.LocalValOffset)...)
	closure.unPack(class, code, v.Type)
	maxStack = jvmSlotSize(v.Type)
	return
}

/*
	closure object is on stack
*/
func (closure *Closure) unPack(class *cg.ClassHighLevel, code *cg.AttributeCode, v *ast.Type) {
	switch v.Type {
	case ast.VARIABLE_TYPE_BOOL:
		fallthrough
	case ast.VARIABLE_TYPE_BYTE:
		fallthrough
	case ast.VARIABLE_TYPE_SHORT:
		fallthrough
	case ast.VARIABLE_TYPE_ENUM:
		fallthrough
	case ast.VARIABLE_TYPE_INT:
		meta := closure.ClosureObjectMetas[CLOSURE_INT_CLASS]
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      meta.className,
			Field:      meta.fieldName,
			Descriptor: meta.fieldDescriptor,
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3

	case ast.VARIABLE_TYPE_LONG:
		meta := closure.ClosureObjectMetas[CLOSURE_LONG_CLASS]
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      meta.className,
			Field:      meta.fieldName,
			Descriptor: meta.fieldDescriptor,
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3

	case ast.VARIABLE_TYPE_FLOAT:
		meta := closure.ClosureObjectMetas[CLOSURE_FLOAT_CLASS]
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      meta.className,
			Field:      meta.fieldName,
			Descriptor: meta.fieldDescriptor,
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3

	case ast.VARIABLE_TYPE_DOUBLE:
		meta := closure.ClosureObjectMetas[CLOSURE_DOUBLE_CLASS]
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      meta.className,
			Field:      meta.fieldName,
			Descriptor: meta.fieldDescriptor,
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3

	case ast.VARIABLE_TYPE_STRING:
		meta := closure.ClosureObjectMetas[CLOSURE_STRING_CLASS]
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      meta.className,
			Field:      meta.fieldName,
			Descriptor: meta.fieldDescriptor,
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3

	case ast.VARIABLE_TYPE_MAP:
		fallthrough
	case ast.VARIABLE_TYPE_OBJECT:
		fallthrough
	case ast.VARIABLE_TYPE_ARRAY:
		fallthrough
	case ast.VARIABLE_TYPE_JAVA_ARRAY:
		meta := closure.ClosureObjectMetas[CLOSURE_OBJECT_CLASS]
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      meta.className,
			Field:      meta.fieldName,
			Descriptor: meta.fieldDescriptor,
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3
		typeConverter.castPointerTypeToRealType(class, code, v)

	}

}
