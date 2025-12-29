package state

//LuaLight/state/api_compare.go
import (
	. "LuaLight/api" // 导入比较操作符常量（CompareOp）、luaValue类型等
)

// Compare 执行Lua的比较运算（对外核心接口）
// idx1/idx2：栈索引（支持绝对/相对索引），指定要比较的两个栈元素；
// op：比较操作类型（LUA_OPEQ/OPLT/OPLE）；
// 返回值：比较结果（bool），不支持的操作/类型会触发panic
func (self *luaState) Compare(idx1, idx2 int, op CompareOp) bool {
	// 从栈中获取两个待比较的元素（先转换索引为有效索引）
	a := self.stack.get(idx1)
	b := self.stack.get(idx2)

	// 根据比较操作类型，调用对应的核心比较函数
	switch op {
	case LUA_OPEQ: // 等于（==）
		return _eq(a, b)
	case LUA_OPLT: // 小于（<）
		return _lt(a, b)
	case LUA_OPLE: // 小于等于（<=）
		return _le(a, b)
	default: // 不支持的比较操作
		panic("invalid compare op!")
	}
}

// _eq 实现Lua的“等于（==）”比较规则（核心：宽松类型匹配）
// Lua的==规则：
// 1. nil只等于nil；
// 2. 布尔值仅和布尔值比较，且值相同；
// 3. 字符串仅和字符串比较，且内容相同；
// 4. 数值类型（int64/float64）跨类型比较（如10==10.0返回true）；
// 5. 其他类型（table/function等）仅引用相同时相等；
func _eq(a, b luaValue) bool {
	switch x := a.(type) {
	case nil: // a是nil：仅当b也是nil时相等
		return b == nil
	case bool: // a是布尔值：b必须也是布尔值且值相同
		y, ok := b.(bool)
		return ok && x == y
	case string: // a是字符串：b必须也是字符串且内容相同
		y, ok := b.(string)
		return ok && x == y
	case int64: // a是整数：支持和int64/float64比较
		switch y := b.(type) {
		case int64: // b是整数：直接比较值
			return x == y
		case float64: // b是浮点数：转浮点数后比较（10==10.0→true）
			return float64(x) == y
		default: // b是其他类型：不相等
			return false
		}
	case float64: // a是浮点数：支持和float64/int64比较
		switch y := b.(type) {
		case float64: // b是浮点数：直接比较值
			return x == y
		case int64: // b是整数：转浮点数后比较（10.0==10→true）
			return x == float64(y)
		default: // b是其他类型：不相等
			return false
		}
	default: // 其他类型（table/function/thread等）：仅引用相同才相等
		return a == b
	}
}

// _lt 实现Lua的“小于（<）”比较规则（核心：严格类型限制，仅支持字符串/数值）
// Lua的<规则：
// 1. 仅支持字符串和字符串比较（按字典序）、数值和数值比较（跨int/float）；
// 2. 其他类型（bool/table等）比较会触发panic；
func _lt(a, b luaValue) bool {
	switch x := a.(type) {
	case string: // a是字符串：b必须也是字符串，按字典序比较
		if y, ok := b.(string); ok {
			return x < y
		}
	case int64: // a是整数：支持和int64/float64比较
		switch y := b.(type) {
		case int64: // b是整数：直接比较
			return x < y
		case float64: // b是浮点数：转浮点数后比较
			return float64(x) < y
		}
		// 注意：_lt未处理float64类型的a！因为float64的a会走default触发panic
		// （这是代码的简化设计，完整实现需补充float64分支，和_le对齐）
	}
	// 不支持的比较类型（如bool/table/浮点数a等）
	panic("comparison error")
}

// _le 实现Lua的“小于等于（<=）”比较规则（核心：比_lt多支持float64类型的a）
// 规则和_lt一致，仅补充了float64类型的处理，支持更完整的数值比较
func _le(a, b luaValue) bool {
	switch x := a.(type) {
	case string: // a是字符串：b必须也是字符串，按字典序比较
		if y, ok := b.(string); ok {
			return x <= y
		}
	case int64: // a是整数：支持和int64/float64比较
		switch y := b.(type) {
		case int64:
			return x <= y
		case float64:
			return float64(x) <= y
		}
	case float64: // a是浮点数：支持和float64/int64比较（_lt未处理此分支）
		switch y := b.(type) {
		case float64:
			return x <= y
		case int64:
			return x <= float64(y)
		}
	}
	// 不支持的比较类型
	panic("comparison error!")
}
