// vm/instruction.go
package vm

// 指令操作数最大值定义（Lua 5.3字节码规范）
const (
	MAXARG_Bx  = 1<<18 - 1      // Bx操作数最大值（无符号18位）：262143
	MAXARG_sBx = MAXARG_Bx >> 1 // sBx操作数最大值（有符号18位）：131071（范围：-131071 ~ 131071）
)

/*
Lua 5.3字节码指令格式（4字节=32位），按编码模式拆分：
 31       22       13       5    0
  +-------+^------+-^-----+-^-----
  |b=9bits |c=9bits |a=8bits|op=6|  IABC模式：op(6) + A(8) + B(9) + C(9)
  +-------+^------+-^-----+-^-----
  |    bx=18bits    |a=8bits|op=6|  IABx模式：op(6) + A(8) + Bx(18)
  +-------+^------+-^-----+-^-----
  |   sbx=18bits    |a=8bits|op=6|  IAsBx模式：op(6) + A(8) + sBx(18，有符号)
  +-------+^------+-^-----+-^-----
  |    ax=26bits            |op=6|  IAx模式：op(6) + Ax(26)
  +-------+^------+-^-----+-^-----
 31      23      15       7      0
*/
type Instruction uint32 // 单条Lua字节码指令（4字节）

// Opcode 提取指令的操作码（低6位）
func (self Instruction) Opcode() int {
	return int(self & 0x3F) // 0x3F=63，掩码提取低6位
}

// ABC 从IABC模式指令中提取A/B/C三个操作数
func (self Instruction) ABC() (a, b, c int) {
	a = int(self >> 6 & 0xFF)   // A：6-13位（8位），0xFF掩码提取
	b = int(self >> 14 & 0x1FF) // B：14-22位（9位），0x1FF掩码提取
	c = int(self >> 23 & 0x1FF) // C：23-31位（9位），0x1FF掩码提取（注：原代码b重复赋值，已修正为c）
	return
}

// ABx 从IABx模式指令中提取A/Bx两个操作数
func (self Instruction) ABx() (a, bx int) {
	a = int(self >> 6 & 0xFF) // A：6-13位（8位）
	bx = int(self >> 14)      // Bx：14-31位（18位，无符号）
	return
}

// AsBx 从IAsBx模式指令中提取A/sBx两个操作数（sBx为有符号数）
func (self Instruction) AsBx() (a, sbx int) {
	a, bx := self.ABx()
	return a, bx - MAXARG_sBx // 转换为有符号数（范围：-MAXARG_sBx ~ MAXARG_sBx）
}

// Ax 从IAx模式指令中提取Ax操作数
func (self Instruction) Ax() int {
	return int(self >> 6) // Ax：6-31位（26位）
}

// OpName 获取指令名称（如"MOVE"、"LOADK"）
func (self Instruction) OpName() string {
	return opcodes[self.Opcode()].name
}

// OpMode 获取指令的编码模式（IABC/IABx/IAsBx/IAx）
func (self Instruction) OpMode() byte {
	return opcodes[self.Opcode()].opMode
}

// BMode 获取操作数B的类型（OpArgN/OpArgU/OpArgR/OpArgK）
func (self Instruction) BMode() byte {
	return opcodes[self.Opcode()].argBMode
}

// CMode 获取操作数C的类型（OpArgN/OpArgU/OpArgR/OpArgK）
func (self Instruction) CMode() byte {
	return opcodes[self.Opcode()].argCMode
}
