package ast

import (
	"fmt"
)

const (
	_ = iota
	STATEMENT_TYPE_EXPRESSION
	STATEMENT_TYPE_IF
	STATEMENT_TYPE_BLOCK
	STATEMENT_TYPE_FOR
	STATEMENT_TYPE_CONTINUE
	STATEMENT_TYPE_RETURN
	STATEMENT_TYPE_BREAK
	STATEMENT_TYPE_SWITCH
	STATEMENT_TYPE_SWITCH_TEMPLATE
	STATEMENT_TYPE_LABEL
	STATEMENT_TYPE_GOTO
	STATEMENT_TYPE_DEFER
	STATEMENT_TYPE_CLASS
	STATEMENT_TYPE_ENUM
	STATEMENT_TYPE_NOP
)

type Statement struct {
	Checked                 bool // if checked
	Pos                     *Position
	Type                    int
	StatementIf             *StatementIF
	Expression              *Expression
	StatementFor            *StatementFor
	StatementReturn         *StatementReturn
	StatementSwitch         *StatementSwitch
	StatementSwitchTemplate *StatementSwitchTemplate
	StatementBreak          *StatementBreak
	Block                   *Block
	StatementContinue       *StatementContinue
	StatementLabel          *StatementLabel
	StatementGoTo           *StatementGoTo
	Defer                   *StatementDefer
	Class                   *Class
	Enum                    *Enum
	/*
		this.super()
		special case
	*/
	IsCallFatherConstructionStatement bool
}

func (s *Statement) StatementName() string {
	switch s.Type {
	case STATEMENT_TYPE_EXPRESSION:
		return "expression statement"
	case STATEMENT_TYPE_IF:
		return "if statement"
	case STATEMENT_TYPE_FOR:
		return "for statement"
	case STATEMENT_TYPE_CONTINUE:
		return "continue statement"
	case STATEMENT_TYPE_BREAK:
		return "break statement"
	case STATEMENT_TYPE_SWITCH:
		return "switch statement"
	case STATEMENT_TYPE_LABEL:
		return "label statement"
	case STATEMENT_TYPE_GOTO:
		return "goto statement"
	case STATEMENT_TYPE_DEFER:
		return "defer statement"
	case STATEMENT_TYPE_BLOCK:
		return "block statement"
	case STATEMENT_TYPE_RETURN:
		return "return statement"
	case STATEMENT_TYPE_CLASS:
		return "class"
	case STATEMENT_TYPE_ENUM:
		return "enum"
	case STATEMENT_TYPE_NOP:
		return "nop"
	case STATEMENT_TYPE_SWITCH_TEMPLATE:
		return "switch template"
	}
	return ""
}

func (s *Statement) isVariableDefinition() bool {
	return s.Type == STATEMENT_TYPE_EXPRESSION &&
		(s.Expression.Type == EXPRESSION_TYPE_COLON_ASSIGN || s.Expression.Type == EXPRESSION_TYPE_VAR)
}

func (s *Statement) check(block *Block) []error { // block is father
	defer func() {
		s.Checked = true
	}()
	errs := []error{}
	switch s.Type {
	case STATEMENT_TYPE_EXPRESSION:
		return s.checkStatementExpression(block)
	case STATEMENT_TYPE_IF:
		return s.StatementIf.check(block)
	case STATEMENT_TYPE_FOR:
		return s.StatementFor.check(block)
	case STATEMENT_TYPE_SWITCH:
		return s.StatementSwitch.check(block)
	case STATEMENT_TYPE_BREAK:
		if block.InheritedAttribute.StatementFor == nil && block.InheritedAttribute.StatementSwitch == nil {
			return []error{fmt.Errorf("%s '%s' cannot in this scope", errMsgPrefix(s.Pos), s.StatementName())}
		} else {
			if block.InheritedAttribute.Defer != nil {
				return []error{fmt.Errorf("%s cannot has '%s' in 'defer'",
					errMsgPrefix(s.Pos), s.StatementName())}
			}
			if f, ok := block.InheritedAttribute.ForBreak.(*StatementFor); ok {
				s.StatementBreak.StatementFor = f
			} else {
				s.StatementBreak.StatementSwitch = block.InheritedAttribute.ForBreak.(*StatementSwitch)
			}
			s.StatementBreak.mkDefers(block)
		}
	case STATEMENT_TYPE_CONTINUE:
		if block.InheritedAttribute.StatementFor == nil {
			return []error{fmt.Errorf("%s '%s' can`t in this scope",
				errMsgPrefix(s.Pos), s.StatementName())}
		}
		if block.InheritedAttribute.Defer != nil {
			return []error{fmt.Errorf("%s cannot has '%s' in 'defer'",
				errMsgPrefix(s.Pos), s.StatementName())}
		}
		s.StatementContinue.StatementFor = block.InheritedAttribute.StatementFor
		s.StatementContinue.mkDefers(block)
	case STATEMENT_TYPE_RETURN:
		if block.InheritedAttribute.Defer != nil {
			return []error{fmt.Errorf("%s cannot has '%s' in 'defer'",
				errMsgPrefix(s.Pos), s.StatementName())}
		}
		es := s.StatementReturn.check(block)
		if len(s.StatementReturn.Defers) > 0 {
			block.InheritedAttribute.Function.MkAutoVarForReturnBecauseOfDefer()
		}
		return es
	case STATEMENT_TYPE_GOTO:
		err := s.checkStatementGoTo(block)
		if err != nil {
			return []error{err}
		}
	case STATEMENT_TYPE_DEFER:
		block.InheritedAttribute.Function.mkAutoVarForException()
		s.Defer.Block.inherit(block)
		s.Defer.Block.InheritedAttribute.Defer = s.Defer
		//s.Defer.allowCatch = block.IsFunctionBlock
		es := s.Defer.Block.checkStatements()
		block.Defers = append(block.Defers, s.Defer)
		return es
	case STATEMENT_TYPE_BLOCK:
		s.Block.inherit(block)
		return s.Block.checkStatements()
	case STATEMENT_TYPE_LABEL:
		// nothing to do
	case STATEMENT_TYPE_CLASS:
		err := block.Insert(s.Class.Name, s.Pos, s.Class)
		if err != nil {
			errs = append(errs, err)
		}
		return append(errs, s.Class.check(block)...)
	case STATEMENT_TYPE_ENUM:
		err := s.Enum.check()
		if err != nil {
			return []error{err}
		}
		err = block.Insert(s.Enum.Name, s.Pos, s.Enum)
		if err != nil {
			return []error{err}
		} else {
			return nil
		}
	case STATEMENT_TYPE_NOP:
		//nop , should be never execute to here
	case STATEMENT_TYPE_SWITCH_TEMPLATE:
		return s.StatementSwitchTemplate.check(block, s)
	}
	return nil
}

func (s *Statement) checkStatementExpression(b *Block) []error {
	errs := []error{}
	//
	if s.Expression.Type == EXPRESSION_TYPE_TYPE_ALIAS { // special case
		t := s.Expression.Data.(*ExpressionTypeAlias)
		err := t.Type.resolve(b)
		if err != nil {
			return []error{err}
		}
		err = b.Insert(t.Name, t.Pos, t.Type)
		if err != nil {
			return []error{err}
		}
		return nil
	}
	if s.Expression.canBeUsedAsStatement() {
		s.Expression.IsStatementExpression = true
		if s.Expression.Type == EXPRESSION_TYPE_FUNCTION {
			f := s.Expression.Data.(*Function)
			err := b.Insert(f.Name, f.Pos, f)
			if err != nil {
				errs = append(errs, err)
			}
			es := f.check(b)
			if errorsNotEmpty(es) {
				errs = append(errs, es...)
			}
			f.IsClosureFunction = f.Closure.NotEmpty(f)
			if f.IsClosureFunction {
				if b.ClosureFunctions == nil {
					b.ClosureFunctions = make(map[string]*Function)
				}
				b.ClosureFunctions[f.Name] = f
			}
			return errs
		}

	} else {
		err := fmt.Errorf("%s expression '%s' evaluate but not used",
			errMsgPrefix(s.Expression.Pos), s.Expression.OpName())
		errs = append(errs, err)
	}
	_, es := s.Expression.check(b)
	if errorsNotEmpty(es) {
		errs = append(errs, es...)
	}
	return errs
}
