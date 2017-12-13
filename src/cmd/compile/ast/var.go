package ast

import (
	"fmt"
	"regexp"
)

type PackageLoader interface {
	LoadPackage(name string) (*Package, error)
}

var (
	small_float          = 0.0001
	negative_small_float = -small_float
	Nodes                *[]*Node //
	packageAliasReg      *regexp.Regexp
	PackageLoad          PackageLoader
)

type NameWithPos struct {
	Name string
	Pos  *Pos
}

func init() {
	var err error
	packageAliasReg, err = regexp.Compile(`^[a-zA-Z][[a-zA-Z1-9\_]+$`)
	if err != nil {
		panic(err)
	}
}

func notFoundError(pos *Pos, typ, name string) error {
	return fmt.Errorf("%s %s named %s not found", errMsgPrefix(pos), typ, name)
}

func errMsgPrefix(pos *Pos) string {
	return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.StartLine, pos.StartColumn)
}

func errsNotEmpty(errs []error) bool {
	return errs != nil && len(errs) > 0
}
func checkEnum(enums []*Enum) []error {
	ret := make([]error, 0)
	for _, v := range enums {
		if len(v.Names) == 0 {
			continue
		}
		is, typ, value, err := v.Init.getConstValue()
		if err != nil || is == false || typ != EXPRESSION_TYPE_INT {
			ret = append(ret, fmt.Errorf("enum type must inited by integer"))
			continue
		}
		for k, vv := range v.Names {
			vv.Value = int64(k) + value.(int64)
		}
	}
	return ret
}
