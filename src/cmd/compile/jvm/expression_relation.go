package jvm

import (
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

func (buildExpression *BuildExpression) buildRelations(class *cg.ClassHighLevel, code *cg.AttributeCode,
	e *ast.Expression, context *Context, state *StackMapState) (maxStack uint16) {
	bin := e.Data.(*ast.ExpressionBinary)
	stackLength := len(state.Stacks)
	defer func() {
		state.popStack(len(state.Stacks) - stackLength)
	}()
	if bin.Left.Value.IsNumber() ||
		bin.Left.Value.Type == ast.VariableTypeEnum { // in this case ,right must be a number type
		maxStack = buildExpression.build(class, code, bin.Left, context, state)
		state.pushStack(class, bin.Left.Value)
		stack := buildExpression.build(class, code, bin.Right, context, state)
		if t := jvmSlotSize(bin.Left.Value) + stack; t > maxStack {
			maxStack = t
		}
		var exit *cg.Exit
		if bin.Left.Value.Type == ast.VariableTypeByte ||
			bin.Left.Value.Type == ast.VariableTypeShort ||
			bin.Left.Value.Type == ast.VariableTypeChar ||
			bin.Left.Value.Type == ast.VariableTypeInt ||
			bin.Left.Value.Type == ast.VariableTypeEnum {
			switch e.Type {
			case ast.ExpressionTypeGt:
				exit = (&cg.Exit{}).Init(cg.OP_if_icmpgt, code)
			case ast.ExpressionTypeLe:
				exit = (&cg.Exit{}).Init(cg.OP_if_icmple, code)
			case ast.ExpressionTypeLt:
				exit = (&cg.Exit{}).Init(cg.OP_if_icmplt, code)
			case ast.ExpressionTypeGe:
				exit = (&cg.Exit{}).Init(cg.OP_if_icmpge, code)
			case ast.ExpressionTypeEq:
				exit = (&cg.Exit{}).Init(cg.OP_if_icmpeq, code)
			case ast.ExpressionTypeNe:
				exit = (&cg.Exit{}).Init(cg.OP_if_icmpne, code)
			}
		} else {
			switch bin.Left.Value.Type {
			case ast.VariableTypeLong:
				code.Codes[code.CodeLength] = cg.OP_lcmp
			case ast.VariableTypeFloat:
				code.Codes[code.CodeLength] = cg.OP_fcmpl
			case ast.VariableTypeDouble:
				code.Codes[code.CodeLength] = cg.OP_dcmpl
			}
			code.CodeLength++
			switch e.Type {
			case ast.ExpressionTypeGt:
				exit = (&cg.Exit{}).Init(cg.OP_ifgt, code)
			case ast.ExpressionTypeLe:
				exit = (&cg.Exit{}).Init(cg.OP_ifle, code)
			case ast.ExpressionTypeLt:
				exit = (&cg.Exit{}).Init(cg.OP_iflt, code)
			case ast.ExpressionTypeGe:
				exit = (&cg.Exit{}).Init(cg.OP_ifge, code)
			case ast.ExpressionTypeEq:
				exit = (&cg.Exit{}).Init(cg.OP_ifeq, code)
			case ast.ExpressionTypeNe:
				exit = (&cg.Exit{}).Init(cg.OP_ifne, code)
			}
		}
		state.popStack(1)
		code.Codes[code.CodeLength] = cg.OP_iconst_0
		code.CodeLength++
		falseExit := (&cg.Exit{}).Init(cg.OP_goto, code)
		writeExits([]*cg.Exit{exit}, code.CodeLength)
		context.MakeStackMap(code, state, code.CodeLength)
		code.Codes[code.CodeLength] = cg.OP_iconst_1
		code.CodeLength++
		writeExits([]*cg.Exit{falseExit}, code.CodeLength)
		state.pushStack(class, &ast.Type{
			Type: ast.VariableTypeBool,
		})
		context.MakeStackMap(code, state, code.CodeLength)
		defer state.popStack(1)
		return
	}
	if bin.Left.Value.Type == ast.VariableTypeBool ||
		bin.Right.Value.Type == ast.VariableTypeBool { // bool type
		maxStack = buildExpression.build(class, code, bin.Left, context, state)
		state.pushStack(class, bin.Left.Value)
		stack := buildExpression.build(class, code, bin.Right, context, state)
		if t := jvmSlotSize(bin.Left.Value) + stack; t > maxStack {
			maxStack = t
		}
		state.popStack(1) // 1 bool value
		var exit *cg.Exit
		if e.Type == ast.ExpressionTypeEq {
			exit = (&cg.Exit{}).Init(cg.OP_if_icmpeq, code)
		} else {
			exit = (&cg.Exit{}).Init(cg.OP_if_icmpne, code)
		}
		code.Codes[code.CodeLength] = cg.OP_iconst_0
		code.CodeLength++
		falseExit := (&cg.Exit{}).Init(cg.OP_goto, code)
		writeExits([]*cg.Exit{exit}, code.CodeLength)
		context.MakeStackMap(code, state, code.CodeLength)
		code.Codes[code.CodeLength] = cg.OP_iconst_1
		code.CodeLength++
		writeExits([]*cg.Exit{falseExit}, code.CodeLength)
		state.pushStack(class, &ast.Type{
			Type: ast.VariableTypeBool,
		})
		context.MakeStackMap(code, state, code.CodeLength)
		defer state.popStack(1)
		return
	}
	if bin.Left.Value.Type == ast.VariableTypeNull ||
		bin.Right.Value.Type == ast.VariableTypeNull { // must not null-null
		var notNullExpression *ast.Expression
		if bin.Left.Value.Type != ast.VariableTypeNull {
			notNullExpression = bin.Left
		} else {
			notNullExpression = bin.Right
		}
		maxStack = buildExpression.build(class, code, notNullExpression, context, state)
		var exit *cg.Exit
		if e.Type == ast.ExpressionTypeEq {
			exit = (&cg.Exit{}).Init(cg.OP_ifnull, code)
		} else { // ne
			exit = (&cg.Exit{}).Init(cg.OP_ifnonnull, code)
		}
		code.Codes[code.CodeLength] = cg.OP_iconst_0
		code.CodeLength++
		falseExit := (&cg.Exit{}).Init(cg.OP_goto, code)
		writeExits([]*cg.Exit{exit}, code.CodeLength)
		context.MakeStackMap(code, state, code.CodeLength)
		code.Codes[code.CodeLength] = cg.OP_iconst_1
		code.CodeLength++
		writeExits([]*cg.Exit{falseExit}, code.CodeLength)
		state.pushStack(class, &ast.Type{
			Type: ast.VariableTypeBool,
		})
		context.MakeStackMap(code, state, code.CodeLength)
		defer state.popStack(1)
		return
	}

	//string compare
	if bin.Left.Value.Type == ast.VariableTypeString {
		maxStack = buildExpression.build(class, code, bin.Left, context, state)
		state.pushStack(class, bin.Left.Value)
		stack := buildExpression.build(class, code, bin.Right, context, state)
		code.Codes[code.CodeLength] = cg.OP_invokevirtual
		class.InsertMethodRefConst(cg.CONSTANT_Methodref_info_high_level{
			Class:      javaStringClass,
			Method:     "compareTo",
			Descriptor: "(Ljava/lang/String;)I",
		}, code.Codes[code.CodeLength+1:code.CodeLength+3])
		code.CodeLength += 3
		if t := 1 + stack; t > maxStack {
			maxStack = t
		}
		state.popStack(1) // pop left string
		var exit *cg.Exit
		switch e.Type {
		case ast.ExpressionTypeGt:
			exit = (&cg.Exit{}).Init(cg.OP_ifgt, code)
		case ast.ExpressionTypeLe:
			exit = (&cg.Exit{}).Init(cg.OP_ifle, code)
		case ast.ExpressionTypeLt:
			exit = (&cg.Exit{}).Init(cg.OP_iflt, code)
		case ast.ExpressionTypeGe:
			exit = (&cg.Exit{}).Init(cg.OP_ifge, code)
		case ast.ExpressionTypeEq:
			exit = (&cg.Exit{}).Init(cg.OP_ifeq, code)
		case ast.ExpressionTypeNe:
			exit = (&cg.Exit{}).Init(cg.OP_ifne, code)
		}
		code.Codes[code.CodeLength] = cg.OP_iconst_0
		code.CodeLength++
		falseExit := (&cg.Exit{}).Init(cg.OP_goto, code)
		writeExits([]*cg.Exit{exit}, code.CodeLength)
		context.MakeStackMap(code, state, code.CodeLength)
		code.Codes[code.CodeLength] = cg.OP_iconst_1
		code.CodeLength++
		writeExits([]*cg.Exit{falseExit}, code.CodeLength)
		state.pushStack(class, &ast.Type{
			Type: ast.VariableTypeBool,
		})
		context.MakeStackMap(code, state, code.CodeLength)
		defer state.popStack(1)
		return
	}

	if bin.Left.Value.IsPointer() && bin.Right.Value.IsPointer() { //
		stack := buildExpression.build(class, code, bin.Left, context, state)
		if stack > maxStack {
			maxStack = stack
		}
		state.pushStack(class, bin.Left.Value)
		stack = buildExpression.build(class, code, bin.Right, context, state)
		if t := stack + 1; t > maxStack {
			maxStack = t
		}
		state.popStack(1)
		var exit *cg.Exit
		if e.Type == ast.ExpressionTypeEq {
			exit = (&cg.Exit{}).Init(cg.OP_if_acmpeq, code)
		} else { // ne
			exit = (&cg.Exit{}).Init(cg.OP_if_acmpne, code)
		}
		code.Codes[code.CodeLength] = cg.OP_iconst_0
		code.CodeLength++
		falseExit := (&cg.Exit{}).Init(cg.OP_goto, code)
		writeExits([]*cg.Exit{exit}, code.CodeLength)
		context.MakeStackMap(code, state, code.CodeLength)
		code.Codes[code.CodeLength] = cg.OP_iconst_1
		code.CodeLength++
		writeExits([]*cg.Exit{falseExit}, code.CodeLength)
		state.pushStack(class, &ast.Type{
			Type: ast.VariableTypeBool,
		})
		context.MakeStackMap(code, state, code.CodeLength)
		defer state.popStack(1)
		return
	}
	return
}
