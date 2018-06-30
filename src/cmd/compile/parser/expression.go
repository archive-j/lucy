package parser

import (
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/lex"
)

type ExpressionParser struct {
	parser *Parser
}

func (expressionParser *ExpressionParser) Next() {
	expressionParser.parser.Next()
}

func (expressionParser *ExpressionParser) parseExpressions() ([]*ast.Expression, error) {
	es := []*ast.Expression{}
	for expressionParser.parser.token.Type != lex.TokenEof {
		e, err := expressionParser.parseExpression(false)
		if err != nil {
			return es, err
		}
		if e.Type == ast.ExpressionTypeList {
			es = append(es, e.Data.([]*ast.Expression)...)
		} else {
			es = append(es, e)
		}
		if expressionParser.parser.token.Type != lex.TokenComma {
			break
		}
		// == ,
		expressionParser.Next() // skip ,
	}
	return es, nil
}

//parse assign expression
func (expressionParser *ExpressionParser) parseExpression(statementLevel bool) (*ast.Expression, error) {
	left, err := expressionParser.parseTernaryExpression() //
	if err != nil {
		return nil, err
	}
	for expressionParser.parser.token.Type == lex.TokenComma && statementLevel { // read more
		expressionParser.Next()                                 //  skip comma
		left2, err := expressionParser.parseTernaryExpression() //
		if err != nil {
			return nil, err
		}
		if left.Type == ast.ExpressionTypeList {
			left.Data = append(left.Data.([]*ast.Expression), left2)
		} else {
			newExpression := &ast.Expression{}
			newExpression.Type = ast.ExpressionTypeList
			newExpression.Pos = left.Pos
			list := []*ast.Expression{left, left2}
			newExpression.Data = list
			left = newExpression
		}
	}
	mustBeOneExpression := func(left *ast.Expression) {
		if left.Type == ast.ExpressionTypeList {
			es := left.Data.([]*ast.Expression)
			left = es[0]
			if len(es) > 1 {
				expressionParser.parser.errs = append(expressionParser.parser.errs,
					fmt.Errorf("%s expect one expression on left",
						expressionParser.parser.errorMsgPrefix(es[1].Pos)))
			}
		}
	}
	mkExpression := func(typ int, multi bool) (*ast.Expression, error) {
		pos := expressionParser.parser.mkPos()
		expressionParser.Next() // skip = :=
		result := &ast.Expression{}
		result.Type = typ
		bin := &ast.ExpressionBinary{}
		result.Data = bin
		bin.Left = left
		result.Pos = pos
		if multi {
			es, err := expressionParser.parseExpressions()
			if err != nil {
				return result, err
			}
			bin.Right = &ast.Expression{}
			bin.Right.Type = ast.ExpressionTypeList
			bin.Right.Data = es
		} else {
			bin.Right, err = expressionParser.parseExpression(false)
		}
		return result, err
	}
	// := += -= *= /= %=
	switch expressionParser.parser.token.Type {
	case lex.TokenAssign:
		return mkExpression(ast.ExpressionTypeAssign, true)
	case lex.TokenColonAssign:
		return mkExpression(ast.ExpressionTypeColonAssign, true)
	case lex.TokenAddAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypePlusAssign, false)
	case lex.TokenSubAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeMinusAssign, false)
	case lex.TokenMulAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeMulAssign, false)
	case lex.TokenDivAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeDivAssign, false)
	case lex.TokenModAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeModAssign, false)
	case lex.TokenLshAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeLshAssign, false)
	case lex.TokenRshAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeRshAssign, false)
	case lex.TokenAndAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeAndAssign, false)
	case lex.TokenOrAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeOrAssign, false)
	case lex.TokenXorAssign:
		mustBeOneExpression(left)
		return mkExpression(ast.ExpressionTypeXorAssign, false)

	}
	return left, nil
}

func (expressionParser *ExpressionParser) parseTypeConversionExpression() (*ast.Expression, error) {
	pos := expressionParser.parser.mkPos()
	t, err := expressionParser.parser.parseType()
	if err != nil {
		return nil, err
	}
	if expressionParser.parser.token.Type != lex.TokenLp {
		return nil, fmt.Errorf("%s not '(' after a type", expressionParser.parser.errorMsgPrefix())
	}
	expressionParser.Next() // skip (
	e, err := expressionParser.parseExpression(false)
	if err != nil {
		return nil, err
	}
	if expressionParser.parser.token.Type != lex.TokenRp {
		return nil, fmt.Errorf("%s '(' and ')' not match", expressionParser.parser.errorMsgPrefix())
	}
	expressionParser.Next() // skip )
	return &ast.Expression{
		Type: ast.ExpressionTypeCheckCast,
		Data: &ast.ExpressionTypeConversion{
			Type:       t,
			Expression: e,
		},
		Pos: pos,
	}, nil
}
