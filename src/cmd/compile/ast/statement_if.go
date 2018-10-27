package ast

import (
	"fmt"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

type StatementIf struct {
	PrefixExpressions   []*Expression
	Condition           *Expression
	Pos                 *Pos
	initExpressionBlock Block
	Block               Block
	ElseIfList          []*StatementElseIf
	Else                *Block
	Exits               []*cg.Exit
}

func (s *StatementIf) check(father *Block) []error {
	s.initExpressionBlock.inherit(father)
	errs := []error{}
	for _, v := range s.PrefixExpressions {
		v.IsStatementExpression = true
		_, es := v.check(&s.initExpressionBlock)
		errs = append(errs, es...)
		if err := v.canBeUsedAsStatement(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.Condition != nil {
		conditionType, es := s.Condition.checkSingleValueContextExpression(&s.initExpressionBlock)
		errs = append(errs, es...)
		if conditionType != nil &&
			conditionType.Type != VariableTypeBool {
			errs = append(errs, fmt.Errorf("%s condition is not a bool expression",
				s.Condition.Pos.ErrMsgPrefix()))
		}
		if err := s.Condition.canBeUsedAsCondition(); err != nil {
			errs = append(errs, err)
		}
	}
	s.Block.inherit(&s.initExpressionBlock)
	errs = append(errs, s.Block.check()...)
	for _, v := range s.ElseIfList {
		v.Block.inherit(&s.initExpressionBlock)
		if v.Condition != nil {
			conditionType, es := v.Condition.checkSingleValueContextExpression(v.Block)
			errs = append(errs, es...)
			if err := v.Condition.canBeUsedAsCondition(); err != nil {
				errs = append(errs, err)
			}
			if conditionType != nil &&
				conditionType.Type != VariableTypeBool {
				errs = append(errs,
					fmt.Errorf("%s condition is not a bool expression",
						conditionType.Pos.ErrMsgPrefix()))
			}
			errs = append(errs, v.Block.check()...)
		}
	}
	if s.Else != nil {
		s.Else.inherit(&s.initExpressionBlock)
		errs = append(errs, s.Else.check()...)
	}
	return errs
}

type StatementElseIf struct {
	Condition *Expression
	Block     *Block
}
