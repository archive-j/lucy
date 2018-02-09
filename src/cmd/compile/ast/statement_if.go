package ast

import (
	"fmt"

	"github.com/756445638/lucy/src/cmd/compile/jvm/cg"
)

type StatementIF struct {
	BackPatchs []*cg.JumpBackPatch
	Condition  *Expression
	Block      *Block
	ElseBlock  *Block
	ElseIfList []*StatementElseIf
}

func (s *StatementIF) check(father *Block) []error {
	s.Block.inherite(father)
	errs := []error{}
	conditionType, es := s.Block.checkExpression(s.Condition)
	if errsNotEmpty(es) {
		errs = append(errs, es...)
	}
	if conditionType != nil {
		if conditionType.Typ != VARIABLE_TYPE_BOOL {
			errs = append(errs, fmt.Errorf("%s condition is not a bool expression",
				errMsgPrefix(s.Condition.Pos)))
		}
	}
	errs = append(errs, s.Block.check()...)
	if s.ElseIfList != nil && len(s.ElseIfList) > 0 {
		for _, v := range s.ElseIfList {
			v.Block.inherite(father)
			conditionType, es := v.Block.checkExpression(v.Condition)
			if errsNotEmpty(es) {
				errs = append(errs, es...)
			}
			if conditionType != nil {
				if conditionType.Typ != VARIABLE_TYPE_BOOL {
					errs = append(errs, fmt.Errorf("%s condition is not a bool expression",
						errMsgPrefix(s.Condition.Pos)))
				}
			}
			errs = append(errs, v.Block.check()...)
		}
	}
	if s.ElseBlock != nil {
		s.ElseBlock.inherite(father)
		errs = append(errs, s.ElseBlock.check()...)
	}
	return errs
}

type StatementElseIf struct {
	Condition *Expression
	Block     *Block
}