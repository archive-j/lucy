package cg

const (
	ConstantPoolMaxSize                  = 65536
	ClassMagicNumber              uint32 = 0xcafebabe
	AttributeNameSourceFile              = "SourceFile"
	AttributeNameConstValue              = "ConstantValue"
	AttributeNameStackMap                = "StackMapTable"
	AttributeNameSignature               = "Signature"
	AttributeNameMethodParameters        = "MethodParameters"
	// lucy attribute
	AttributeNameLucyFieldDescriptor    = "LucyFieldDescriptor"
	AttributeNameLucyMethodDescriptor   = "LucyMethodDescriptor"
	AttributeNameLucyTriggerPackageInit = "LucyTriggerPackageInitMethod"
	AttributeNameLucyDefaultParameters  = "LucyDefaultParameters"
	AttributeNameLucyTypeAlias          = "LucyTypeAlias"
	AttributeNameLucyEnum               = "LucyEnum"
	AttributeNameLucyConst              = "LucyConst" // indicate a package const
	AttributeNameLucyReturnListNames    = "LucyReturnListName"
	AttributeNameLucyTemplateFunction   = "LucyTemplateFunction"
	AttributeNameLucyComment            = "LucyComment"
	AttributeNameLucyClassConst         = "LucyClassConst"
)
