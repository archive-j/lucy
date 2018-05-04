package jvm

import (
	"encoding/binary"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

func (m *MakeExpression) buildUnary(class *cg.ClassHighLevel, code *cg.AttributeCode,
	e *ast.Expression, context *Context, state *StackMapState) (maxstack uint16) {

	if e.Typ == ast.EXPRESSION_TYPE_NEGATIVE {
		maxstack, _ = m.build(class, code, e.Data.(*ast.Expression), context, state)
		switch e.Value.Typ {
		case ast.VARIABLE_TYPE_BYTE:
			fallthrough
		case ast.VARIABLE_TYPE_SHORT:
			fallthrough
		case ast.VARIABLE_TYPE_INT:
			code.Codes[code.CodeLength] = cg.OP_ineg
		case ast.VARIABLE_TYPE_FLOAT:
			code.Codes[code.CodeLength] = cg.OP_fneg
		case ast.VARIABLE_TYPE_DOUBLE:
			code.Codes[code.CodeLength] = cg.OP_dneg
		case ast.VARIABLE_TYPE_LONG:
			code.Codes[code.CodeLength] = cg.OP_lneg
		}
		code.CodeLength++
		return
	}
	if e.Typ == ast.EXPRESSION_TYPE_BITWISE_ {
		ee := e.Data.(*ast.Expression)
		maxstack, _ = m.build(class, code, ee, context, state)
		if t := jvmSize(ee.Value) * 2; t > maxstack {
			maxstack = t
		}
		switch e.Value.Typ {
		case ast.VARIABLE_TYPE_BYTE:
			code.Codes[code.CodeLength] = cg.OP_bipush
			code.Codes[code.CodeLength+1] = 255
			code.Codes[code.CodeLength+2] = cg.OP_ixor
			code.CodeLength += 3
		case ast.VARIABLE_TYPE_SHORT:
			code.Codes[code.CodeLength] = cg.OP_sipush
			code.Codes[code.CodeLength+1] = 255
			code.Codes[code.CodeLength+2] = 255
			code.Codes[code.CodeLength+3] = cg.OP_ixor
			code.CodeLength += 4
		case ast.VARIABLE_TYPE_INT:
			code.Codes[code.CodeLength] = cg.OP_ldc_w
			binary.BigEndian.PutUint16(code.Codes[code.CodeLength+1:code.CodeLength+3], uint16(len(class.Class.ConstPool)))
			code.Codes[code.CodeLength+3] = cg.OP_ixor
			code.CodeLength += 4
			t := &cg.ConstPool{}
			t.Tag = cg.CONSTANT_POOL_TAG_Integer
			t.Info = make([]byte, 4)
			t.Info[0] = 255
			t.Info[1] = 255
			t.Info[2] = 255
			t.Info[3] = 255
			class.Class.ConstPool = append(class.Class.ConstPool, t)
		case ast.VARIABLE_TYPE_LONG:
			code.Codes[code.CodeLength] = cg.OP_ldc2_w
			binary.BigEndian.PutUint16(code.Codes[code.CodeLength+1:code.CodeLength+3], uint16(len(class.Class.ConstPool)))
			code.Codes[code.CodeLength+3] = cg.OP_lxor
			code.CodeLength += 4
			t := &cg.ConstPool{}
			t.Tag = cg.CONSTANT_POOL_TAG_Long
			t.Info = make([]byte, 8)
			t.Info[0] = 255
			t.Info[1] = 255
			t.Info[2] = 255
			t.Info[3] = 255
			t.Info[4] = 255
			t.Info[5] = 255
			t.Info[6] = 255
			t.Info[7] = 255
			class.Class.ConstPool = append(class.Class.ConstPool, t, nil)
		}
		return
	}

	if e.Typ == ast.EXPRESSION_TYPE_NOT {
		ee := e.Data.(*ast.Expression)
		var es []*cg.JumpBackPatch
		maxstack, es = m.build(class, code, ee, context, state)
		if len(es) > 0 {

			state.pushStack(class, ee.Value)
			backPatchEs(es, code.CodeLength)
			context.MakeStackMap(code, state, code.CodeLength)
			state.popStack(1)
		}

		context.MakeStackMap(code, state, code.CodeLength+7)
		state.pushStack(class, ee.Value)
		context.MakeStackMap(code, state, code.CodeLength+8)
		state.popStack(1)
		code.Codes[code.CodeLength] = cg.OP_ifne
		binary.BigEndian.PutUint16(code.Codes[code.CodeLength+1:], uint16(7))
		code.Codes[code.CodeLength+3] = cg.OP_iconst_1
		code.Codes[code.CodeLength+4] = cg.OP_goto
		binary.BigEndian.PutUint16(code.Codes[code.CodeLength+5:], uint16(4))
		code.Codes[code.CodeLength+7] = cg.OP_iconst_0
		code.CodeLength += 8
	}

	return
}
