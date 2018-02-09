package jvm

import (
	"github.com/756445638/lucy/src/cmd/compile/ast"
	"github.com/756445638/lucy/src/cmd/compile/jvm/cg"
)

func (m *MakeClass) buildStatement(class *cg.ClassHighLevel, code *cg.AttributeCode, s *ast.Statement, context *Context) (maxstack uint16) {
	switch s.Typ {
	case ast.STATEMENT_TYPE_EXPRESSION:
		//code.MKLineNumber(s.Expression.Pos.StartLine)
		var es []*cg.JumpBackPatch
		maxstack, _ = m.MakeExpression.build(class, code, s.Expression, context)
		backPatchEs(es, code.CodeLength)
	case ast.STATEMENT_TYPE_IF:
		maxstack = m.buildIfStatement(class, code, s.StatementIf, context)
		backPatchEs(s.StatementIf.BackPatchs, code.CodeLength)
	case ast.STATEMENT_TYPE_BLOCK:
		m.buildBlock(class, code, s.Block, context)
	case ast.STATEMENT_TYPE_FOR:
		maxstack = m.buildForStatement(class, code, s.StatementFor, context)
		backPatchEs(s.StatementFor.BackPatchs, code.CodeLength)
		backPatchEs(s.StatementFor.ContinueBackPatchs, s.StatementFor.ContinueOPOffset)
	case ast.STATEMENT_TYPE_CONTINUE:
		s.StatementContinue.StatementFor.ContinueBackPatchs = append(s.StatementContinue.StatementFor.ContinueBackPatchs,
			(&cg.JumpBackPatch{}).FromCode(cg.OP_goto, code))
	case ast.STATEMENT_TYPE_BREAK:
		code.Codes[code.CodeLength] = cg.OP_goto
		b := (&cg.JumpBackPatch{}).FromCode(cg.OP_goto, code)
		if s.StatementBreak.StatementFor != nil {
			s.StatementBreak.StatementFor.BackPatchs = append(s.StatementBreak.StatementFor.BackPatchs, b)
		} else { // switch
			s.StatementBreak.StatementSwitch.BackPatchs = append(s.StatementBreak.StatementSwitch.BackPatchs, b)
		}
	case ast.STATEMENT_TYPE_RETURN:
		maxstack = m.buildReturnStatement(class, code, s.StatementReturn, context)
	case ast.STATEMENT_TYPE_SWITCH:
		maxstack = m.buildSwitchStatement(class, code, s.StatementSwitch, context)
		backPatchEs(s.StatementSwitch.BackPatchs, code.CodeLength)
	case ast.STATEMENT_TYPE_SKIP: // skip this block
		panic("no skip")
	case ast.STATEMENT_TYPE_GOTO:
		b := (&cg.JumpBackPatch{}).FromCode(cg.OP_goto, code)
		s.StatementGoto.StatementLable.BackPatches = append(s.StatementGoto.StatementLable.BackPatches, b)
	case ast.STATEMENT_TYPE_LABLE:
		backPatchEs(s.StatmentLable.BackPatches, code.CodeLength) // back patch
	}
	return
}

func (m *MakeClass) buildSwitchStatement(class *cg.ClassHighLevel, code *cg.AttributeCode, s *ast.StatementSwitch, context *Context) (maxstack uint16) {
	return
}
