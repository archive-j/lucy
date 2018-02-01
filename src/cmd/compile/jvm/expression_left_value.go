package jvm

import (
	"github.com/756445638/lucy/src/cmd/compile/ast"
	"github.com/756445638/lucy/src/cmd/compile/jvm/cg"
)

func (m *MakeExpression) getCaptureIdentiferLeftValue(class *cg.ClassHighLevel, code *cg.AttributeCode, e *ast.Expression, context *Context) (maxstack, remainStack uint16, op []byte, target *ast.VariableType, classname, fieldname, fieldDescriptor string) {
	return
}

func (m *MakeExpression) getLeftValue(class *cg.ClassHighLevel, code *cg.AttributeCode, e *ast.Expression, context *Context) (maxstack, remainStack uint16, op []byte, target *ast.VariableType, classname, fieldname, fieldDescriptor string) {
	switch e.Typ {
	case ast.EXPRESSION_TYPE_IDENTIFIER:
		identifier := e.Data.(*ast.ExpressionIdentifer)
		if identifier.Var.BeenCaptured {
			return m.getCaptureIdentiferLeftValue(class, code, e, context)
		}
		if identifier.Name == ast.NO_NAME_IDENTIFIER {
			return //
		}
		switch identifier.Var.Typ.Typ {
		case ast.VARIABLE_TYPE_BOOL:
			fallthrough
		case ast.VARIABLE_TYPE_BYTE:
			fallthrough
		case ast.VARIABLE_TYPE_SHORT:
			fallthrough
		case ast.VARIABLE_TYPE_INT:
			if identifier.Var.LocalValOffset == 0 {
				op = []byte{cg.OP_istore_0}
			} else if identifier.Var.LocalValOffset == 1 {
				op = []byte{cg.OP_istore_1}
			} else if identifier.Var.LocalValOffset == 2 {
				op = []byte{cg.OP_istore_2}
			} else if identifier.Var.LocalValOffset == 3 {
				op = []byte{cg.OP_istore_3}
			} else if identifier.Var.LocalValOffset <= 255 {
				op = []byte{cg.OP_istore, byte(identifier.Var.LocalValOffset)}
			} else {
				panic("local int var offset > 255")
			}
		case ast.VARIABLE_TYPE_FLOAT:
			if identifier.Var.LocalValOffset == 0 {
				op = []byte{cg.OP_fstore_0}
			} else if identifier.Var.LocalValOffset == 1 {
				op = []byte{cg.OP_fstore_1}
			} else if identifier.Var.LocalValOffset == 2 {
				op = []byte{cg.OP_fstore_2}
			} else if identifier.Var.LocalValOffset == 3 {
				op = []byte{cg.OP_fstore_3}
			} else if identifier.Var.LocalValOffset <= 255 {
				op = []byte{cg.OP_fstore, byte(identifier.Var.LocalValOffset)}
			} else {
				panic("local float var out of range")
			}
		case ast.VARIABLE_TYPE_DOUBLE:
			if identifier.Var.LocalValOffset == 0 {
				op = []byte{cg.OP_dstore_0}
			} else if identifier.Var.LocalValOffset == 1 {
				op = []byte{cg.OP_dstore_1}
			} else if identifier.Var.LocalValOffset == 2 {
				op = []byte{cg.OP_dstore_2}
			} else if identifier.Var.LocalValOffset == 3 {
				op = []byte{cg.OP_dstore_3}
			} else if identifier.Var.LocalValOffset <= 255 {
				op = []byte{cg.OP_dstore, byte(identifier.Var.LocalValOffset)}
			} else {
				panic("local float var out of range")
			}
		case ast.VARIABLE_TYPE_LONG:
			if identifier.Var.LocalValOffset == 0 {
				op = []byte{cg.OP_lstore_0}
			} else if identifier.Var.LocalValOffset == 1 {
				op = []byte{cg.OP_lstore_1}
			} else if identifier.Var.LocalValOffset == 2 {
				op = []byte{cg.OP_lstore_2}
			} else if identifier.Var.LocalValOffset == 3 {
				op = []byte{cg.OP_lstore_3}
			} else if identifier.Var.LocalValOffset <= 255 {
				op = []byte{cg.OP_lstore, byte(identifier.Var.LocalValOffset)}
			} else {
				panic("local float var out of range")
			}
		default: // must be a object type
			if identifier.Var.LocalValOffset == 0 {
				op = []byte{cg.OP_astore_0}
			} else if identifier.Var.LocalValOffset == 1 {
				op = []byte{cg.OP_astore_1}
			} else if identifier.Var.LocalValOffset == 2 {
				op = []byte{cg.OP_astore_2}
			} else if identifier.Var.LocalValOffset == 3 {
				op = []byte{cg.OP_astore_3}
			} else if identifier.Var.LocalValOffset <= 255 {
				op = []byte{cg.OP_astore, byte(identifier.Var.LocalValOffset)}
			} else {
				panic("local float var out of range")
			}
		}
		target = identifier.Var.Typ
	case ast.EXPRESSION_TYPE_INDEX:
		index := e.Data.(*ast.ExpressionIndex)
		maxstack, _ = m.build(class, code, index.Expression, context)
		stack, _ := m.build(class, code, index.Index, context)
		if t := stack + 1; t > maxstack {
			maxstack = t
		}
		switch e.VariableType.Typ {
		case ast.VARIABLE_TYPE_BOOL:
			fallthrough
		case ast.VARIABLE_TYPE_BYTE:
			op = []byte{cg.OP_bastore}
		case ast.VARIABLE_TYPE_SHORT:
			op = []byte{cg.OP_sastore}
		case ast.VARIABLE_TYPE_INT:
			op = []byte{cg.OP_iastore}
		case ast.VARIABLE_TYPE_FLOAT:
			op = []byte{cg.OP_fastore}
		case ast.VARIABLE_TYPE_DOUBLE:
			op = []byte{cg.OP_dastore}
		case ast.VARIABLE_TYPE_LONG:
			op = []byte{cg.OP_lastore}
		default:
			op = []byte{cg.OP_aastore}
		}
		remainStack = 2 // objectref index
	case ast.EXPRESSION_TYPE_DOT:
		index := e.Data.(*ast.ExpressionIndex)
		if index.Expression.Typ == ast.VARIABLE_TYPE_CLASS {
			op = []byte{cg.OP_getstatic}
			classname = index.Expression.VariableType.Class.Name
			fieldname = index.Name
			fieldDescriptor = index.Field.Descriptor
		} else {
			maxstack, _ = m.build(class, code, index.Expression, context)
			classname = index.Expression.VariableType.Class.Name
			fieldname = index.Name
			fieldDescriptor = index.Field.Descriptor
		}
	default:
		panic("unkown type ")
	}
	return
}
