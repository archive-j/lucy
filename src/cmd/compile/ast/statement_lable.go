package ast

import (
	"errors"
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

type StatementLable struct {
	CodeOffsetGenerated bool
	CodeOffset          int
	Block               *Block
	Name                string
	BackPatches         []*cg.JumpBackPatch
	Statement           *Statement
}

func (s *StatementLable) Ready(from *Pos) error {
	ss := []*Statement{}
	for _, v := range s.Block.Statements {
		if v.StatmentLable == s {
			break
		}
		if v.isVariableDefinition() && v.Checked == false {
			ss = append(ss, v)
		}
	}
	if len(ss) == 0 {
		return nil
	}
	errmsg := fmt.Sprintf("%s cannot jump over variable definition:\n", errMsgPrefix(from))
	for _, v := range ss {
		errmsg += fmt.Sprintf("\t%s constains variable definition\n", errMsgPrefix(v.Pos))
	}
	return errors.New(errmsg)
}
