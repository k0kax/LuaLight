// LuaLight/state/api_state.go
package state

// GetTop 获取当前栈顶的绝对索引（Lua栈索引，非Go数组下标）
// 栈为空时返回0，栈中有n个元素时返回n
func (self *luaState) GetTop() int {
	return self.stack.top
}

// AbsIndex 将传入的索引（相对/绝对）转换为绝对索引
// 相对索引（如-1=栈顶）→ 绝对索引；绝对索引直接返回
func (self *luaState) AbsIndex(idx int) int {
	return self.stack.absIndex(idx)
}

// CheckStack 检查栈剩余空间，确保能容纳n个新元素
// 空间不足时自动扩容，返回true表示检查/扩容成功（不会失败）
func (self *luaState) CheckStack(n int) bool {
	self.stack.check(n)
	return true
}

// Pop 弹出栈顶n个元素（n≤0时无操作）
// 底层通过SetTop实现：-n-1 表示栈顶需调整到「原栈顶 - n」的位置
func (self *luaState) Pop(n int) {
	// 注释保留原实现思路，体现优化逻辑
	// for i := 0; i < n; i++ {
	// 	self.stack.pop()
	// }
	self.SetTop(-n - 1)
}

// Copy 将fromIdx位置的元素复制到toIdx位置
// fromIdx/toIdx支持相对/绝对索引，复制后原位置元素保留
func (self *luaState) Copy(fromIdx, toIdx int) {
	val := self.stack.get(fromIdx) // 读取源位置值
	self.stack.set(toIdx, val)     // 写入目标位置
}

// PushValue 将指定索引处的元素复制并压入栈顶
// 原位置元素保留，栈顶索引+1
func (self *luaState) PushValue(idx int) {
	val := self.stack.get(idx) // 读取指定位置值
	self.stack.push(val)       // 压入栈顶
}

// Replace 将栈顶元素弹出，覆盖到指定索引位置
// 栈顶索引-1，指定位置原有值被替换
func (self *luaState) Replace(idx int) {
	val := self.stack.pop()  // 弹出栈顶值
	self.stack.set(idx, val) // 写入指定位置
}

// Insert 将栈顶元素弹出，插入到指定索引位置
// 插入后，指定位置及以上元素上移一位，栈顶索引不变
// 底层通过Rotate实现：将[idx, top]区间上旋1位
func (self *luaState) Insert(idx int) {
	self.Rotate(idx, 1)
}

// Remove 删除指定索引处的元素
// 被删除位置上方的元素全部下移一位，栈顶索引-1
// 底层通过Rotate实现：先下旋1位将目标元素移到栈顶，再Pop弹出
func (self *luaState) Remove(idx int) {
	self.Rotate(idx, -1)
	self.Pop(1)
}

// Rotate 将[idx, 栈顶]索引区间内的元素朝栈顶方向旋转n个位置
// n>0：上旋（栈顶方向），n<0：下旋（栈底方向）
// 核心逻辑：通过三次反转实现旋转，是高效的数组局部旋转算法
func (self *luaState) Rotate(idx, n int) {
	t := self.stack.top - 1           // 栈顶元素的Go数组下标（top是Lua索引，转下标需-1）
	p := self.stack.absIndex(idx) - 1 // 目标起始位置的Go数组下标
	var m int
	if n >= 0 {
		m = t - n // 上旋时的分割点下标
	} else {
		m = p - n - 1 // 下旋时的分割点下标
	}

	// 三次反转实现旋转：[p,m] → [m+1,t] → [p,t]
	self.stack.reverse(p, m)
	self.stack.reverse(m+1, t)
	self.stack.reverse(p, t)
}

// SetTop 设置栈顶索引为指定值（核心栈调整方法）
// idx支持相对/绝对索引，调整规则：
// 1. 新栈顶 < 原栈顶：弹出多余元素（栈截断）
// 2. 新栈顶 > 原栈顶：补充nil元素（栈扩容）
// 3. idx为负且转换后<0：触发栈下溢panic
func (self *luaState) SetTop(idx int) {
	newTop := self.stack.absIndex(idx) // 转换为绝对索引
	if newTop < 0 {
		panic("stack underflow!") // 栈下溢：索引不能为负
	}

	n := self.stack.top - newTop // 计算栈顶调整差值（n>0：要弹出；n<0：要补充）
	if n > 0 {
		// 弹出n个元素：栈顶从self.stack.top → newTop
		for i := 0; i < n; i++ {
			self.stack.pop()
		}
	} else if n < 0 {
		// 补充-n个nil：栈顶从self.stack.top → newTop
		for i := 0; i > n; i-- {
			self.stack.push(nil)
		}
	}
}
