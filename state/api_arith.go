package state

//LuaLight/state/api_arith.go
import (
	. "LuaLight/api"
	"LuaLight/number"
	"math"
)

// 定义算术/按位运算的具体实现函数
// 分为整数版本（i前缀）和浮点数版本（f前缀），适配Lua的数值类型特性
var (
	// 加法
	iadd = func(a, b int64) int64 { return a + b }
	fadd = func(a, b float64) float64 { return a + b }
	// 减法
	isub = func(a, b int64) int64 { return a - b }
	fsub = func(a, b float64) float64 { return a - b }
	// 乘法
	imul = func(a, b int64) int64 { return a * b }
	fmul = func(a, b float64) float64 { return a * b }
	// 取模（复用number包的实现，适配Lua取模规则）
	imod = number.IMod
	fmod = number.FMod
	// 幂运算（仅浮点数版本）
	pow = math.Pow
	// 普通除法（仅浮点数版本，向0截断）
	div = func(a, b float64) float64 { return a / b }
	// 向下取整除法（Lua//运算符，复用number包实现）
	iidiv = number.IFloorDiv
	fidiv = number.FFloorDiv
	// 按位与（仅整数版本）
	band = func(a, b int64) int64 { return a & b }
	// 按位或（仅整数版本）
	bor = func(a, b int64) int64 { return a | b }
	// 按位异或（仅整数版本）
	bxor = func(a, b int64) int64 { return a ^ b }
	// 按位左移（仅整数版本）
	shl = number.ShiftLeft
	// 按位右移（仅整数版本）
	shr = number.ShiftRight
	// 负号运算（单目操作，第二个参数无意义）
	iunm = func(a, _ int64) int64 { return -a }
	funm = func(a, _ float64) float64 { return -a }
	// 按位非（单目操作，第二个参数无意义）
	bnot = func(a, _ int64) int64 { return ^a }
)

// operator 封装单个运算的整数/浮点数实现
// integerFunc：整数运算函数（按位运算/整数算术运算）
// floatFunc：浮点数运算函数（算术运算）
type operator struct {
	integerFunc func(int64, int64) int64       // 整数运算实现
	floatFunc   func(float64, float64) float64 // 浮点数运算实现
}

// operators 按ArithOp常量顺序映射所有运算实现
// 索引对应api包中ArithOp枚举值（如LUA_OPADD=0，LUA_OPSUB=1...）
var operators = []operator{
	operator{iadd, fadd},   // LUA_OPADD：加法
	operator{isub, fsub},   // LUA_OPSUB：减法
	operator{imul, fmul},   // LUA_OPMUL：乘法
	operator{imod, fmod},   // LUA_OPMOD：取模
	operator{nil, pow},     // LUA_OPPOW：幂运算（仅浮点数）
	operator{nil, div},     // LUA_OPDIV：普通除法（仅浮点数）
	operator{iidiv, fidiv}, // LUA_OPIDIV：向下取整除法（//）
	operator{band, nil},    // LUA_OPBAND：按位与（仅整数）
	operator{bor, nil},     // LUA_OPBOR：按位或（仅整数）
	operator{bxor, nil},    // LUA_OPBXOR：按位异或（仅整数）
	operator{shl, nil},     // LUA_OPSHL：按位左移（仅整数）
	operator{shr, nil},     // LUA_OPSHR：按位右移（仅整数）
	operator{iunm, funm},   // LUA_OPUNM：负号（单目运算）
	operator{bnot, nil},    // LUA_OPBNOT：按位非（单目运算）
}

// Arith 执行Lua算术/按位运算（核心对外接口）
// op：运算类型（ArithOp枚举），通过栈完成操作数入参/结果出参：
// 1. 双目运算：弹出栈顶两个值（b=栈顶，a=次顶），计算后将结果压栈；
// 2. 单目运算（UNM/BNOT）：仅弹出栈顶一个值（a=b=栈顶），计算后压栈；
// 3. 运算失败则触发panic（如类型不支持）
func (self *luaState) Arith(op ArithOp) {
	var a, b luaValue
	b = self.stack.pop() // 弹出栈顶值作为第二个操作数
	// 单目运算（负号/按位非）：仅需一个操作数，a复用b的值
	if op != LUA_OPUNM && op != LUA_OPBNOT {
		a = self.stack.pop() // 双目运算：弹出次顶值作为第一个操作数
	} else {
		a = b // 单目运算：a和b指向同一个值
	}

	// 获取当前运算的实现函数
	operator := operators[op]
	// 执行运算并处理结果
	if result := _arith(a, b, operator); result != nil {
		self.stack.push(result) // 运算成功：结果压栈
	} else {
		panic("arithmetic error") // 运算失败：类型不支持/转换失败
	}
}

// _arith 核心运算执行逻辑（内部辅助函数）
// a/b：运算操作数；op：运算实现；返回值：运算结果（nil表示失败）
// 执行优先级：
// 1. 仅整数运算（按位操作）：尝试将a/b转为整数，执行整数函数；
// 2. 混合运算（算术操作）：优先用整数运算，失败则转浮点数运算；
// 3. 仅浮点数运算（幂/普通除法）：尝试将a/b转为浮点数，执行浮点数函数；
func _arith(a, b luaValue, op operator) luaValue {
	// 分支1：仅支持整数运算（按位操作，floatFunc为nil）
	if op.floatFunc == nil {
		// 尝试将a/b转为整数，成功则执行整数运算
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.integerFunc(x, y)
			}
		}
		// 分支2：支持浮点数/混合运算（算术操作）
	} else {
		// 子分支2.1：优先执行整数运算（提升性能）
		if op.integerFunc != nil {
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.integerFunc(x, y)
				}
			}
		}
		// 子分支2.2：整数运算失败，尝试转浮点数运算
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.floatFunc(x, y)
			}
		}
	}
	// 所有转换/运算都失败（如操作数是字符串/table）
	return nil
}
