package cg

import (
	"fmt"
	"io"
)

const (
	ACC_CLASS_PUBLIC     uint16 = 0x0001 // 可以被包的类外访问。
	ACC_CLASS_FINAL      uint16 = 0x0010 //不允许有子类。
	ACC_CLASS_SUPER      uint16 = 0x0020 //当用到invokespecial指令时，需要特殊处理③的父类方法。
	ACC_CLASS_INTERFACE  uint16 = 0x0200 // 标识定义的是接口而不是类。
	ACC_CLASS_ABSTRACT   uint16 = 0x0400 //  不能被实例化。
	ACC_CLASS_SYNTHETIC  uint16 = 0x1000 //标识并非Java源码生成的代码。
	ACC_CLASS_ANNOTATION uint16 = 0x2000 // 标识注解类型
	ACC_CLASS_ENUM       uint16 = 0x4000 // 标识枚举类型
)

type Class struct {
	destination            io.Writer
	magic                  uint32 //0xCAFEBABE
	MinorVersion           uint16
	MajorVersion           uint16
	ConstPool              []*ConstPool
	AccessFlag             uint16
	ThisClass              uint16
	SuperClass             uint16
	Interfaces             []uint16
	Fields                 []*FieldInfo
	Methods                []*MethodInfo
	Attributes             []*AttributeInfo
	AttributeCompilerAuto  *AttributeCompilerAuto
	AttributeGroupedByName AttributeGroupedByName
	TypeAlias              []*AttributeLucyTypeAlias
	AttributeLucyEnum      *AttributeLucyEnum
	//const caches
	Utf8Consts               map[string]*ConstPool
	IntConsts                map[int32]*ConstPool
	LongConsts               map[int64]*ConstPool
	FloatConsts              map[float32]*ConstPool
	DoubleConsts             map[float64]*ConstPool
	ClassConsts              map[string]*ConstPool
	StringConsts             map[string]*ConstPool
	FieldRefConsts           map[CONSTANT_Fieldref_info_high_level]*ConstPool
	NameAndTypeConsts        map[CONSTANT_NameAndType_info_high_level]*ConstPool
	MethodrefConsts          map[CONSTANT_Methodref_info_high_level]*ConstPool
	InterfaceMethodrefConsts map[CONSTANT_InterfaceMethodref_info_high_level]*ConstPool
}

func (c *Class) InsertInterfaceMethodrefConst(n CONSTANT_InterfaceMethodref_info_high_level) uint16 {
	if c.InterfaceMethodrefConsts == nil {
		c.InterfaceMethodrefConsts = make(map[CONSTANT_InterfaceMethodref_info_high_level]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.InterfaceMethodrefConsts[n]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_InterfaceMethodref_info{
		classIndex: c.InsertClassConst(n.Class),
		nameAndTypeIndex: c.InsertNameAndType(CONSTANT_NameAndType_info_high_level{
			Name:       n.Method,
			Descriptor: n.Descriptor,
		}),
	}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.InterfaceMethodrefConsts[n] = info
	return info.selfindex
}

func (c *Class) InsertMethodrefConst(n CONSTANT_Methodref_info_high_level) uint16 {
	if c.MethodrefConsts == nil {
		c.MethodrefConsts = make(map[CONSTANT_Methodref_info_high_level]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.MethodrefConsts[n]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Methodref_info{
		classIndex: c.InsertClassConst(n.Class),
		nameAndTypeIndex: c.InsertNameAndType(CONSTANT_NameAndType_info_high_level{
			Name:       n.Method,
			Descriptor: n.Descriptor,
		}),
	}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.MethodrefConsts[n] = info
	return info.selfindex
}

func (c *Class) InsertNameAndType(n CONSTANT_NameAndType_info_high_level) uint16 {
	if c.NameAndTypeConsts == nil {
		c.NameAndTypeConsts = make(map[CONSTANT_NameAndType_info_high_level]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.NameAndTypeConsts[n]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_NameAndType_info{
		name:       c.insertUtf8Const(n.Name),
		descriptor: c.insertUtf8Const(n.Descriptor),
	}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.NameAndTypeConsts[n] = info
	return info.selfindex
}
func (c *Class) InsertFieldRefConst(f CONSTANT_Fieldref_info_high_level) uint16 {
	if c.FieldRefConsts == nil {
		c.FieldRefConsts = make(map[CONSTANT_Fieldref_info_high_level]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.FieldRefConsts[f]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Fieldref_info{
		classIndex: c.InsertClassConst(f.Class),
		nameAndTypeIndex: c.InsertNameAndType(CONSTANT_NameAndType_info_high_level{
			Name:       f.Field,
			Descriptor: f.Descriptor,
		}),
	}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.FieldRefConsts[f] = info
	return info.selfindex
}
func (c *Class) insertUtf8Const(s string) uint16 {
	if c.Utf8Consts == nil {
		c.Utf8Consts = make(map[string]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.Utf8Consts[s]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Utf8_info{uint16(len(s)), []byte(s)}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.Utf8Consts[s] = info
	return info.selfindex
}

func (c *Class) InsertIntConst(i int32) uint16 {
	if c.IntConsts == nil {
		c.IntConsts = make(map[int32]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.IntConsts[i]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Integer_info{i}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.IntConsts[i] = info
	return info.selfindex
}
func (c *Class) InsertLongConst(i int64) uint16 {
	if c.LongConsts == nil {
		c.LongConsts = make(map[int64]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.LongConsts[i]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Long_info{i}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info, nil)
	c.LongConsts[i] = info
	return info.selfindex
}

func (c *Class) InsertFloatConst(f float32) uint16 {
	if c.FloatConsts == nil {
		c.FloatConsts = make(map[float32]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.FloatConsts[f]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Float_info{f}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.FloatConsts[f] = info
	return info.selfindex
}

func (c *Class) InsertDoubleConst(f float64) uint16 {
	if c.DoubleConsts == nil {
		c.DoubleConsts = make(map[float64]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.DoubleConsts[f]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Double_info{f}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info, nil)
	c.DoubleConsts[f] = info
	return info.selfindex
}

func (c *Class) InsertClassConst(name string) uint16 {
	if c.ClassConsts == nil {
		c.ClassConsts = make(map[string]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.ClassConsts[name]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_Class_info{c.insertUtf8Const(name)}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.ClassConsts[name] = info
	return info.selfindex
}

func (c *Class) InsertStringConst(s string) uint16 {
	if c.StringConsts == nil {
		c.StringConsts = make(map[string]*ConstPool)
	}
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil}
	}
	if con, ok := c.StringConsts[s]; ok {
		return con.selfindex
	}
	info := (&CONSTANT_String_info{c.insertUtf8Const(s)}).ToConstPool()
	info.selfindex = c.constPoolUint16Length()
	c.ConstPool = append(c.ConstPool, info)
	c.StringConsts[s] = info
	return info.selfindex
}

func (high *ClassHighLevel) ToLow(jvmVersion int) *Class {
	high.Class.fromHighLevel(high, jvmVersion)
	return &high.Class
}

func (c *Class) fromHighLevel(high *ClassHighLevel, jvmVersion int) {
	c.MinorVersion = 0
	c.MajorVersion = uint16(jvmVersion)
	if len(c.ConstPool) == 0 {
		c.ConstPool = []*ConstPool{nil} // jvm const pool index begin at 1
	}
	c.AccessFlag = high.AccessFlags
	c.ThisClass = c.InsertClassConst(high.Name)
	c.SuperClass = c.InsertClassConst(high.SuperClass)
	for _, i := range high.Interfaces {
		inter := (&CONSTANT_Class_info{c.insertUtf8Const(i)}).ToConstPool()
		index := c.constPoolUint16Length()
		c.Interfaces = append(c.Interfaces, index)
		c.ConstPool = append(c.ConstPool, inter)
	}
	for _, f := range high.Fields {
		field := &FieldInfo{}
		field.AccessFlags = f.AccessFlags
		field.NameIndex = c.insertUtf8Const(f.Name)
		if f.AttributeConstantValue != nil {
			field.Attributes = append(field.Attributes, f.AttributeConstantValue.ToAttributeInfo(c))
		}
		field.DescriptorIndex = c.insertUtf8Const(f.Descriptor)
		if f.AttributeLucyFieldDescritor != nil {
			field.Attributes = append(field.Attributes, f.AttributeLucyFieldDescritor.ToAttributeInfo(c))
		}
		if f.AttributeLucyConst != nil {
			field.Attributes = append(field.Attributes, f.AttributeLucyConst.ToAttributeInfo(c))
		}
		c.Fields = append(c.Fields, field)
	}
	for _, ms := range high.Methods {
		for _, m := range ms {
			info := &MethodInfo{}
			info.AccessFlags = m.AccessFlags
			info.NameIndex = c.insertUtf8Const(m.Name)
			info.DescriptorIndex = c.insertUtf8Const(m.Descriptor)
			if m.Code != nil {
				info.Attributes = append(info.Attributes, m.Code.ToAttributeInfo(c))
			}

			if m.AttributeLucyMethodDescritor != nil {
				info.Attributes = append(info.Attributes, m.AttributeLucyMethodDescritor.ToAttributeInfo(c))
			}
			if m.AttributeLucyTriggerPackageInitMethod != nil {
				info.Attributes = append(info.Attributes, m.AttributeLucyTriggerPackageInitMethod.ToAttributeInfo(c))
			}
			if m.AttributeDefaultParameters != nil {
				info.Attributes = append(info.Attributes, m.AttributeDefaultParameters.ToAttributeInfo(c))
			}
			if m.AttributeMethodParameters != nil {
				t := m.AttributeMethodParameters.ToAttributeInfo(c)
				if t != nil {
					info.Attributes = append(info.Attributes, t)
				}
			}
			if m.AttributeCompilerAuto != nil {
				info.Attributes = append(info.Attributes, m.AttributeCompilerAuto.ToAttributeInfo(c))
			}
			if m.AttributeLucyReturnListNames != nil {
				t := m.AttributeLucyReturnListNames.ToAttributeInfo(c, ATTRIBUTE_NAME_LUCY_RETURNLIST_NAMES)
				if t != nil {
					info.Attributes = append(info.Attributes, t)
				}
			}
			c.Methods = append(c.Methods, info)
		}
	}
	//source file
	c.Attributes = append(c.Attributes, (&AttributeSourceFile{high.getSourceFile()}).ToAttributeInfo(c))
	if c.AttributeCompilerAuto != nil {
		c.Attributes = append(c.Attributes, c.AttributeCompilerAuto.ToAttributeInfo(c))
	}
	for _, v := range c.TypeAlias {
		c.Attributes = append(c.Attributes, v.ToAttributeInfo(c))
	}
	if c.AttributeLucyEnum != nil {
		c.Attributes = append(c.Attributes, c.AttributeLucyEnum.ToAttributeInfo(c))
	}
	for _, v := range high.TemplateFunctions {
		c.Attributes = append(c.Attributes, v.ToAttributeInfo(c))
	}
	c.ifConstPoolOverMaxSize()
	return
}

func (c *Class) constPoolUint16Length() uint16 {
	return uint16(len(c.ConstPool))
}
func (c *Class) ifConstPoolOverMaxSize() {
	if len(c.ConstPool) > CONSTANT_POOL_MAX_SIZE {
		panic(fmt.Sprintf("const pool max size is:%d", CONSTANT_POOL_MAX_SIZE))
	}
}
