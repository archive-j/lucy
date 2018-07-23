package ast

type Variable struct {
	IsBuildIn                bool
	LocalValOffset           uint16
	IsGlobal                 bool
	IsFunctionParameter      bool
	IsFunctionReturnVariable bool
	BeenCaptured             bool
	Used                     bool   // use as right value
	AccessFlags              uint16 // public private or protected
	Pos                      *Pos
	Expression               *Expression
	Name                     string
	Type                     *Type
	JvmDescriptor            string // jvm
}
