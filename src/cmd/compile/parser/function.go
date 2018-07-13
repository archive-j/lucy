package parser

import (
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/lex"
)

type FunctionParser struct {
	parser *Parser
}

func (functionParser *FunctionParser) Next() {
	functionParser.parser.Next()
}

func (functionParser *FunctionParser) consume(until map[int]bool) {
	functionParser.parser.consume(until)
}

func (functionParser *FunctionParser) parse(needName bool) (f *ast.Function, err error) {
	f = &ast.Function{}
	f.Pos = functionParser.parser.mkPos()
	var offset int
	offset = functionParser.parser.token.Offset
	functionParser.Next() // skip fn key word
	if needName && functionParser.parser.token.Type != lex.TokenIdentifier {
		err := fmt.Errorf("%s expect function name,but '%s'",
			functionParser.parser.errorMsgPrefix(), functionParser.parser.token.Description)
		functionParser.parser.errs = append(functionParser.parser.errs, err)
		functionParser.parser.consume(untilLp)
	}
	if functionParser.parser.token.Type == lex.TokenIdentifier {
		f.Name = functionParser.parser.token.Data.(string)
		functionParser.Next()
	}
	f.Type, err = functionParser.parser.parseFunctionType()
	if err != nil {
		//functionParser.consume(untilLc)
		return nil, err
	}
	if functionParser.parser.token.Type != lex.TokenLc {
		err = fmt.Errorf("%s except '{' but '%s'",
			functionParser.parser.errorMsgPrefix(), functionParser.parser.token.Description)
		functionParser.parser.errs = append(functionParser.parser.errs, err)
		functionParser.consume(untilLc)
	}
	f.Block.IsFunctionBlock = true
	functionParser.Next() // skip {
	functionParser.parser.BlockParser.parseStatementList(&f.Block, false)
	if functionParser.parser.token.Type != lex.TokenRc {
		err = fmt.Errorf("%s expect '}', but '%s'",
			functionParser.parser.errorMsgPrefix(), functionParser.parser.token.Description)
		functionParser.parser.errs = append(functionParser.parser.errs, err)
		functionParser.consume(untilRc)
	} else {
		f.SourceCodes =
			functionParser.parser.
				bs[offset : functionParser.parser.token.Offset+1]
		functionParser.Next()
	}
	return f, err
}
