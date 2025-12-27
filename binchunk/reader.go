// LuaLight/binchunk/reader.go
package binchunk

import (
	"encoding/binary"
	"math"
)

type reader struct {
	data []byte
}

// 从字节流取一个字节
func (self *reader) readByte() byte {
	b := self.data[0]
	self.data = self.data[1:]
	return b
}

// 使用小端从字节流中读取一个cint储存类型 4字节
func (self *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(self.data)
	self.data = self.data[4:]
	return i
}

// 使用小端从字节流中读取一个size_t储存类型 8字节
func (self *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(self.data)
	self.data = self.data[8:]
	return i
}

// 从字节流中读lua整数
func (self *reader) readLuaInteger() int64 {
	return int64(self.readUint64())
}

// 从字节流中读浮点数8字节
func (self *reader) readLuaNumber() float64 {
	return math.Float64frombits(self.readUint64())
}

// 从字节流中读字符串
func (self *reader) readString() string {
	size := uint(self.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(self.readUint64())
	}
	bytes := self.readBytes(size - 1)
	return string(bytes)
}

// 从字节流中读取n个字节
func (self *reader) readBytes(n uint) []byte {
	bytes := self.data[:n]
	self.data = self.data[n:]
	return bytes
}

// checkHeader 读取字节码文件头并校验所有字段，不匹配则直接panic
func (self *reader) checkHeader() {
	// 校验魔数（是否为Lua预编译字节码）
	if string(self.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
		// 校验Lua版本（如5.3=0x53）
	} else if self.readByte() != LUAC_VERSION {
		panic("version mismatch!")
		// 校验字节码格式版本
	} else if self.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
		// 校验固定校验字节序列
	} else if string(self.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
		// 校验C int类型字节数
	} else if self.readByte() != CINT_SIZE {
		panic("int size mismatch!")
		// 校验C size_t类型字节数
	} else if self.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
		// 校验字节码指令字节数
	} else if self.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
		// 校验Lua整数类型字节数
	} else if self.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
		// 校验Lua浮点类型字节数
	} else if self.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
		// 校验整数解析字节序
	} else if self.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch!")
		// 校验浮点数解析格式
	} else if self.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}

// readProto 读取单个Lua函数原型（Chunk），parentSource为父函数源文件名
func (self *reader) readProto(parentSource string) *Prototype {
	// 读取源文件名，空则继承父函数的源文件（子函数场景）
	source := self.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,                  // 函数对应的源文件名
		LineDefined:     self.readUint32(),       // 函数定义起始行号
		LastLineDefined: self.readUint32(),       // 函数定义结束行号
		NumParams:       self.readByte(),         // 函数固定参数个数
		IsVararg:        self.readByte(),         // 是否为可变参数函数（1=是，0=否）
		MaxStackSize:    self.readByte(),         // 函数运行所需最大栈槽数
		Code:            self.readCode(),         // 字节码指令序列
		Constants:       self.readConstants(),    // 常量池
		Upvalues:        self.readUpvalues(),     // Upvalue列表
		Protos:          self.readProtos(source), // 子函数原型列表
		LineInfo:        self.readLineInfo(),     // 指令行号映射（调试用）
		LocVars:         self.readLocVars(),      // 局部变量表
		UpvalueNames:    self.readUpvalueNames(), // Upvalue名称列表
	}
}

// 从字节流中读指令表
func (self *reader) readCode() []uint32 {
	code := make([]uint32, self.readUint32())
	for i := range code {
		code[i] = self.readUint32()
	}
	return code
}

// 从字节流中读常量表
func (self *reader) readConstants() []interface{} {
	constants := make([]interface{}, self.readUint32())
	for i := range constants {
		constants[i] = self.readConstant()
	}
	return constants
}

// readConstant 读取单个Lua常量，返回对应类型的值（nil/布尔/整数/浮点数/字符串）
func (self *reader) readConstant() interface{} {
	// 根据常量类型标签读取对应值
	switch self.readByte() {
	case TAG_NIL: // 空值
		return nil
	case TAG_BOOLEAN: // 布尔值（1=true，0=false）
		return self.readByte() != 0
	case TAG_INTEGER: // 64位整数
		return self.readLuaInteger()
	case TAG_NUMBER: // 64位浮点数
		return self.readLuaNumber()
	case TAG_LONG_STR: // 长字符串
		return self.readString()
	case TAG_SHORT_STR: // 短字符串
		return self.readString()
	default: // 未知常量类型，判定字节码损坏
		panic("corrupted!")
	}
}

// 从字节流中读Upvalue表
func (self *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, self.readUint32())
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: self.readByte(), // 是否在栈中：1=是，0=否
			Idx:     self.readByte(), // 栈索引/Upvalue数组索引
		}
	}
	return upvalues
}

// 从字节流中读子函数原型表
func (self *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, self.readUint32())
	for i := range protos {
		protos[i] = self.readProto(parentSource)
	}
	return protos
}

// 从字节流中读行号表
func (self *reader) readLineInfo() []uint32 {
	lineInfo := make([]uint32, self.readUint32())
	for i := range lineInfo {
		lineInfo[i] = self.readUint32()
	}
	return lineInfo
}

// 从字节流中读取局部变量表
func (self *reader) readLocVars() []LocVar {
	LocVars := make([]LocVar, self.readUint32())
	for i := range LocVars {
		LocVars[i] = LocVar{
			VarName: self.readString(),
			StartPC: self.readUint32(),
			EndPC:   self.readUint32(),
		}
	}
	return LocVars
}

// 从字节流中读取Upvalue名列表
func (self *reader) readUpvalueNames() []string {
	names := make([]string, self.readUint32())
	for i := range names {
		names[i] = self.readString()
	}
	return names
}
