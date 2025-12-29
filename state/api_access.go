package state

//从栈中获取信息

// 对齐Lua官方API语义，支持Lua基础类型的判断、转换

import (
	. "LuaLight/api" // 导入Lua类型常量（如LUA_TNIL/LUA_TBOOLEAN等）
	"fmt"
)

// TypeName 将Lua类型标识（LuaType）转换为可读的类型名称
// tp：LuaType类型常量（如LUA_TNIL）；返回值：对应的字符串名称（如"nil"）
func (self *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:
		return "no value" // 无值（索引无效时的类型）
	case LUA_TNIL:
		return "nil" // nil值
	case LUA_TBOOLEAN:
		return "boolean" // 布尔值
	case LUA_TNUMBER:
		return "number" // 数值（整数/浮点数）
	case LUA_TSTRING:
		return "string" // 字符串
	case LUA_TTABLE:
		return "table" // 表
	case LUA_TFUNCTION:
		return "function" // 函数
	case LUA_TTHREAD:
		return "thread" // 协程
	default:
		return "userdata" // 用户数据（轻量/全量）
	}
}

// Type 根据索引获取栈中元素的Lua类型标识
// idx：栈索引（支持相对/绝对）；返回值：LuaType类型常量
// 若索引无效（指向不存在的元素），返回LUA_TNONE
func (self *luaState) Type(idx int) LuaType {
	if self.stack.isValid(idx) { // 先校验索引有效性
		val := self.stack.get(idx) // 读取栈元素
		return typeOf(val)         // 底层判断元素的Lua类型
	}
	return LUA_TNONE // 索引无效 → 无值类型
}

// IsNone 判断指定索引位置是否为「无值」（LUA_TNONE）
// 无值表示索引无效，并非nil（nil是有效类型）
func (self *luaState) IsNone(idx int) bool {
	return self.Type(idx) == LUA_TNONE
}

// IsNil 判断指定索引位置的元素是否为nil（LUA_TNIL）
func (self *luaState) IsNil(idx int) bool {
	return self.Type(idx) == LUA_TNIL
}

// IsNoneOrNil 判断指定索引位置是否为「无值」或nil
// 因LUA_TNONE=-1、LUA_TNIL=0，故直接判断Type返回值≤0即可
func (self *luaState) IsNoneOrNil(idx int) bool {
	return self.Type(idx) <= LUA_TNIL
}

// IsBoolean 判断指定索引位置的元素是否为布尔类型（LUA_TBOOLEAN）
func (self *luaState) IsBoolean(idx int) bool {
	return self.Type(idx) == LUA_TBOOLEAN
}

// IsString 判断指定索引位置的元素是否为字符串类型（Lua特有语义）
// Lua中数值（number）可自动转换为字符串，故number也返回true
func (self *luaState) IsString(idx int) bool {
	t := self.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

// IsNumber 判断指定索引位置的元素是否为数值类型（可转换为number）
// 底层通过ToNumberX判断是否能成功转换为float64，而非仅判断类型
func (self *luaState) IsNumber(idx int) bool {
	_, ok := self.ToNumberX(idx)
	return ok
}

// IsInteger 判断指定索引位置的元素是否为整数类型（int64）
// 仅严格匹配int64类型，数值类型的float64（如10.0）会返回false
func (self *luaState) IsInteger(idx int) bool {
	val := self.stack.get(idx)
	_, ok := val.(int64) // 类型断言判断是否为int64
	return ok
}

// ToBoolean 将指定索引位置的元素转换为布尔值（Lua布尔转换规则）
// Lua中：nil/false→false，其余值（包括0/""）→true
func (self *luaState) ToBoolean(idx int) bool {
	val := self.stack.get(idx)
	return convertToBoolean(val) // 底层按Lua规则转换
}

// ToNumber 将指定索引位置的元素转换为浮点数（float64）
// 转换失败时返回0（无第二个返回值，容错版）
func (self *luaState) ToNumber(idx int) float64 {
	n, _ := self.ToNumberX(idx)
	return n
}

// ToNumberX 将指定索引位置的元素转换为浮点数（带转换结果）
// 返回值1：转换后的float64值；返回值2：是否转换成功
// 支持的类型：float64（直接返回）、int64（强转为float64），其余类型失败
func (self *luaState) ToNumberX(idx int) (float64, bool) {
	val := self.stack.get(idx)
	return convertToFloat(val)
}

// ToInteger 将指定索引位置的元素转换为整数（int64）
// 转换失败时返回0（无第二个返回值，容错版）
func (self *luaState) ToInteger(idx int) int64 {
	i, _ := self.ToIntegerX(idx)
	return i
}

// ToIntegerX 将指定索引位置的元素转换为整数（带转换结果）
// 返回值1：转换后的int64值；返回值2：是否转换成功
// 仅严格匹配int64类型，float64（如10.0）会转换失败
func (self *luaState) ToIntegerX(idx int) (int64, bool) {
	val := self.stack.get(idx)
	return convertToInteger(val)
}

// ToStringX 将指定索引位置的元素转换为字符串（带转换结果，Lua特有语义）
// 返回值1：转换后的字符串；返回值2：是否转换成功
// 转换规则：
// 1. string类型：直接返回；
// 2. number类型（int64/float64）：转为字符串并替换栈中原值（Lua自动转换特性）；
// 3. 其余类型：转换失败
func (self *luaState) ToStringX(idx int) (string, bool) {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x) // 数值转字符串
		self.stack.set(idx, s)    // 替换栈中原值（Lua的自动类型转换）
		return s, true
	default:
		return "", false // 非字符串/数值类型转换失败
	}
}

// ToString 将指定索引位置的元素转换为字符串
// 转换失败时返回空字符串（无第二个返回值，容错版）
func (self *luaState) ToString(idx int) string {
	s, _ := self.ToStringX(idx)
	return s
}
