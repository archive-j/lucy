package parser

import (
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/lex"
)

func (b *Block) parseIf() (i *ast.StatementIF, err error) {
	b.Next() // skip if

	var e *ast.Expression
	e, err = b.parser.Expression.parseExpression(false)
	if err != nil {
		b.parser.errs = append(b.parser.errs, err)
		b.consume(untils_lc)
		b.Next()
	}
	i = &ast.StatementIF{}
	i.Condition = e
	for b.parser.token.Type == lex.TOKEN_SEMICOLON {
		if i.Condition != nil {
			i.PreExpressions = append(i.PreExpressions, i.Condition)
		}
		b.Next() // skip ;
		i.Condition, err = b.parser.Expression.parseExpression(false)
		if err != nil {
			b.parser.errs = append(b.parser.errs, err)
			b.consume(untils_lc)
			b.Next()
		}

	}

	if b.parser.token.Type != lex.TOKEN_LC {
		err = fmt.Errorf("%s missing '{' after a expression,but '%s'", b.parser.errorMsgPrefix(), b.parser.token.Desp)
		b.parser.errs = append(b.parser.errs, err)
		b.consume(untils_lc) // consume and next
		b.Next()
	}

	b.Next() //skip {
	b.parseStatementList(&i.Block, false)
	if b.parser.token.Type != lex.TOKEN_RC {
		b.parser.errs = append(b.parser.errs, fmt.Errorf("%s expect '}', but '%s'"))
		b.consume(untils_rc)
	}
	b.Next() // skip }
	if b.parser.token.Type == lex.TOKEN_ELSEIF {
		es, err := b.parseElseIfList()
		if err != nil {
			return i, err
		}
		i.ElseIfList = es
	}
	if b.parser.token.Type == lex.TOKEN_ELSE {
		b.Next()
		if b.parser.token.Type != lex.TOKEN_LC {
			err = fmt.Errorf("%s missing '{' after else", b.parser.errorMsgPrefix())
			b.parser.errs = append(b.parser.errs, err)
			return i, err
		}
		i.ElseBlock = &ast.Block{}
		b.Next()
		b.parseStatementList(i.ElseBlock, false)
		if b.parser.token.Type != lex.TOKEN_RC {
			err = fmt.Errorf("%s expect '}', but '%s'")
			b.consume(untils_rc)
		}
		b.Next()

	}
	return i, err
}

func (b *Block) parseElseIfList() (es []*ast.StatementElseIf, err error) {
	es = []*ast.StatementElseIf{}
	var e *ast.Expression
	for (b.parser.token.Type == lex.TOKEN_ELSEIF) && b.parser.token.Type != lex.TOKEN_EOF {
		b.Next() // skip elseif token
		e, err = b.parser.Expression.parseExpression(false)
		if err != nil {
			b.parser.errs = append(b.parser.errs, err)
			return es, err
		}
		if b.parser.token.Type != lex.TOKEN_LC {
			err = fmt.Errorf("%s not { after a expression,but %s", b.parser.errorMsgPrefix(), b.parser.token.Desp)
			b.parser.errs = append(b.parser.errs)
			return es, err
		}
		block := &ast.Block{}
		b.Next()
		b.parseStatementList(block, false)

		es = append(es, &ast.StatementElseIf{
			Condition: e,
			Block:     block,
		})
	}
	return es, err
}
