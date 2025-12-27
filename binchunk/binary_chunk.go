// LuaLight/binchunk/binary_chunk.go
package binchunk

//lua字节码的二进制结构

//header的一些常量
const (
	LUA_SIGNATURE    = "\x1bLua"            // Lua字节码的“魔数”
	LUAC_VERSION     = 0x53                 // Lua 5.3版本标识（0x53=83，对应5.3）
	LUAC_FORMAT      = 0                    // 字节码格式版本（固定0）
	LUAC_DATA        = "\x19\x93\r\n\x1a\n" // 格式校验用的固定字节
	CINT_SIZE        = 4                    // C语言int类型的字节数（32位）
	CSIZET_SIZE      = 8                    // C语言size_t类型的字节数（64位）
	INSTRUCTION_SIZE = 4                    // Lua字节码指令的字节数（每条指令4字节）
	LUA_INTEGER_SIZE = 8                    // Lua整数类型的字节数（64位）
	LUA_NUMBER_SIZE  = 8                    // Lua浮点数类型的字节数（64位）
	LUAC_INT         = 0x5678               // 校验用的测试整数
	LUAC_NUM         = 370.5                // 校验用的测试浮点数
)

//常量表
const (
	TAG_NIL       = 0x00 //空
	TAG_BOOLEAN   = 0x01 //字节0、1
	TAG_NUMBER    = 0x03 //Lua浮点数
	TAG_INTEGER   = 0x13 //Lua整数
	TAG_SHORT_STR = 0x04 //短字符串
	TAG_LONG_STR  = 0x14 //长字符串
)

//chunk结构
type binaryChunk struct {
	header                  //头部
	sizeUpvalues byte       //主函数upvalue环境数量
	mainFunc     *Prototype //主函数类型
}

//头
type header struct {
	signature       [4]byte // 魔法数，固定"\x1bLua"
	version         byte    // Lua版本，5.3=0x53
	format          byte    // 格式版本，固定0
	luacData        [6]byte // 校验字节，固定"\x19\x93\r\n\x1a\n"
	cintSize        byte    // C int字节数，64位=4
	sizetSize       byte    // C size_t字节数，64位=8
	instructionSize byte    // 指令字节数，固定4
	luaIntegerSize  byte    // Lua整数字节数，固定8
	luaNumberSize   byte    // Lua浮点数字节数，固定8
	luacInt         int64   // 测试整数，固定0x5678
	luacNum         float64 // 测试浮点数，固定370.5
}

//一些函数原型
type Prototype struct {
	Source          string        // 源文件名
	LineDefined     uint32        // 函数开始行号
	LastLineDefined uint32        // 函数结束行号
	NumParams       byte          // 固定参数个数（主函数是0）
	IsVararg        byte          // 是否是可变参数函数（主函数是0）Varag
	MaxStackSize    byte          // 函数运行需要的最大栈空间
	Code            []uint32      // 字节码指令
	Constants       []interface{} // 常量池
	Upvalues        []Upvalue     // Upvalue列表
	Protos          []*Prototype  // 函数内的子函数（没有就是空）
	LineInfo        []uint32      // 行号映射（指令对应源码的哪一行）
	LocVars         []LocVar      // 局部变量列表（没有就是空）
	UpvalueNames    []string      // Upvalue的名字
}

//Upvalue环境表
type Upvalue struct {
	Instack byte // 是否在栈中：1=是，0=否
	Idx     byte // 栈索引/Upvalue数组索引
}

//局部变量表
type LocVar struct {
	VarName string // 局部变量名
	StartPC uint32 // 变量生效起始指令位置
	EndPC   uint32 // 变量失效结束指令位置
}

// 解析chunk
func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()         //校验头部
	reader.readByte()            //跳过Upvalue数量
	return reader.readProto(" ") //读取函数原型
}
