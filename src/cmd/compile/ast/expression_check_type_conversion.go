package ast

import (
	"fmt"
)

func (e *Expression) checkTypeConversionExpression(block *Block, errs *[]error) *Type {
	conversion := e.Data.(*ExpressionTypeConversion)
	on, es := conversion.Expression.checkSingleValueContextExpression(block)
	if esNotEmpty(es) {
		*errs = append(*errs, es...)
	}
	if on == nil {
		return nil
	}
	err := conversion.Type.resolve(block)
	if err != nil {
		*errs = append(*errs, err)
		return nil
	}
	ret := conversion.Type.Clone()
	ret.Pos = e.Pos

	if on.IsNumber() && conversion.Type.IsNumber() {
		if conversion.Expression.IsLiteral() {
			conversion.Expression.convertNumberLiteralTo(conversion.Type.Type)
			//rewrite
			pos := e.Pos
			*e = *conversion.Expression
			e.Pos = pos // keep pos
		}
		return ret
	}

	// string([]byte)
	if conversion.Type.Type == VariableTypeString &&
		on.Type == VariableTypeArray && on.Array.Type == VariableTypeByte {
		return ret
	}
	// string(byte[])
	if conversion.Type.Type == VariableTypeString &&
		on.Type == VariableTypeJavaArray && on.Array.Type == VariableTypeByte {
		return ret
	}

	// []byte("hello world")
	if conversion.Type.Type == VariableTypeArray && conversion.Type.Array.Type == VariableTypeByte &&
		on.Type == VariableTypeString {
		return ret
	}
	// byte[]("hello world")
	if conversion.Type.Type == VariableTypeJavaArray && conversion.Type.Array.Type == VariableTypeByte &&
		on.Type == VariableTypeString {
		return ret
	}
	if conversion.Type.validForTypeAssertOrConversion() && on.IsPointer() {
		return ret
	}
	*errs = append(*errs, fmt.Errorf("%s cannot convert '%s' to '%s'",
		errMsgPrefix(e.Pos), on.TypeString(), conversion.Type.TypeString()))
	return ret
}
