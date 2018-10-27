package ast

import (
	"errors"
	"fmt"
	"path/filepath"
)

func (c *Class) check(father *Block) []error {
	c.Block.inherit(father)
	c.Block.InheritedAttribute.Class = c
	errs := c.checkPhase1()
	errs = append(errs, c.checkPhase2()...)
	return errs
}

func (c *Class) checkPhase1() []error {
	c.mkDefaultConstruction()
	errs := c.Block.checkConstants()
	err := c.resolveFather()
	if err != nil {
		errs = append(errs, err)
	} else {
		err = c.checkIfClassHierarchyErr()
		if err != nil {
			errs = append(errs, err)
		}
	}
	errs = append(errs, c.checkModifierOk()...)
	errs = append(errs, c.resolveFieldsAndMethodsType()...)
	return errs
}

func (c *Class) checkPhase2() []error {
	errs := []error{}
	if c.Block.InheritedAttribute.ClassAndFunctionNames == "" {
		c.Block.InheritedAttribute.ClassAndFunctionNames = filepath.Base(c.Name)
	} else {
		c.Block.InheritedAttribute.ClassAndFunctionNames += "$" + filepath.Base(c.Name)
	}
	errs = append(errs, c.checkFields()...)
	if PackageBeenCompile.shouldStop(errs) {
		return errs
	}
	c.mkClassInitMethod()
	for name, ms := range c.Methods {
		if c.Fields != nil && c.Fields[name] != nil {
			f := c.Fields[name]
			if f.Pos.Line < ms[0].Function.Pos.Line {
				errMsg := fmt.Sprintf("%s method named '%s' already declared as field,at:\n",
					errMsgPrefix(ms[0].Function.Pos), name)
				errMsg += fmt.Sprintf("\t%s", errMsgPrefix(f.Pos))
				errs = append(errs, errors.New(errMsg))
			} else {
				errMsg := fmt.Sprintf("%s field named '%s' already declared as method,at:\n",
					errMsgPrefix(f.Pos), name)
				errMsg += fmt.Sprintf("\t%s", errMsgPrefix(ms[0].Function.Pos))
				errs = append(errs, errors.New(errMsg))
			}
			continue
		}
		if len(ms) > 1 {
			errMsg := fmt.Sprintf("%s class method named '%s' has declared %d times,which are:\n",
				errMsgPrefix(ms[0].Function.Pos),
				ms[0].Function.Name, len(ms))
			for _, v := range ms {
				errMsg += fmt.Sprintf("\t%s\n", errMsgPrefix(v.Function.Pos))
			}
			errs = append(errs, errors.New(errMsg))
		}
	}
	errs = append(errs, c.checkMethods()...)
	if PackageBeenCompile.shouldStop(errs) {
		return errs
	}
	errs = append(errs, c.checkIfOverrideFinalMethod()...)
	errs = append(errs, c.resolveInterfaces()...)
	if c.IsInterface() {
		errs = append(errs, c.checkOverrideInterfaceMethod()...)
	}
	if c.IsAbstract() {
		errs = append(errs, c.checkOverrideAbstractMethod()...)
	}
	errs = append(errs, c.suitableForInterfaces()...)
	if c.SuperClass != nil {
		errs = append(errs, c.suitableSubClassForAbstract(c.SuperClass)...)
	}
	return errs
}

func (c *Class) suitableSubClassForAbstract(super *Class) []error {
	errs := []error{}
	if super.Name != JavaRootClass {
		err := super.loadSuperClass(c.Pos)
		if err != nil {
			errs = append(errs, err)
			return errs
		}
		if super.SuperClass == nil {
			return errs
		}
		length := len(errs)
		errs = append(errs, c.suitableSubClassForAbstract(super.SuperClass)...)
		if len(errs) > length {
			return errs
		}
	}
	if super.IsAbstract() {
		for _, v := range super.Methods {
			m := v[0]
			if m.IsAbstract() == false {
				continue
			}
			var nameMatch *ClassMethod
			implementation := c.implementMethod(c.Pos, m, &nameMatch, false, &errs)
			if implementation != nil {
				if err := m.implementationMethodIsOk(c.Pos, implementation); err != nil {
					errs = append(errs, err)
				}
			} else {
				pos := c.Pos
				if nameMatch != nil && nameMatch.Function.Pos != nil {
					pos = nameMatch.Function.Pos
				}
				if nameMatch != nil {
					errMsg := fmt.Sprintf("%s method is suitable for abstract super class\n", errMsgPrefix(pos))
					errMsg += fmt.Sprintf("\t have %s\n", nameMatch.Function.readableMsg())
					errMsg += fmt.Sprintf("\t want %s\n", m.Function.readableMsg())
					errs = append(errs, errors.New(errMsg))
				} else {
					errs = append(errs,
						fmt.Errorf("%s missing implementation method '%s' define on abstract class '%s'",
							pos.ErrMsgPrefix(), m.Function.readableMsg(), super.Name))
				}
			}
		}
	}
	return errs
}

func (c *Class) interfaceMethodExists(name string) *Class {
	if c.IsInterface() == false {
		panic("not a interface")
	}
	if c.Methods != nil && len(c.Methods[name]) > 0 {
		return c
	}
	for _, v := range c.Interfaces {
		if v.interfaceMethodExists(name) != nil {
			return v
		}
	}
	return nil
}

func (c *Class) abstractMethodExists(pos *Pos, name string) (*Class, error) {
	if c.IsAbstract() {
		if c.Methods != nil && len(c.Methods[name]) > 0 {
			method := c.Methods[name][0]
			if method.IsAbstract() {
				return c, nil
			}
		}
	}
	if c.Name == JavaRootClass {
		return nil, nil
	}
	err := c.loadSuperClass(pos)
	if err != nil {
		return nil, err
	}
	if c.SuperClass == nil {
		return nil, nil
	}
	return c.SuperClass.abstractMethodExists(pos, name)
}

func (c *Class) checkOverrideAbstractMethod() []error {
	errs := []error{}
	err := c.loadSuperClass(c.Pos)
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	if c.SuperClass == nil {
		return errs
	}
	for _, v := range c.Methods {
		m := v[0]
		name := m.Function.Name
		if m.IsAbstract() == false {
			continue
		}
		exist, err := c.SuperClass.abstractMethodExists(m.Function.Pos, name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if exist != nil {
			errs = append(errs, fmt.Errorf("%s method '%s' override '%s'",
				errMsgPrefix(v[0].Function.Pos), name, exist.Name))
		}
	}
	return errs
}

func (c *Class) checkOverrideInterfaceMethod() []error {
	errs := []error{}
	for name, v := range c.Methods {
		var exist *Class
		for _, vv := range c.Interfaces {
			exist = vv.interfaceMethodExists(name)
			if exist != nil {
				break
			}
		}
		if exist != nil {
			errs = append(errs, fmt.Errorf("%s method '%s' override '%s'",
				errMsgPrefix(v[0].Function.Pos), name, exist.Name))
		}
	}
	return errs
}

func (c *Class) checkIfClassHierarchyErr() error {
	m := make(map[string]struct{})
	arr := []string{}
	is := false
	class := c
	pos := c.Pos
	if err := c.loadSuperClass(pos); err != nil {
		return err
	}
	if c.SuperClass == nil {
		return nil
	}

	if c.SuperClass.IsFinal() {
		return fmt.Errorf("%s class name '%s' have super class  named '%s' that is final",
			c.Pos.ErrMsgPrefix(), c.Name, c.SuperClass.Name)
	}
	for class.Name != JavaRootClass {
		_, ok := m[class.Name]
		if ok {
			arr = append(arr, class.Name)
			is = true
			break
		}
		m[class.Name] = struct{}{}
		arr = append(arr, class.Name)
		err := class.loadSuperClass(pos)
		if err != nil {
			return err
		}
		if c.SuperClass == nil {
			return nil
		}
		class = class.SuperClass
	}
	if is == false {
		return nil
	}
	errMsg := fmt.Sprintf("%s class named '%s' detects a circularity in class hierarchy",
		c.Pos.ErrMsgPrefix(), c.Name)
	tab := "\t"
	index := len(arr) - 1
	for index >= 0 {
		errMsg += tab + arr[index] + "\n"
		tab += " "
		index--
	}
	return fmt.Errorf(errMsg)
}

func (c *Class) checkIfOverrideFinalMethod() []error {
	errs := []error{}
	if c.SuperClass != nil {
		for name, v := range c.Methods {
			if name == SpecialMethodInit {
				continue
			}
			if len(v) == 0 {
				continue
			}
			if len(c.SuperClass.Methods[name]) == 0 {
				// this class not found at super
				continue
			}
			m := v[0]
			for _, v := range c.SuperClass.Methods[name] {
				if v.IsFinal() == false {
					continue
				}
				f1 := &Type{
					Type:         VariableTypeFunction,
					FunctionType: &m.Function.Type,
				}
				f2 := &Type{
					Type:         VariableTypeFunction,
					FunctionType: &v.Function.Type,
				}
				if f1.Equal(f2) {
					errs = append(errs, fmt.Errorf("%s override final method",
						errMsgPrefix(m.Function.Pos)))
				}
			}
		}
	}
	return errs
}

func (c *Class) suitableForInterfaces() []error {
	errs := []error{}
	if c.IsInterface() {
		return errs
	}
	for _, i := range c.Interfaces {
		errs = append(errs, c.suitableForInterface(i)...)
	}
	return errs
}

func (c *Class) suitableForInterface(inter *Class) []error {
	errs := []error{}
	err := inter.loadSelf(c.Pos)
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	for _, v := range inter.Methods {
		m := v[0]
		var nameMatch *ClassMethod
		implementation := c.implementMethod(c.Pos, m, &nameMatch, false, &errs)
		if implementation != nil {
			if err := m.implementationMethodIsOk(c.Pos, implementation); err != nil {
				errs = append(errs, err)
			}
		} else {
			pos := c.Pos
			if nameMatch != nil && nameMatch.Function.Pos != nil {
				pos = nameMatch.Function.Pos
			}
			errs = append(errs, fmt.Errorf("%s missing implementation method '%s' define on interface '%s'",
				pos.ErrMsgPrefix(), m.Function.readableMsg(), inter.Name))
		}
	}
	for _, v := range inter.Interfaces {
		es := c.suitableForInterface(v)
		errs = append(errs, es...)
	}
	return errs
}

func (c *Class) checkFields() []error {
	errs := []error{}
	if c.IsInterface() {
		for _, v := range c.Fields {
			errs = append(errs, fmt.Errorf("%s interface '%s' expect no field named '%s'",
				errMsgPrefix(v.Pos), c.Name, v.Name))
		}
		return errs
	}
	staticFieldAssignStatements := []*Statement{}
	for _, v := range c.Fields {
		if v.DefaultValueExpression != nil {
			assignment, es := v.DefaultValueExpression.
				checkSingleValueContextExpression(&c.Methods[SpecialMethodInit][0].Function.Block)
			errs = append(errs, es...)
			if assignment == nil {
				continue
			}
			if v.Type.assignAble(&errs, assignment) == false {
				errs = append(errs, fmt.Errorf("%s cannot assign '%s' as '%s' for default value",
					errMsgPrefix(v.Pos), assignment.TypeString(), v.Type.TypeString()))
				continue
			}
			if assignment.Type == VariableTypeNull {
				errs = append(errs, fmt.Errorf("%s pointer types default value is '%s' already",
					v.Pos.ErrMsgPrefix(), assignment.TypeString()))
				continue
			}
			if v.IsStatic() &&
				v.DefaultValueExpression.isLiteral() {
				v.DefaultValue = v.DefaultValueExpression.Data
				continue
			}
			if v.IsStatic() == false {
				// nothing to do
				continue
			}
			bin := &ExpressionBinary{}
			bin.Right = &Expression{
				Type: ExpressionTypeList,
				Op:   "list",
				Data: []*Expression{v.DefaultValueExpression},
			}
			{
				selection := &ExpressionSelection{}
				selection.Expression = &Expression{}
				selection.Expression.Op = "selection"
				selection.Expression.Value = &Type{
					Type:  VariableTypeClass,
					Class: c,
				}
				selection.Name = v.Name
				selection.Field = v
				left := &Expression{
					Type: ExpressionTypeSelection,
					Data: selection,
					Op:   "selection",
				}
				left.Value = v.Type
				bin.Left = &Expression{
					Type: ExpressionTypeList,
					Data: []*Expression{left},
				}
			}
			e := &Expression{
				Type: ExpressionTypeAssign,
				Data: bin,
				IsStatementExpression: true,
				Op: "assign",
			}
			staticFieldAssignStatements = append(staticFieldAssignStatements, &Statement{
				Type:                      StatementTypeExpression,
				Expression:                e,
				isStaticFieldDefaultValue: true,
			})
		}
	}
	if len(staticFieldAssignStatements) > 0 {
		b := &Block{}
		b.Statements = staticFieldAssignStatements
		if c.StaticBlocks != nil {
			c.StaticBlocks = append([]*Block{b}, c.StaticBlocks...)
		} else {
			c.StaticBlocks = []*Block{b}
		}
	}
	return errs
}

func (c *Class) checkMethods() []error {
	errs := []error{}
	if c.IsInterface() {
		return errs
	}
	for name, methods := range c.Methods {
		for _, method := range methods {
			errs = append(errs, method.checkModifierOk()...)
			if method.IsAbstract() {
				//nothing
			} else {
				if c.IsInterface() {
					errs = append(errs, fmt.Errorf("%s interface method cannot have implementation",
						errMsgPrefix(method.Function.Pos)))
					continue
				}
				errs = append(errs, method.Function.checkReturnVarExpression()...)
				isConstruction := name == SpecialMethodInit
				if isConstruction {
					if method.IsFirstStatementCallFatherConstruction() == false {
						errs = append(errs, fmt.Errorf("%s construction method should call father construction method first",
							errMsgPrefix(method.Function.Pos)))
					}
				}
				if isConstruction && method.Function.Type.VoidReturn() == false {
					errs = append(errs, fmt.Errorf("%s construction method expect no return values",
						errMsgPrefix(method.Function.Type.ParameterList[0].Pos)))
				}
				method.Function.Block.InheritedAttribute.IsConstructionMethod = isConstruction
				method.Function.checkBlock(&errs)
			}
		}
	}
	return errs
}

func (c *Class) checkModifierOk() []error {
	errs := []error{}
	if c.IsInterface() && c.IsFinal() {
		errs = append(errs, fmt.Errorf("%s interface '%s' cannot be final",
			errMsgPrefix(c.FinalPos), c.Name))
	}
	if c.IsAbstract() && c.IsFinal() {
		errs = append(errs, fmt.Errorf("%s abstract class '%s' cannot be final",
			errMsgPrefix(c.FinalPos), c.Name))
	}
	return errs
}
