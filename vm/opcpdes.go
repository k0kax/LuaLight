// LuaLight/vm/opcodes.go
package vm

// 指令编码模式（Lua 5.3字节码指令的4种编码格式）
const (
	IABC  = iota // 3个操作数：A + B + C（基础格式）
	IABx         // 2个操作数：A + Bx（Bx为无符号整数）
	IAsBx        // 2个操作数：A + sBx（sBx为有符号整数）
	IAx          // 1个操作数：Ax（Ax为无符号整数）
)

// 操作数类型（标记操作数的使用规则）
const (
	OpArgN = iota // 无该操作数
	OpArgU        // 操作数为无符号值
	OpArgR        // 操作数为寄存器索引
	OpArgK        // 操作数为常量/寄存器索引
)

// 操作码（Lua 5.3核心指令，对应虚拟机执行的具体操作）
const (
	OP_MOVE     = iota // 寄存器间值传递
	OP_LOADK           // 加载常量到寄存器
	OP_LOADKX          // 扩展加载常量（配合EXTRAARG）
	OP_LOADBOOL        // 加载布尔值到寄存器
	OP_LOADNIL         // 加载nil值到连续寄存器
	OP_GETUPVAL        // 获取Upvalue值到寄存器
	OP_GETTABUP        // 获取Upvalue表中元素到寄存器
	OP_GETTABLE        // 获取表中元素到寄存器
	OP_SETTABUP        // 设置Upvalue表中元素
	OP_SETUPVAL        // 设置Upvalue值
	OP_SETTABLE        // 设置表中元素
	OP_NEWTABLE        // 创建新表
	OP_SELF            // 准备对象方法调用（self）
	OP_ADD             // 加法运算
	OP_SUB             // 减法运算
	OP_MUL             // 乘法运算
	OP_MOD             // 取模运算
	OP_POW             // 幂运算
	OP_DIV             // 除法运算
	OP_IDIV            // 整数除法运算
	OP_BAND            // 按位与
	OP_BOR             // 按位或
	OP_BXOR            // 按位异或
	OP_SHL             // 按位左移
	OP_SHR             // 按位右移
	OP_UNM             // 取负（一元运算）
	OP_BNOT            // 按位取反
	OP_NOT             // 逻辑取反
	OP_LEN             // 获取字符串/表长度
	OP_CONCAT          // 拼接多个值为字符串
	OP_JMP             // 无条件跳转
	OP_EQ              // 相等比较
	OP_LT              // 小于比较
	OP_LE              // 小于等于比较
	OP_TEST            // 条件测试（无赋值）
	OP_TESTSET         // 条件测试并赋值
	OP_CALL            // 函数调用
	OP_TAILCALL        // 尾调用（优化版函数调用）
	OP_RETURN          // 函数返回
	OP_FORLOOP         // for循环迭代
	OP_FORPREP         // for循环初始化
	OP_TFORCALL        // 泛型for调用迭代器
	OP_TFORLOOP        // 泛型for循环
	OP_SETLIST         // 设置表的列表部分元素
	OP_CLOSURE         // 创建函数闭包
	OP_VARARG          // 处理可变参数
	OP_EXTRAARG        // 扩展操作数（配合LOADKX等指令）
)

// opcode 指令元信息结构，描述每条操作码的执行规则
type opcode struct {
	testFlag byte   // 是否为条件测试指令（1=是）
	setAFlag byte   // 是否设置寄存器A的值（1=是）
	argBMode byte   // 操作数B的类型（OpArgN/OpArgU/OpArgR/OpArgK）
	argCMode byte   // 操作数C的类型（同上）
	opMode   byte   // 指令编码模式（IABC/IABx/IAsBx/IAx）
	name     string // 指令名称（用于调试/反汇编）
}

// opcodes 指令元信息数组，按操作码常量顺序定义，描述每条指令的执行规则
// 字段顺序：testFlag | setAFlag | argBMode | argCMode | opMode | name | 执行逻辑注释
var opcodes = []opcode{
	opcode{0, 1, OpArgR, OpArgN, IABC, "MOVE    "},  // R(A) := R(B)          寄存器B的值赋值给寄存器A
	opcode{0, 1, OpArgK, OpArgN, IABx, "LOADK   "},  // R(A) := Kst(Bx)       加载常量池Bx位置的常量到寄存器A
	opcode{0, 1, OpArgN, OpArgN, IABx, "LOADKX  "},  // R(A) := Kst(extra arg) 扩展加载常量（配合EXTRAARG指令）
	opcode{0, 1, OpArgU, OpArgU, IABC, "LOADBOOL"},  // R(A) := (bool)B; if (C) pc++ 加载布尔值B到A，C=1则跳过下条指令
	opcode{0, 1, OpArgU, OpArgN, IABC, "LOADNIL "},  // R(A), R(A+1), ..., R(A+B) := nil 批量加载nil到连续寄存器
	opcode{0, 1, OpArgU, OpArgN, IABC, "GETUPVAL"},  // R(A) := UpValue[B]     获取Upvalue[B]的值到寄存器A
	opcode{0, 1, OpArgU, OpArgK, IABC, "GETTABUP"},  // R(A) := UpValue[B][RK(C)] 获取Upvalue表B中RK(C)对应元素到A
	opcode{0, 1, OpArgR, OpArgK, IABC, "GETTABLE"},  // R(A) := R(B)[RK(C)]    获取寄存器B表中RK(C)对应元素到A
	opcode{0, 0, OpArgK, OpArgK, IABC, "SETTABUP"},  // UpValue[A][RK(B)] := RK(C) 设置Upvalue表A中RK(B)位置的值为RK(C)
	opcode{0, 0, OpArgU, OpArgN, IABC, "SETUPVAL"},  // UpValue[B] := R(A)     将寄存器A的值赋值给UpValue[B]
	opcode{0, 0, OpArgK, OpArgK, IABC, "SETTABLE"},  // R(A)[RK(B)] := RK(C)   设置寄存器A表中RK(B)位置的值为RK(C)
	opcode{0, 1, OpArgU, OpArgU, IABC, "NEWTABLE"},  // R(A) := {} (size = B,C) 创建新表，预分配B个数组元素、C个哈希元素
	opcode{0, 1, OpArgR, OpArgK, IABC, "SELF    "},  // R(A+1) := R(B); R(A) := R(B)[RK(C)] 准备对象方法调用（self）
	opcode{0, 1, OpArgK, OpArgK, IABC, "ADD     "},  // R(A) := RK(B) + RK(C)  加法运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "SUB     "},  // R(A) := RK(B) - RK(C)  减法运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "MUL     "},  // R(A) := RK(B) * RK(C)  乘法运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "MOD     "},  // R(A) := RK(B) % RK(C)  取模运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "POW     "},  // R(A) := RK(B) ^ RK(C)  幂运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "DIV     "},  // R(A) := RK(B) / RK(C)  除法运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "IDIV    "},  // R(A) := RK(B) // RK(C) 整数除法运算
	opcode{0, 1, OpArgK, OpArgK, IABC, "BAND    "},  // R(A) := RK(B) & RK(C)  按位与
	opcode{0, 1, OpArgK, OpArgK, IABC, "BOR     "},  // R(A) := RK(B) | RK(C)  按位或
	opcode{0, 1, OpArgK, OpArgK, IABC, "BXOR    "},  // R(A) := RK(B) ~ RK(C)  按位异或
	opcode{0, 1, OpArgK, OpArgK, IABC, "SHL     "},  // R(A) := RK(B) << RK(C) 按位左移
	opcode{0, 1, OpArgK, OpArgK, IABC, "SHR     "},  // R(A) := RK(B) >> RK(C) 按位右移
	opcode{0, 1, OpArgR, OpArgN, IABC, "UNM     "},  // R(A) := -R(B)          取负（一元运算）
	opcode{0, 1, OpArgR, OpArgN, IABC, "BNOT    "},  // R(A) := ~R(B)          按位取反
	opcode{0, 1, OpArgR, OpArgN, IABC, "NOT     "},  // R(A) := not R(B)       逻辑取反
	opcode{0, 1, OpArgR, OpArgN, IABC, "LEN     "},  // R(A) := length of R(B) 获取字符串/表的长度
	opcode{0, 1, OpArgR, OpArgR, IABC, "CONCAT  "},  // R(A) := R(B).. ... ..R(C) 拼接B到C寄存器的值为字符串
	opcode{0, 0, OpArgR, OpArgN, IAsBx, "JMP     "}, // pc+=sBx; if (A) close all upvalues >= R(A - 1) 无条件跳转，A非0则关闭Upvalue
	opcode{1, 0, OpArgK, OpArgK, IABC, "EQ      "},  // if ((RK(B) == RK(C)) ~= A) then pc++ 相等比较，结果与A相反则跳转
	opcode{1, 0, OpArgK, OpArgK, IABC, "LT      "},  // if ((RK(B) <  RK(C)) ~= A) then pc++ 小于比较，结果与A相反则跳转
	opcode{1, 0, OpArgK, OpArgK, IABC, "LE      "},  // if ((RK(B) <= RK(C)) ~= A) then pc++ 小于等于比较，结果与A相反则跳转
	opcode{1, 0, OpArgN, OpArgU, IABC, "TEST    "},  // if not (R(A) <=> C) then pc++ 条件测试，不满足则跳转（无赋值）
	opcode{1, 1, OpArgR, OpArgU, IABC, "TESTSET "},  // if (R(B) <=> C) then R(A) := R(B) else pc++ 条件测试，满足则赋值，否则跳转
	opcode{0, 1, OpArgU, OpArgU, IABC, "CALL    "},  // R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1)) 函数调用，B=参数个数，C=返回值个数
	opcode{0, 1, OpArgU, OpArgU, IABC, "TAILCALL"},  // return R(A)(R(A+1), ... ,R(A+B-1)) 尾调用（无栈帧开销）
	opcode{0, 0, OpArgU, OpArgN, IABC, "RETURN  "},  // return R(A), ... ,R(A+B-2) 函数返回，B=返回值个数
	opcode{0, 1, OpArgR, OpArgN, IAsBx, "FORLOOP "}, // R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) } for循环迭代
	opcode{0, 1, OpArgR, OpArgN, IAsBx, "FORPREP "}, // R(A)-=R(A+2); pc+=sBx  for循环初始化（预减步长）
	opcode{0, 0, OpArgN, OpArgU, IABC, "TFORCALL"},  // R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2)); 泛型for调用迭代器，C=返回值个数
	opcode{0, 1, OpArgR, OpArgN, IAsBx, "TFORLOOP"}, // if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx } 泛型for循环迭代
	opcode{0, 0, OpArgU, OpArgU, IABC, "SETLIST "},  // R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B 设置表的数组部分元素，FPF=50
	opcode{0, 1, OpArgU, OpArgN, IABx, "CLOSURE "},  // R(A) := closure(KPROTO[Bx]) 创建函数闭包，Bx为原型索引
	opcode{0, 1, OpArgU, OpArgN, IABC, "VARARG  "},  // R(A), R(A+1), ..., R(A+B-2) = vararg 处理可变参数，B=参数个数
	opcode{0, 0, OpArgU, OpArgU, IAx, "EXTRAARG"},   // extra (larger) argument for previous opcode 扩展操作数（配合LOADKX等指令）
}
