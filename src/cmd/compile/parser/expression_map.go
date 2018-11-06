package parser

import (
	"fmt"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/lex"
)

func (ep *ExpressionParser) parseMapExpression() (*ast.Expression, error) {
	var typ *ast.Type
	var err error
	if ep.parser.token.Type == lex.TokenMap {
		typ, err = ep.parser.parseType()
		if err != nil {
			return nil, err
		}
		ep.parser.ifTokenIsLfThenSkip()
	}
	if ep.parser.token.Type != lex.TokenLc {
		err := fmt.Errorf("%s expect '{',but '%s'", ep.parser.errMsgPrefix(), ep.parser.token.Description)
		ep.parser.errs = append(ep.parser.errs, err)
		return nil, err
	}
	ep.Next(lfNotToken) // skip {
	ret := &ast.Expression{
		Type: ast.ExpressionTypeMap,
		Op:   "mapLiteral",
	}
	m := &ast.ExpressionMap{}
	m.Type = typ
	ret.Data = m
	for ep.parser.token.Type != lex.TokenEof &&
		ep.parser.token.Type != lex.TokenRc {
		// key
		k, err := ep.parseExpression(false)
		if err != nil {
			return ret, err
		}
		ep.parser.unExpectNewLineAndSkip()
		// arrow
		if ep.parser.token.Type != lex.TokenArrow {
			err := fmt.Errorf("%s expect '->',but '%s'",
				ep.parser.errMsgPrefix(), ep.parser.token.Description)
			ep.parser.errs = append(ep.parser.errs, err)
			return ret, err
		}
		ep.Next(lfNotToken) // skip ->
		// value
		v, err := ep.parseExpression(false)
		if err != nil {
			return ret, err
		}
		m.KeyValuePairs = append(m.KeyValuePairs, &ast.ExpressionKV{
			Key:   k,
			Value: v,
		})
		if ep.parser.token.Type == lex.TokenComma {
			// read next  key value pair
			ep.Next(lfNotToken)
		} else {
			break
		}
	}
	ep.parser.ifTokenIsLfThenSkip()
	if ep.parser.token.Type != lex.TokenRc {
		err := fmt.Errorf("%s expect '}',but '%s'",
			ep.parser.errMsgPrefix(), ep.parser.token.Description)
		ep.parser.errs = append(ep.parser.errs, err)
		ep.parser.consume(untilRc)
	}
	ret.Pos = ep.parser.mkPos()
	ep.Next(lfIsToken) // skip }
	return ret, nil
}
