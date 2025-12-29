//LuaLight/api/consts.go

package api

// Lua基础数据类型标识（与Lua官方API定义一致）
const (
	LUA_TNONE          = iota - 1 // 无类型（-1）
	LUA_TNIL                      // nil值
	LUA_TBOOLEAN                  // 布尔值
	LUA_TLIGHTUSERDATA            // 轻量用户数据（不参与GC）
	LUA_TNUMBER                   // 数值（整数/浮点数）
	LUA_TSTRING                   // 字符串
	LUA_TTABLE                    // 表
	LUA_TFUNCTION                 // 函数（Lua函数/C函数）
	LUA_TUSERDATA                 // 全量用户数据（参与GC）
	LUA_TTHREAD                   // 协程
)

const (
	LUA_OPADD  = iota // +
	LUA_OPSUB         // -
	LUA_OPMUL         // *
	LUA_OPMOD         // %
	LUA_OPPOW         // ^
	LUA_OPDIV         // /
	LUA_OPIDIV        // //整除
	LUA_OPBAND        // &
	LUA_OPBOR         // |
	LUA_OPBXOR        // ~
	LUA_OPSHL         // <<
	LUA_OPSHR         // >>
	LUA_OPUNM         // - (unary minus)
	LUA_OPBNOT        // ~
)

const (
	LUA_OPEQ = iota // ==
	LUA_OPLT        // <
	LUA_OPLE        // <=
)
