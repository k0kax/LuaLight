// LuaLight/api/lua_state.go
package api

//接口库
type LuaType = int
type ArithOp = int
type CompareOp = int

// LuaState 定义Lua虚拟机栈操作和类型交互的核心接口，对齐Lua官方C API语义
type LuaState interface {
	/* basic stack manipulation - 栈基础操作 */
	GetTop() int             // 获取栈顶索引
	AbsIndex(idx int) int    // 将相对索引转换为绝对索引
	CheckStack(n int) bool   // 检查栈空间，确保能容纳n个新值（不足则扩容）
	Pop(n int)               // 弹出栈顶n个值
	Copy(fromIdx, toIdx int) // 复制fromIdx位置的值到toIdx位置
	PushValue(idx int)       // 将指定索引的值压入栈顶
	Replace(idx int)         // 弹出栈顶值，替换指定索引位置的值
	Insert(idx int)          // 将栈顶值插入到指定索引位置
	Remove(idx int)          // 删除指定索引位置的值，后续值下移
	Rotate(idx, n int)       // 对idx以上的栈元素旋转n步（n正=上旋，n负=下旋）
	SetTop(idx int)          // 设置栈顶索引（截断/填充nil）

	/* access functions (stack -> Go) - 栈值读取到Go类型 */
	TypeName(tp LuaType) string        // 根据类型标识返回类型名称（如"nil"/"table"）
	Type(idx int) LuaType              // 获取指定索引值的类型标识
	IsNone(idx int) bool               // 检查指定索引是否为无类型（LUA_TNONE）
	IsNil(idx int) bool                // 检查指定索引值是否为nil
	IsNoneOrNil(idx int) bool          // 检查指定索引是否为无类型或nil
	IsBoolean(idx int) bool            // 检查指定索引值是否为布尔类型
	IsInteger(idx int) bool            // 检查指定索引值是否为整数类型
	IsNumber(idx int) bool             // 检查指定索引值是否为数值类型（整数/浮点数）
	IsString(idx int) bool             // 检查指定索引值是否为字符串类型
	ToBoolean(idx int) bool            // 将指定索引值转换为布尔值（非nil/非false均为true）
	ToInteger(idx int) int64           // 将指定索引值转换为整数（失败返回0）
	ToIntegerX(idx int) (int64, bool)  // 转换为整数，返回值+是否成功
	ToNumber(idx int) float64          // 将指定索引值转换为浮点数（失败返回0）
	ToNumberX(idx int) (float64, bool) // 转换为浮点数，返回值+是否成功
	ToString(idx int) string           // 将指定索引值转换为字符串（失败返回""）
	ToStringX(idx int) (string, bool)  // 转换为字符串，返回值+是否成功

	/* push functions (Go -> stack) - Go类型值压入栈 */
	PushNil()             // 压入nil值到栈顶
	PushBoolean(b bool)   // 压入布尔值到栈顶
	PushInteger(n int64)  // 压入整数值到栈顶
	PushNumber(n float64) // 压入浮点数值到栈顶
	PushString(s string)  // 压入字符串值到栈顶

	Arith(op ArithOp)                          //用于执行算术和按位运算
	Compare(idx1, odx2 int, op CompareOp) bool //用于执行比较运算
	Len(idx int)                               //用于执行取长度运算
	Concat(n int)                              //用于执行字符串拼接运算
}
