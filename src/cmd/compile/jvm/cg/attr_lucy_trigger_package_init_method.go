package cg

type AttributeLucyTriggerPackageInitMethod struct {
}

func (a *AttributeLucyTriggerPackageInitMethod) ToAttributeInfo(class *Class) *AttributeInfo {
	ret := &AttributeInfo{}
	ret.NameIndex = class.insertUtfConst(ATTRIBUTE_NAME_LUCY_TRIGGER_PACKAGE_INIT)
	return ret
}