package jvm

import (
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

type Context struct {
	function          *ast.Function
	currentSoureFile  string
	currentLineNUmber int
	Defers            []*ast.Defer
}

func (c *Context) appendLimeNumberAndSourceFile(pos *ast.Pos, code *cg.AttributeCode, class *cg.ClassHighLevel) {
	if pos == nil {
		return
	}
	if pos.Filename != c.currentSoureFile {
		if class.SourceFiles == nil {
			class.SourceFiles = make(map[string]struct{})
		}
		class.SourceFiles[pos.Filename] = struct{}{}
		c.currentSoureFile = pos.Filename
		c.currentLineNUmber = pos.StartLine
		code.MKLineNumber(pos.StartLine)
		return
	}
	if c.currentLineNUmber != pos.StartLine {
		code.MKLineNumber(pos.StartLine)
		c.currentLineNUmber = pos.StartLine
	}
}
