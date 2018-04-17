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
	STATEMENT_TYPE_SKIP // skip this block
	STATEMENT_TYPE_LABLE
	STATEMENT_TYPE_GOTO
	STATEMENT_TYPE_DEFER
)

type Statement struct {
	Checked           bool
	Pos               *Pos
	Typ               int
	StatementIf       *StatementIF
	Expression        *Expression // expression statment like a=123
	StatementFor      *StatementFor
	StatementReturn   *StatementReturn
	StatementSwitch   *StatementSwitch
	StatementBreak    *StatementBreak
	Block             *Block
	StatementContinue *StatementContinue
	StatmentLable     *StatementLable
	StatementGoto     *StatementGoto
	Defer             *Defer
}

func (s *Statement) statementName() string {
	switch s.Typ {
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
	case STATEMENT_TYPE_SKIP:
		return "skip statement"
	case STATEMENT_TYPE_LABLE:
		return "lable statement"
	case STATEMENT_TYPE_GOTO:
		return "goto statement"
	case STATEMENT_TYPE_DEFER:
		return "defer statement"
	case STATEMENT_TYPE_BLOCK:
		return "block statement"
	}
	return ""
}

func (s *Statement) isVariableDefinition() bool {
	return s.Typ == STATEMENT_TYPE_EXPRESSION &&
		(s.Expression.Typ == EXPRESSION_TYPE_COLON_ASSIGN || s.Expression.Typ == EXPRESSION_TYPE_VAR)
}

func (s *Statement) check(b *Block) []error { // b is father
	defer func() {
		s.Checked = true
	}()
	if b.Defers != nil && len(b.Defers) > 0 {
		b.InheritedAttribute.Defers = append(b.InheritedAttribute.Defers, b.Defers...)
		defer func() {
			b.InheritedAttribute.Defers = b.InheritedAttribute.Defers[0 : len(b.InheritedAttribute.Defers)-len(b.Defers)]
		}()
	}
	switch s.Typ {
	case STATEMENT_TYPE_EXPRESSION:
		return s.checkStatementExpression(b)
	case STATEMENT_TYPE_IF:
		return s.StatementIf.check(b)
	case STATEMENT_TYPE_FOR:
		return s.StatementFor.check(b)
	case STATEMENT_TYPE_SWITCH:
		return s.StatementSwitch.check(b)
	case STATEMENT_TYPE_BREAK:
		if b.InheritedAttribute.StatementFor == nil && b.InheritedAttribute.StatementSwitch == nil {
			return []error{fmt.Errorf("%s '%s' cannot in this scope", errMsgPrefix(s.Pos), s.statementName())}
		} else {
			if b.InheritedAttribute.StatementFor != nil && b.InheritedAttribute.Defer != nil {
				return []error{fmt.Errorf("%s cannot has '%s' in both 'defer' and 'for'",
					errMsgPrefix(s.Pos), s.statementName())}
			}
			s.StatementBreak = &StatementBreak{}
			if f, ok := b.InheritedAttribute.mostCloseForOrSwitchForBreak.(*StatementFor); ok {
				s.StatementBreak.StatementFor = f
			} else {
				s.StatementBreak.StatementSwitch = b.InheritedAttribute.mostCloseForOrSwitchForBreak.(*StatementSwitch)
			}
		}
	case STATEMENT_TYPE_CONTINUE:
		if b.InheritedAttribute.StatementFor == nil {
			return []error{fmt.Errorf("%s '%s' can`t in this scope",
				errMsgPrefix(s.Pos), s.statementName())}
		}
		if b.InheritedAttribute.StatementFor != nil && b.InheritedAttribute.Defer != nil {
			return []error{fmt.Errorf("%s cannot has '%s' in both 'defer' and 'for'",
				errMsgPrefix(s.Pos), s.statementName())}
		}
		if s.StatementContinue == nil {
			s.StatementContinue = &StatementContinue{}
		}
		if s.StatementContinue.StatementFor == nil { // for
			s.StatementContinue.StatementFor = b.InheritedAttribute.StatementFor
		}
	case STATEMENT_TYPE_RETURN:
		if b.InheritedAttribute.Defer != nil {
			return []error{fmt.Errorf("%s cannot has '%s' in 'defer'",
				errMsgPrefix(s.Pos), s.statementName())}
		}
		return s.StatementReturn.check(b)
	case STATEMENT_TYPE_GOTO:
		err := s.checkStatementGoto(b)
		if err != nil {
			return []error{err}
		}
	case STATEMENT_TYPE_DEFER:
		s.Defer.Block.inherite(b)
		return s.Defer.Block.check()
	case STATEMENT_TYPE_BLOCK:
		s.Block.inherite(b)
		return s.Block.check()
	case STATEMENT_TYPE_LABLE: // nothing to do
	case STATEMENT_TYPE_SKIP:
		if b.InheritedAttribute.Function.isPackageBlockFunction == false {
			return []error{fmt.Errorf("cannot have '%s' at this scope", s.statementName())}
		}
		return nil
	}

	return nil
}

func (s *Statement) checkStatementExpression(b *Block) []error {
	errs := []error{}
	//
	if s.Expression.Typ == EXPRESSION_TYPE_TYPE_ALIAS {
		t := s.Expression.Data.(*ExpressionTypeAlias)
		err := t.Typ.resolve(b)

		if err != nil {
			return []error{err}
		}

		err = b.insert(t.Name, t.Pos, t.Typ)
		if err != nil {
			return []error{err}
		}
		return nil
	}

	if s.Expression.canBeUsedAsStatementExpression() {
		if s.Expression.Typ == EXPRESSION_TYPE_FUNCTION {
			f := s.Expression.Data.(*Function)
			if f.Name == "" {
				err := fmt.Errorf("%s function must have a name",
					errMsgPrefix(s.Expression.Pos), s.Expression.OpName())
				errs = append(errs, err)
			} else {
				err := b.insert(f.Name, f.Pos, f)
				if err != nil {
					errs = append(errs, err)
				}
			}
			es := f.check(b)
			if errsNotEmpty(es) {
				errs = append(errs, es...)
			}
			f.IsClosureFunction = f.ClosureVars.NotEmpty(f)
			if f.IsClosureFunction {
				f.VarOffSetForClosure = b.InheritedAttribute.Function.VarOffset
				b.InheritedAttribute.Function.VarOffset++
				b.InheritedAttribute.Function.OffsetDestinations = append(b.InheritedAttribute.Function.OffsetDestinations, &f.VarOffSetForClosure)
			}
			return errs
		}
	} else {
		err := fmt.Errorf("%s expression '%s' evaluate but not used",
			errMsgPrefix(s.Expression.Pos), s.Expression.OpName())
		errs = append(errs, err)
	}
	s.Expression.IsStatementExpression = true
	_, es := b.checkExpression_(s.Expression)
	if errsNotEmpty(es) {
		errs = append(errs, es...)
	}
	return errs
}
