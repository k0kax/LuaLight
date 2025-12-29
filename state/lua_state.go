// LuaLight/state/lua_state.go
package state

//接口的实现
type luaState struct {
	stack *luaStack
}

func New() *luaState {
	return &luaState{
		stack: newLuaStack(20),
	}
}
