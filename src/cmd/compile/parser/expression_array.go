package parser

import (
	"fmt"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/lex"
)

// []int{1,2,3}
func (ep *ExpressionParser) parseArrayExpression() (*ast.Expression, error) {
	ep.parser.Next(lfIsToken) // skip [
	ep.parser.unExpectNewLineAndSkip()
	var err error
	if ep.parser.token.Type != lex.TokenRb {
		/*
			[1 ,2]
		*/
		arr := &ast.ExpressionArray{}
		arr.Expressions, err = ep.parseExpressions(lex.TokenRb)
		ep.parser.ifTokenIsLfThenSkip()
		if ep.parser.token.Type != lex.TokenRb {
			err = fmt.Errorf("%s '[' and ']' not match", ep.parser.errMsgPrefix())
			return nil, err
		}
		pos := ep.parser.mkPos()
		ep.Next(lfIsToken) // skip ]
		return &ast.Expression{
			Type: ast.ExpressionTypeArray,
			Data: arr,
			Pos:  pos,
			Op:   "arrayLiteral",
		}, err
	}
	ep.Next(lfIsToken) // skip ]
	ep.parser.unExpectNewLineAndSkip()
	array, err := ep.parser.parseType()
	if err != nil {
		return nil, err
	}
	pos := ep.parser.mkPos()
	if ep.parser.token.Type == lex.TokenLp {
		/*
			[]byte("1111111111")
		*/
		ep.Next(lfNotToken) // skip (
		e, err := ep.parseExpression(false)
		if err != nil {
			return nil, err
		}
		if ep.parser.token.Type != lex.TokenRp {
			err = fmt.Errorf("%s '(' and  ')' not match",
				ep.parser.errMsgPrefix())
			ep.parser.errs = append(ep.parser.errs, err)
			return nil, err
		}
		ret := &ast.Expression{}
		ret.Op = "checkCast"
		ret.Pos = pos
		ret.Type = ast.ExpressionTypeCheckCast
		data := &ast.ExpressionTypeConversion{}
		data.Type = &ast.Type{}
		data.Type.Type = ast.VariableTypeArray
		data.Type.Pos = pos
		data.Type.Array = array
		data.Expression = e
		ret.Data = data
		ep.Next(lfIsToken) // skip )
		return ret, nil
	}
	ep.parser.unExpectNewLineAndSkip()
	arr := &ast.ExpressionArray{}
	if array != nil {
		arr.Type = &ast.Type{}
		arr.Type.Type = ast.VariableTypeArray
		arr.Type.Array = array
		arr.Type.Pos = array.Pos
	}
	/*
		[]int { 1, 2}
	*/

	arr.Expressions, err = ep.parseArrayValues()
	return &ast.Expression{
		Type: ast.ExpressionTypeArray,
		Data: arr,
		Pos:  pos,
		Op:   "arrayLiteral",
	}, err

}

//{1,2,3}  {{1,2,3},{456}}
func (ep *ExpressionParser) parseArrayValues() ([]*ast.Expression, error) {
	if ep.parser.token.Type != lex.TokenLc {
		err := fmt.Errorf("%s expect '{',but '%s'",
			ep.parser.errMsgPrefix(), ep.parser.token.Description)
		ep.parser.errs = append(ep.parser.errs, err)
		return nil, err
	}
	ep.Next(lfNotToken) // skip {
	es := []*ast.Expression{}
	for ep.parser.token.Type != lex.TokenEof &&
		ep.parser.token.Type != lex.TokenRc {
		if ep.parser.token.Type == lex.TokenComment ||
			ep.parser.token.Type == lex.TokenMultiLineComment {
			ep.Next(lfIsToken)
			continue
		}
		if ep.parser.token.Type == lex.TokenLc {
			ees, err := ep.parseArrayValues()
			if err != nil {
				return es, err
			}
			arrayExpression := &ast.Expression{
				Type: ast.ExpressionTypeArray,
				Pos:  ep.parser.mkPos(),
			}
			arrayExpression.Op = "arrayLiteral"
			data := ast.ExpressionArray{}
			data.Expressions = ees
			arrayExpression.Data = data
			es = append(es, arrayExpression)
		} else {
			e, err := ep.parseExpression(false)
			if e != nil {
				es = append(es, e)
			}
			if err != nil {
				return es, err
			}
		}
		if ep.parser.token.Type == lex.TokenComma {
			ep.Next(lfNotToken) // skip ,
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
	ep.Next(lfIsToken)
	return es, nil
}
