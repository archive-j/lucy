package cg

import (
	"encoding/binary"
)

const (
	CONSTANT_POOL_MAX_SIZE                            = 65536
	CLASS_MAGIC                                uint32 = 0xcafebabe
	ATTRIBUTE_NAME_SOURCE_FILE                        = "SourceFile"
	ATTRIBUTE_NAME_LUCY_FIELD_DESCRIPTOR              = "LucyFieldDescriptor"
	ATTRIBUTE_NAME_LUCY_METHOD_DESCRIPTOR             = "LucyMethodDescriptor"
	ATTRIBUTE_NAME_LUCY_CLOSURE_FUNCTION_CLASS        = "LucyClosureFunctionClass"
	ATTRIBUTE_NAME_LUCY_INNER_STATIC_METHOD           = "LucyInnerStaticMethod"
	ATTRIBUTE_NAME_CONST_VALUE                        = "ConstantValue"
	ATTRIBUTE_LUCY_TYPE_ALIAS                         = "LucyTypeAlias"
)

func backPatchIndex(locations [][]byte, index uint16) {
	for _, v := range locations {
		binary.BigEndian.PutUint16(v, index)
	}
}

type JumpBackPatch struct {
	CurrentCodeLength int
	Bs                []byte
}

func (j *JumpBackPatch) FromCode(op byte, code *AttributeCode) *JumpBackPatch {
	j.CurrentCodeLength = code.CodeLength
	code.Codes[code.CodeLength] = op
	j.Bs = code.Codes[code.CodeLength+1 : code.CodeLength+3]
	code.CodeLength += 3
	return j
}
