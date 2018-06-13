package jvm

import (
	"encoding/binary"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

func (m *MakeExpression) getCaptureIdentifierLeftValue(
	class *cg.ClassHighLevel, code *cg.AttributeCode, e *ast.Expression,
	context *Context, state *StackMapState) (
	maxStack, remainStack uint16, op []byte,
	target *ast.VariableType, className, fieldName, fieldDescriptor string) {
	identifier := e.Data.(*ast.ExpressionIdentifier)
	target = identifier.Var.Typ
	op = []byte{cg.OP_putfield}
	meta := closure.getMeta(identifier.Var.Typ.Typ)
	if context.function.Closure.ClosureVariableExist(identifier.Var) { // capture var exits
		copyOP(code, loadSimpleVarOp(ast.VARIABLE_TYPE_OBJECT, 0)...)
		code.Codes[code.CodeLength] = cg.OP_getfield
		class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
			Class:      class.Name,
			Field:      identifier.Name,
			Descriptor: "L" + meta.className + ";",
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3
	} else {
		copyOP(code, loadSimpleVarOp(ast.VARIABLE_TYPE_OBJECT, identifier.Var.LocalValOffset)...)
	}
	state.pushStack(class, state.newObjectVariableType(meta.className))
	maxStack = 1
	remainStack = 1
	className = meta.className
	fieldName = meta.fieldName
	fieldDescriptor = meta.fieldDescriptor
	return
}

func (m *MakeExpression) getMapLeftValue(
	class *cg.ClassHighLevel, code *cg.AttributeCode, e *ast.Expression,
	context *Context, state *StackMapState) (
	maxStack, remainStack uint16, op []byte,
	target *ast.VariableType, className, name, descriptor string) {
	index := e.Data.(*ast.ExpressionIndex)
	maxStack, _ = m.build(class, code, index.Expression, context, state)
	state.pushStack(class, state.newObjectVariableType(java_hashmap_class))
	stack, _ := m.build(class, code, index.Index, context, state)
	if t := 1 + stack; t > maxStack {
		maxStack = t
	}
	if index.Index.Value.IsPointer() == false {
		typeConverter.putPrimitiveInObject(class, code, index.Index.Value)
	}
	state.pushStack(class, state.newObjectVariableType(java_root_class))
	remainStack = 2
	op = []byte{}
	if index.Expression.Value.Map.V.IsPointer() == false {
		op = append(op,
			typeConverter.putPrimitiveInObjectBytes(class, index.Expression.Value.Map.V)...)
	}
	bs4 := make([]byte, 4)
	bs4[0] = cg.OP_invokevirtual
	class.InsertMethodRefConst(cg.CONSTANT_Methodref_info_high_level{
		Class:      java_hashmap_class,
		Method:     "put",
		Descriptor: "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;",
	}, bs4[1:3])
	bs4[3] = cg.OP_pop
	op = append(op, bs4...)
	target = index.Expression.Value.Map.V
	className = java_hashmap_class
	return
}

func (m *MakeExpression) getLeftValue(
	class *cg.ClassHighLevel, code *cg.AttributeCode,
	e *ast.Expression, context *Context, state *StackMapState) (
	maxStack, remainStack uint16, op []byte,
	target *ast.VariableType, className, name, descriptor string) {
	switch e.Typ {
	case ast.EXPRESSION_TYPE_IDENTIFIER:
		identifier := e.Data.(*ast.ExpressionIdentifier)
		if identifier.Var.IsGlobal {
			op = []byte{cg.OP_putstatic}
			target = identifier.Var.Typ
			className = m.MakeClass.mainClass.Name
			name = identifier.Name
			descriptor = Descriptor.typeDescriptor(identifier.Var.Typ)
			return
		}
		if identifier.Var.BeenCaptured {
			return m.getCaptureIdentifierLeftValue(class, code, e, context, state)
		}
		if identifier.Name == ast.NO_NAME_IDENTIFIER {
			panic("this is not happening")
		}
		switch identifier.Var.Typ.Typ {
		case ast.VARIABLE_TYPE_BOOL:
			fallthrough
		case ast.VARIABLE_TYPE_BYTE:
			fallthrough
		case ast.VARIABLE_TYPE_SHORT:
			fallthrough
		case ast.VARIABLE_TYPE_ENUM:
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
				panic("over 255")
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
				panic("over 255")
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
				panic("over 255")
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
				panic("over 255")
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
				panic("over 255")
			}
		}
		target = identifier.Var.Typ
	case ast.EXPRESSION_TYPE_INDEX:
		index := e.Data.(*ast.ExpressionIndex)
		if index.Expression.Value.Typ == ast.VARIABLE_TYPE_ARRAY {
			meta := ArrayMetas[index.Expression.Value.ArrayType.Typ]
			maxStack, _ = m.build(class, code, index.Expression, context, state)
			state.pushStack(class, index.Expression.Value)
			code.Codes[code.CodeLength] = cg.OP_dup
			code.CodeLength++
			code.Codes[code.CodeLength] = cg.OP_getfield
			class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
				Class:      meta.className,
				Field:      "end",
				Descriptor: "I",
			}, code.Codes[code.CodeLength+1:code.CodeLength+3])
			code.CodeLength += 3
			code.Codes[code.CodeLength] = cg.OP_swap
			code.CodeLength++
			code.Codes[code.CodeLength] = cg.OP_dup_x1
			code.CodeLength++
			code.Codes[code.CodeLength] = cg.OP_getfield
			class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
				Class:      meta.className,
				Field:      "start",
				Descriptor: "I",
			}, code.Codes[code.CodeLength+1:code.CodeLength+3])
			code.CodeLength += 3
			state.pushStack(class, &ast.VariableType{Typ: ast.VARIABLE_TYPE_INT})
			state.pushStack(class, &ast.VariableType{Typ: ast.VARIABLE_TYPE_INT})

			stack, _ := m.build(class, code, index.Index, context, state)
			if t := stack + 3; t > maxStack {
				maxStack = t
			}
			code.Codes[code.CodeLength] = cg.OP_iadd
			code.CodeLength++
			code.Codes[code.CodeLength] = cg.OP_dup_x1
			code.CodeLength++
			{
				state.popStack(3)
				state.pushStack(class, state.newObjectVariableType(meta.className))
				state.pushStack(class, &ast.VariableType{Typ: ast.VARIABLE_TYPE_INT})
				context.MakeStackMap(code, state, code.CodeLength+6)
				context.MakeStackMap(code, state, code.CodeLength+16)
				state.popStack(2)
			}
			code.Codes[code.CodeLength] = cg.OP_if_icmple
			binary.BigEndian.PutUint16(code.Codes[code.CodeLength+1:code.CodeLength+3], 6)
			code.Codes[code.CodeLength+3] = cg.OP_goto
			binary.BigEndian.PutUint16(code.Codes[code.CodeLength+4:code.CodeLength+6], 13)
			code.Codes[code.CodeLength+6] = cg.OP_pop // incase stack over flow
			code.Codes[code.CodeLength+7] = cg.OP_pop
			code.Codes[code.CodeLength+8] = cg.OP_new
			class.InsertClassConst(java_index_out_of_range_exception_class, code.Codes[code.CodeLength+9:code.CodeLength+11])
			code.Codes[code.CodeLength+11] = cg.OP_dup
			code.Codes[code.CodeLength+12] = cg.OP_invokespecial
			class.InsertMethodRefConst(cg.CONSTANT_Methodref_info_high_level{
				Class:      java_index_out_of_range_exception_class,
				Method:     special_method_init,
				Descriptor: "()V",
			}, code.Codes[code.CodeLength+13:code.CodeLength+15])
			code.Codes[code.CodeLength+15] = cg.OP_athrow
			// index not out of range
			code.Codes[code.CodeLength+16] = cg.OP_swap
			code.Codes[code.CodeLength+17] = cg.OP_getfield
			class.InsertFieldRefConst(cg.CONSTANT_Fieldref_info_high_level{
				Class:      meta.className,
				Field:      "elements",
				Descriptor: meta.elementsFieldDescriptor,
			}, code.Codes[code.CodeLength+18:code.CodeLength+20])
			code.CodeLength += 20
			code.Codes[code.CodeLength] = cg.OP_swap
			code.CodeLength++
			{
				t := &ast.VariableType{}
				t.Typ = ast.VARIABLE_TYPE_JAVA_ARRAY
				t.ArrayType = index.Expression.Value.ArrayType
				state.pushStack(class, t)
				state.pushStack(class, &ast.VariableType{Typ: ast.VARIABLE_TYPE_INT})
			}
			switch e.Value.Typ {
			case ast.VARIABLE_TYPE_BOOL:
				op = []byte{cg.OP_bastore}
			case ast.VARIABLE_TYPE_BYTE:
				op = []byte{cg.OP_bastore}
			case ast.VARIABLE_TYPE_SHORT:
				op = []byte{cg.OP_sastore}
			case ast.VARIABLE_TYPE_ENUM:
				fallthrough
			case ast.VARIABLE_TYPE_INT:
				op = []byte{cg.OP_iastore}
			case ast.VARIABLE_TYPE_LONG:
				op = []byte{cg.OP_lastore}
			case ast.VARIABLE_TYPE_FLOAT:
				op = []byte{cg.OP_fastore}
			case ast.VARIABLE_TYPE_DOUBLE:
				op = []byte{cg.OP_dastore}
			case ast.VARIABLE_TYPE_STRING:
				fallthrough
			case ast.VARIABLE_TYPE_OBJECT:
				fallthrough
			case ast.VARIABLE_TYPE_MAP:
				fallthrough
			case ast.VARIABLE_TYPE_ARRAY:
				fallthrough
			case ast.VARIABLE_TYPE_JAVA_ARRAY:
				op = []byte{cg.OP_aastore}
			}
			remainStack = 2 // [arrayref ,index]
			target = e.Value
		} else if index.Expression.Value.Typ == ast.VARIABLE_TYPE_MAP { // map
			return m.getMapLeftValue(class, code, e, context, state)
		} else { // java array
			maxStack, _ = m.build(class, code, index.Expression, context, state)
			state.pushStack(class, index.Expression.Value)
			stack, _ := m.build(class, code, index.Index, context, state)
			if t := stack + 1; t > maxStack {
				maxStack = t
			}
			target = e.Value
			remainStack = 2 // [objectref ,index]
			state.pushStack(class, &ast.VariableType{Typ: ast.VARIABLE_TYPE_INT})
			switch e.Value.Typ {
			case ast.VARIABLE_TYPE_BOOL:
				op = []byte{cg.OP_bastore}
			case ast.VARIABLE_TYPE_BYTE:
				op = []byte{cg.OP_bastore}
			case ast.VARIABLE_TYPE_SHORT:
				op = []byte{cg.OP_sastore}
			case ast.VARIABLE_TYPE_ENUM:
				fallthrough
			case ast.VARIABLE_TYPE_INT:
				op = []byte{cg.OP_iastore}
			case ast.VARIABLE_TYPE_LONG:
				op = []byte{cg.OP_lastore}
			case ast.VARIABLE_TYPE_FLOAT:
				op = []byte{cg.OP_fastore}
			case ast.VARIABLE_TYPE_DOUBLE:
				op = []byte{cg.OP_dastore}
			case ast.VARIABLE_TYPE_STRING:
				fallthrough
			case ast.VARIABLE_TYPE_OBJECT:
				fallthrough
			case ast.VARIABLE_TYPE_MAP:
				fallthrough
			case ast.VARIABLE_TYPE_ARRAY:
				fallthrough
			case ast.VARIABLE_TYPE_JAVA_ARRAY:
				op = []byte{cg.OP_aastore}
			}
			return
		}
	case ast.EXPRESSION_TYPE_SELECT:
		dot := e.Data.(*ast.ExpressionSelection)
		if dot.Expression.Value.Typ == ast.VARIABLE_TYPE_PACKAGE {
			op = []byte{cg.OP_putstatic}
			target = dot.PackageVariable.Typ
			className = dot.Expression.Value.Package.Name + "/main"
			name = dot.PackageVariable.Name
			descriptor = dot.PackageVariable.Descriptor
			maxStack = 0
			remainStack = 0
		} else {
			className = dot.Expression.Value.Class.Name
			target = dot.Field.VariableDefinition.Typ
			name = dot.Name
			if dot.Field.LoadFromOutSide {
				descriptor = dot.Field.Descriptor
			} else {
				descriptor = Descriptor.typeDescriptor(target)
			}
			if dot.Field.IsStatic() {
				op = []byte{cg.OP_putstatic}
			} else {
				maxStack, _ = m.build(class, code, dot.Expression, context, state)
				remainStack = 1
				state.pushStack(class, dot.Expression.Value)
				op = []byte{cg.OP_putfield}
			}
		}
	}
	return
}
