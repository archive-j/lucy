package parser

import (
	"fmt"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/ast"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/lex"
)

//at least one name
func (parser *Parser) parseNameList() (names []*ast.NameWithPos, err error) {
	if parser.token.Type != lex.TokenIdentifier {
		err = fmt.Errorf("%s expect identifier,but '%s'",
			parser.errMsgPrefix(), parser.token.Description)
		parser.errs = append(parser.errs, err)
		return nil, err
	}
	names = []*ast.NameWithPos{}
	for parser.token.Type == lex.TokenIdentifier {
		names = append(names, &ast.NameWithPos{
			Name: parser.token.Data.(string),
			Pos:  parser.mkPos(),
		})
		parser.Next(lfIsToken)
		if parser.token.Type != lex.TokenComma {
			// not a ,
			break
		} else {
			parser.Next(lfNotToken) // skip comma
			if parser.token.Type != lex.TokenIdentifier {
				err = fmt.Errorf("%s not a 'identifier' after a comma,but '%s'",
					parser.errMsgPrefix(), parser.token.Description)
				parser.errs = append(parser.errs, err)
				return names, err
			}
		}
	}
	return
}
