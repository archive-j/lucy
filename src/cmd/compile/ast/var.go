package ast

import (
	"fmt"
	"regexp"
)

type LoadImport interface {
	LoadName(resourceName string) (interface{}, error)
}

const (
	MAIN_FUNCTION_NAME       = "main"
	THIS                     = "this"
	NO_NAME_IDENTIFIER       = "_"
	LUCY_ROOT_CLASS          = "lucy/lang/Lucy"
	JAVA_ROOT_CLASS          = "java/lang/Object"
	DEFAULT_EXCEPTION_CLASS  = "java/lang/Exception"
	JAVA_THROWABLE_CLASS     = "java/lang/Throwable"
	JAVA_STRING_CLASS        = "java/lang/String"
	SUPER_FIELD_NAME         = "super"
	CONSTRUCTION_METHOD_NAME = "<init>"
)

var (
	packageAliasReg      *regexp.Regexp
	ImportsLoader        LoadImport
	PackageBeenCompile   Package
	buildInFunctionsMap  = make(map[string]*Function)
	lucyBuildInPackage   *Package
	ParseFunctionHandler func(bs []byte, pos *Pos) (*Function, []error)
	javaStringClass      *Class
)

func init() {
	var err error
	packageAliasReg, err = regexp.Compile(`^[a-zA-Z][[a-zA-Z1-9\_]+$`)
	if err != nil {
		panic(err)
	}
}

func loadJavaStringClass(pos *Pos) error {
	if javaStringClass != nil {
		return nil
	}
	c, err := ImportsLoader.LoadName(JAVA_STRING_CLASS)
	if err != nil {
		return err
	}
	if cc, ok := c.(*Class); ok && cc != nil {
		javaStringClass = cc
		return nil
	} else {
		return fmt.Errorf("%s '%s' is not class",
			errMsgPrefix(pos), JAVA_STRING_CLASS)
	}
}
