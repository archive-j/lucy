package ast

import (
	"fmt"

	"gitee.com/yuyang-fine/lucy/src/cmd/compile/common"
	"gitee.com/yuyang-fine/lucy/src/cmd/compile/jvm/cg"
)

type Block struct {
	Exits []*cg.Exit // for switch template
	/*
		should analyse at ast stage
	*/
	WillNotExecuteToEnd             bool
	Defers                          []*StatementDefer
	IsGlobalVariableDefinitionBlock bool
	IsFunctionBlock                 bool // function block
	IsClassBlock                    bool // class block
	IsForBlock                      bool // for top block
	IsSwitchBlock                   bool // switch statement list block
	IsSwitchTemplateBlock           bool // template swtich statement list block
	Pos                             *Position
	EndPos                          *Position
	Outer                           *Block
	InheritedAttribute              InheritedAttribute
	Statements                      []*Statement
	Constants                       map[string]*Constant
	Functions                       map[string]*Function
	Classes                         map[string]*Class
	Enums                           map[string]*Enum
	EnumNames                       map[string]*EnumName
	Labels                          map[string]*StatementLabel
	TypeAliases                     map[string]*Type
	Variables                       map[string]*Variable
	ClosureFunctions                map[string]*Function //in "Functions" too
}

func (b *Block) HaveVariableDefinition() bool {
	return len(b.ClosureFunctions) > 0 || len(b.Variables) > 0
}

func (b *Block) NameExists(name string) (interface{}, bool) {
	if b.Functions != nil {
		if t, ok := b.Functions[name]; ok {
			return t, true
		}
	}
	if b.Variables != nil {
		if t, ok := b.Variables[name]; ok {
			return t, true
		}
	}
	if b.Constants != nil {
		if t, ok := b.Constants[name]; ok {
			return t, true
		}
	}
	if b.EnumNames != nil {
		if t, ok := b.EnumNames[name]; ok {
			return t, true
		}
	}
	if b.Classes != nil {
		if t, ok := b.Classes[name]; ok {
			return t, true
		}
	}
	if b.Enums != nil {
		if t, ok := b.Enums[name]; ok {
			return t, true
		}
	}
	if b.TypeAliases != nil {
		if t, ok := b.TypeAliases[name]; ok {
			return t, true
		}
	}
	if b.Labels != nil { // should be useless
		if t, ok := b.Labels[name]; ok {
			return t, true
		}
	}
	return nil, false
}

/*
	search label
*/
func (b *Block) searchLabel(name string) *StatementLabel {
	bb := b
	for {
		if bb.Labels != nil {
			if l, ok := bb.Labels[name]; ok {
				return l
			}
		}
		if bb.IsFunctionBlock {
			return nil
		}
		bb = bb.Outer
	}
	return nil
}

/*
	search type
*/
func (b *Block) searchType(name string) interface{} {
	bb := b
	for bb != nil {
		if bb.Classes != nil {
			if t, ok := bb.Classes[name]; ok {
				return t
			}
			if t, ok := bb.Enums[name]; ok {
				return t
			}
			if t, ok := bb.TypeAliases[name]; ok {
				return t
			}
		}
		bb = bb.Outer
	}
	return nil
}

/*
	search identifier
*/
func (b *Block) searchIdentifier(name string) (interface{}, error) {
	if b.Functions != nil {
		if t, ok := b.Functions[name]; ok {
			return t, nil
		}
	}
	if b.Variables != nil {
		if t, ok := b.Variables[name]; ok {
			return t, nil
		}
	}
	if b.Constants != nil {
		if t, ok := b.Constants[name]; ok {
			return t, nil
		}
	}
	if b.EnumNames != nil {
		if t, ok := b.EnumNames[name]; ok {
			return t, nil
		}
	}
	if b.Enums != nil {
		if t, ok := b.Enums[name]; ok {
			return t, nil
		}
	}
	if b.Classes != nil {
		if t, ok := b.Classes[name]; ok {
			return t, nil
		}
	}
	if b.TypeAliases != nil {
		if t, ok := b.TypeAliases[name]; ok {
			return t, nil
		}
	}
	// search closure
	if b.InheritedAttribute.Function != nil {
		v := b.InheritedAttribute.Function.Closure.Search(name)
		if v != nil {
			return v, nil
		}
	}
	if b.IsFunctionBlock &&
		len(b.InheritedAttribute.Function.parameterTypes) > 0 { // this is a template function
		return searchBuildIns(name), nil
	}
	if b.Outer == nil {
		return searchBuildIns(name), nil
	}
	t, err := b.Outer.searchIdentifier(name) // search by outer block
	if err != nil {
		return t, err
	}
	if t != nil { //
		if _, ok := t.(*Variable); ok && b.IsFunctionBlock &&
			len(b.InheritedAttribute.Function.parameterTypes) > 0 { // template function
			return nil, nil
		}
		if v, ok := t.(*Variable); ok && v.IsGlobal == false { // not a global variable
			if b.IsFunctionBlock &&
				b.InheritedAttribute.Function.IsGlobal == false { // 	b.InheritedAttribute.Function.IsGlobal == false  no need to check
				if v.Name == THIS {
					return nil, nil // capture this not allow
				}
				b.InheritedAttribute.Function.Closure.InsertVar(v)
			}
			//cannot search variable from class body
			if b.InheritedAttribute.Class != nil && b.IsClassBlock {
				return nil, nil //
			}
		}
		// if it is a function
		if f, ok := t.(*Function); ok && f.IsGlobal == false {
			if b.IsFunctionBlock {
				b.InheritedAttribute.Function.Closure.InsertFunction(f)
			}
			if b.IsClassBlock && f.IsClosureFunction {
				return nil, fmt.Errorf("trying to access closure function '%s' from class", name)
			}
		}
	}
	return t, nil
}

func (b *Block) inherit(father *Block) {
	if b == father {
		panic("inherit  from self")
	}
	b.InheritedAttribute = father.InheritedAttribute
	b.Outer = father
}

func (b *Block) checkUnUsedVariables() (es []error) {
	es = []error{}
	for _, v := range b.Variables {
		if v.Used ||
			v.IsGlobal ||
			v.IsFunctionParameter ||
			v.IsFunctionReturnVariable ||
			v.Name == THIS {
			continue
		}
		es = append(es, fmt.Errorf("%s variable '%s' has declared,but not used",
			errMsgPrefix(v.Pos), v.Name))
	}
	return es
}

func (b *Block) checkStatements() []error {
	errs := []error{}
	for k, s := range b.Statements {
		if s.isStaticFieldDefaultValue {
			//  compiler auto statement, no need to check
			continue
		}
		b.InheritedAttribute.StatementOffset = k
		errs = append(errs, s.check(b)...)
		if PackageBeenCompile.shouldStop(errs) {
			return errs
		}
	}
	if common.CompileFlags.DisableCheckUnUsedVariable == false {
		errs = append(errs, b.checkUnUsedVariables()...)
	}
	return errs
}

func (b *Block) checkConstants() []error {
	errs := make([]error, 0)
	for _, c := range b.Constants {
		if c.Name == NoNameIdentifier {
			err := fmt.Errorf("%s '%s' is not a valid name",
				errMsgPrefix(c.Pos), c.Name)
			errs = append(errs, err)
			delete(b.Constants, c.Name)
			continue
		}
		err := checkConst(b, c, &errs)
		if err != nil && c.Type == nil {
			delete(b.Constants, c.Name)
		}
	}
	return errs
}

func (b *Block) checkNameExist(name string, pos *Position) error {
	if b.Variables == nil {
		b.Variables = make(map[string]*Variable)
	}
	if v, ok := b.Variables[name]; ok {
		errMsg := fmt.Sprintf("%s name '%s' already declared as variable,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(v.Pos))
		return fmt.Errorf(errMsg)
	}
	if b.Classes == nil {
		b.Classes = make(map[string]*Class)
	}
	if c, ok := b.Classes[name]; ok {
		errMsg := fmt.Sprintf("%s name '%s' already declared as class,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(c.Pos))
		return fmt.Errorf(errMsg)
	}
	if b.Functions == nil {
		b.Functions = make(map[string]*Function)
	}
	if f, ok := b.Functions[name]; ok {
		errMsg := fmt.Sprintf("%s name '%s' already declared as function,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(f.Pos))
		return fmt.Errorf(errMsg)
	}
	if b.Constants == nil {
		b.Constants = make(map[string]*Constant)
	}
	if c, ok := b.Constants[name]; ok {
		errMsg := fmt.Sprintf("%s name '%s' already declared as const,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(c.Pos))
		return fmt.Errorf(errMsg)
	}
	if b.EnumNames == nil {
		b.EnumNames = make(map[string]*EnumName)
	}
	if en, ok := b.EnumNames[name]; ok {
		errMsg := fmt.Sprintf("%s name '%s' already declared as enumName,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(en.Pos))
		return fmt.Errorf(errMsg)
	}
	if b.TypeAliases == nil {
		b.TypeAliases = make(map[string]*Type)
	}
	if t, ok := b.TypeAliases[name]; ok {
		errMsg := fmt.Sprintf("%s name '%s' already declared as enumName,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(t.Pos))
		return fmt.Errorf(errMsg)
	}
	if b.Enums == nil {
		b.Enums = make(map[string]*Enum)
	}
	if e, ok := b.Enums[name]; ok {
		errMsg := fmt.Sprintf("%s name %s already declared as enum,first declared at:\n",
			errMsgPrefix(pos), name)
		errMsg += fmt.Sprintf("\t%s", errMsgPrefix(e.Pos))
		return fmt.Errorf(errMsg)
	}
	return nil
}

func (b *Block) Insert(name string, pos *Position, d interface{}) error {
	//fmt.Println(name, pos)
	// global var Insert into block
	if v, ok := d.(*Variable); ok && b.InheritedAttribute.Function.isGlobalVariableDefinition {
		b := PackageBeenCompile.Block
		err := b.checkNameExist(name, pos)
		if err != nil {
			return err
		}
		b.Variables[name] = v
		v.IsGlobal = true // it`s global
		return nil
	}
	if name == "" {
		return fmt.Errorf("%s name is null string", errMsgPrefix(pos))
	}
	if name == THIS {
		return fmt.Errorf("%s '%s' already been taken", errMsgPrefix(pos), THIS)
	}
	if name == "_" {
		return fmt.Errorf("%s '%s' is not a valid name", errMsgPrefix(pos), name)
	}
	if isMagicIdentifier(name) {
		return fmt.Errorf("%s '%s' is not a magic identifier", errMsgPrefix(pos), name)
	}
	// name exists in buildIn, not allow
	if lucyBuildInPackage != nil {
		if _, exits := lucyBuildInPackage.Block.NameExists(name); exits {
			return fmt.Errorf("%s '%s' is buildin", errMsgPrefix(pos), name)
		}
	}
	if label, ok := d.(*StatementLabel); ok && label != nil {
		if b.Labels == nil {
			b.Labels = make(map[string]*StatementLabel)
		}
		if l, ok := b.Labels[name]; ok {
			errMsg := fmt.Sprintf("%s name '%s' already declared as enumName,first declared at:",
				errMsgPrefix(pos), name)
			errMsg += fmt.Sprintf("\t%s", errMsgPrefix(l.Statement.Pos))
			return fmt.Errorf(errMsg)
		}
		b.Labels[name] = label
		return nil
	}
	err := b.checkNameExist(name, pos)
	if err != nil {
		return err
	}
	switch d.(type) {
	case *Class:
		b.Classes[name] = d.(*Class)
	case *Function:
		t := d.(*Function)
		if buildInFunctionsMap[t.Name] != nil {
			return fmt.Errorf("%s function named '%s' is buildin",
				errMsgPrefix(pos), name)
		}
		b.Functions[name] = t
	case *Constant:
		b.Constants[name] = d.(*Constant)
	case *Variable:
		t := d.(*Variable)
		b.Variables[name] = t
	case *Enum:
		e := d.(*Enum)
		b.Enums[name] = e
		for _, v := range e.Enums {
			err := b.Insert(v.Name, v.Pos, v)
			if err != nil {
				return err
			}
		}
	case *EnumName:
		b.EnumNames[name] = d.(*EnumName)
	case *Type:
		b.TypeAliases[name] = d.(*Type)
	default:
		panic(d) // panic d
	}
	return nil
}
