package jvm

import (
	"fmt"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

func (buildPackage *BuildPackage) buildBlock(
	class *cg.ClassHighLevel,
	code *cg.AttributeCode,
	b *ast.Block,
	context *Context,
	state *StackMapState) {
	notToHere := false
	for _, s := range b.Statements {
		if notToHere == true && s.Type == ast.StatementTypeLabel {
			notToHere = len(s.StatementLabel.Exits) == 0
			//continue compile block from this label statement
		}
		if notToHere {
			continue
		}
		if s.IsCallFatherConstructionStatement {
			// special case
			// no need to build
			// this statement is build before
			continue
		}
		maxStack := buildPackage.buildStatement(class, code, b, s, context, state)
		if maxStack > code.MaxStack {
			code.MaxStack = maxStack
		}

		if len(state.Stacks) > 0 {
			var ss []string
			for _, v := range state.Stacks {
				ss = append(ss, v.ToString())
			}
			fmt.Println("stacks:", ss)
			panic(fmt.Sprintf("stack is not empty:%d", len(state.Stacks)))
		}
		//unCondition goto
		if buildPackage.statementIsUnConditionJump(s) {
			notToHere = true
			continue
		}
		//block deadEnd
		if s.Type == ast.StatementTypeBlock {
			notToHere = s.Block.NotExecuteToLastStatement
			continue
		}
		if s.Type == ast.StatementTypeIf && s.StatementIf.Else != nil {
			t := s.StatementIf.Block.NotExecuteToLastStatement
			for _, v := range s.StatementIf.ElseIfList {
				t = t && v.Block.NotExecuteToLastStatement
			}
			t = t && s.StatementIf.Else.NotExecuteToLastStatement
			notToHere = t
			continue
		}
		if s.Type == ast.StatementTypeSwitch && s.StatementSwitch.Default != nil {
			t := s.StatementSwitch.Default.NotExecuteToLastStatement
			for _, v := range s.StatementSwitch.StatementSwitchCases {
				if v.Block != nil {
					t = t && v.Block.NotExecuteToLastStatement
				} else {
					//this will fallthrough
					t = false
					break
				}
			}
			t = t && s.StatementSwitch.Default.NotExecuteToLastStatement
			notToHere = t
			continue
		}
	}
	b.NotExecuteToLastStatement = notToHere
	if b.IsFunctionBlock == false &&
		len(b.Defers) > 0 &&
		b.NotExecuteToLastStatement == false {
		buildPackage.buildDefers(class, code, context, b.Defers, state)
	}
	return
}

func (buildPackage *BuildPackage) statementIsUnConditionJump(s *ast.Statement) bool {
	return s.Type == ast.StatementTypeReturn ||
		s.Type == ast.StatementTypeGoTo ||
		s.Type == ast.StatementTypeContinue ||
		s.Type == ast.StatementTypeBreak
}
