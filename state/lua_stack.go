// LuaLight/state/lua_stack.go
package state

// luaStack 定义Lua虚拟机的栈结构（底层存储核心）
// slots：存储栈元素的底层数组（0索引），元素类型为luaValue（支持Lua所有基础类型）
// top：栈顶的绝对索引（Lua栈索引，从1开始），栈为空时top=0，有n个元素时top=n
type luaStack struct {
	slots []luaValue // 栈元素存储容器（Go数组，0索引）
	top   int        // 栈顶的Lua绝对索引（非数组下标）
}

// newLuaStack 创建指定初始容量的Lua栈
// size：栈底层数组的初始容量，top初始化为0（栈空）
func newLuaStack(size int) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size), // 初始化底层数组
		top:   0,                      // 初始栈空，栈顶索引为0
	}
}

// check 检查栈剩余空间，确保能容纳至少n个新元素
// 若空闲空间不足，自动扩容底层数组（追加nil），避免push时溢出
func (self *luaStack) check(n int) {
	free := len(self.slots) - self.top // 计算剩余空闲空间（数组总长度 - 当前栈顶索引）
	// 空闲不足时，循环追加nil扩容，直到能容纳n个新元素
	for i := free; i < n; i++ {
		self.slots = append(self.slots, nil)
	}
}

// push 将值压入栈顶
// val：要压入的luaValue类型值
// 注意：调用前需确保栈有空间（通常通过check方法），否则触发stack overflow panic
func (self *luaStack) push(val luaValue) {
	if self.top == len(self.slots) {
		panic("stack overflow!") // 栈溢出：无空间压入新元素
	}
	self.slots[self.top] = val // 存入数组（top是Lua索引，对应数组下标）
	self.top++                 // 栈顶索引上移一位
}

// pop 弹出栈顶元素并返回
// 返回值：栈顶的luaValue类型值，弹出后原位置置nil（避免内存泄漏）
// 若栈空（top<1），触发stack underflow panic
func (self *luaStack) pop() luaValue {
	if self.top < 1 {
		panic("stack underflow!") // 栈下溢：无元素可弹出
	}
	self.top--                  // 栈顶索引下移一位
	val := self.slots[self.top] // 读取栈顶元素（数组下标=top）
	self.slots[self.top] = nil  // 清空原栈顶位置（避免悬空引用）
	return val
}

/*
索引规则说明（核心）：
1. 绝对索引：从栈底开始计数，正数（1=栈底第一个元素，top=栈顶元素）；
2. 相对索引：从栈顶开始计数，负数（-1=栈顶元素，-top=栈底第一个元素）；
3. 有效索引：指向栈中已存在元素的索引（绝对/相对均可）；
4. Lua索引 → 数组下标：absIndex(idx) - 1（因数组是0索引，Lua索引是1开始）。
*/

// absIndex 将传入的索引（相对/绝对）转换为绝对索引
// idx≥0：已是绝对索引，直接返回；
// idx<0：相对索引，转换为绝对索引（公式：idx + top + 1）；
// 例：栈顶top=3，idx=-1 → 3 + (-1) + 1 = 3（栈顶绝对索引）
func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 {
		return idx
	}
	return idx + self.top + 1
}

// isValid 判断传入的索引（相对/绝对）是否为有效索引
// 有效索引：转换为绝对索引后，1 ≤ absIdx ≤ top（指向栈中已存在元素）
func (self *luaStack) isValid(idx int) bool {
	absIdx := self.absIndex(idx) // 先转换为绝对索引
	return absIdx > 0 && absIdx <= self.top
}

// get 根据索引（相对/绝对）从栈中取值
// idx：要读取的索引（支持相对/绝对）；
// 返回值：对应位置的luaValue（无效索引返回nil）；
// 核心转换：Lua绝对索引 → 数组下标（absIdx - 1）
func (self *luaStack) get(idx int) luaValue {
	absIdx := self.absIndex(idx)
	// 仅当索引有效时，读取数组对应位置的值
	if absIdx > 0 && absIdx <= self.top {
		return self.slots[absIdx-1] // 数组下标 = 绝对索引 - 1
	}
	return nil // 无效索引返回nil
}

// set 根据索引（相对/绝对）往栈中写入值
// idx：要写入的索引（支持相对/绝对）；val：要写入的luaValue；
// 仅当索引有效时写入，无效索引触发invalid index panic
func (self *luaStack) set(idx int, val luaValue) {
	absIdx := self.absIndex(idx)
	// 仅当索引有效时，写入数组对应位置
	if absIdx > 0 && absIdx <= self.top {
		self.slots[absIdx-1] = val // 数组下标 = 绝对索引 - 1
		return                     // 写入成功后返回，避免触发panic
	}
	panic("invalid index!") // 无效索引：无法写入
}

// reverse 反转底层数组中[from, to]区间的元素（辅助函数）
// from/to：底层数组的下标（0开始），而非Lua栈索引；
// 用途：为Rotate操作提供底层支持，通过三次反转实现高效的局部旋转
func (self *luaStack) reverse(from, to int) {
	slots := self.slots // 简化底层数组引用
	// 双指针向中间靠拢，逐个交换元素实现反转
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}
