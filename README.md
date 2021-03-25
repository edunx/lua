# rock lua vm 虚拟机
基于gopher-lua 做了一些底层改造 , 增加一些轻量的数据类型 ，减少运行逻辑，更加符合项目开发

## LState.ExData
- 说明: 主要作用是在lua虚拟机底层直接写入一些值 减少盏交互次数 在thread 模式下很有用
- 函数: LState.ExData.Get(string) , LState.ExData.Set( string )
```go
    L := lua.NewState()
    co := L.NewThread()
    
    ctx := struct {name string , val interface{} }{"edunx" , []byte("hhh")}
    co.ExData.Set("ctx" , ctx)
    
    v := co.ExData.Get("ctx")
    //todo other code
```

## UserKV.*
主要是简化lua 到goalng的table逻辑， 很多情况下key-val只要简单的key-val结构 不许要复杂的和执行过程
而且大部分情况下 table 底层绑定的是map的hash结构和Slice结构，执行过程中搜索步骤较长 而且map底层
县城读写并不安全

- 函数： UserKV.Get(string) , UserKV.Set( string )
```go
    kv := &UserKV{}
    ud := L.NewLightUserData( obj )
    kv.Set("name" , ud)
    kv.Set("other" , "ooo")
    
    kv.Get("name")
    kv.Get("other")
    
    L.SetGlobal("user" , kv)
```
- lua示例
```lua
    print(type(user.name))
    print(user.other)
```


## lua.Args.*
- 说明: 主要用在GFunction中的参数模块,可以快速获取参数
- 用法: 有常见的CheckString ,CheckInt ...
```go
    str := args.CheckString(L , 1)
    num := args.CheckString(L , 2)
    ..
    n := args.Len()
```
## GFunction
- 说明: 没有嵌套callframe 不用生成LFunction 运行GO语言逻辑，减少运行步骤
- 语法: 一定要满足 func(*lua.LState , args *lua.Args) lua.LValue的GO function
- 用法: 跟string,int的使用方法一样 如下:
```go
    type A struct {
	    lua.Super
	    name string
	    num  int
    }
    
    func (a *A) debug(L *lua.LState , args *lua.Args) lua.LValue {
        return lua.LString(fmt.Sprintf("name:%s , num: %d" , a.name , a.num))	
    }
    
    func (a *A) Index(L *lua.LState , key string) lua.LValue {
        if key == "debug" { return lua.NewGFunction(a) }	
        return lua.LNil
    }

    //构造的GFunction
    func GFunc(L *lua.LState , args *lua.Args) lua.LValue {
        name := args.CheckString(L , 1)	
        num  := args.CheckInt(L , 2)
        ud := &A{name: name , num: num } 
        return  L.NewLightUserData(ud)
    }
    
    //注入 
    L.SetGlobal("gn" , lua.NewGFunction(GFunc))
```

```lua
    local ud = gn("edunx" , 18)
    print( ud.debug() )
```

## lightuserdata 
- 说明： lightuserdata 是类似于 C lua 中的 lightuserdata
- 用法： 减少了MT 方法的绑定 利用 luaSetGetFunc 直接获sturct 对象资源
- 结构定义如下
```go
    type LCallBack  func( interface{} ) // 用于Lcheck方法 满足传进来的obj 数据类型后 会把执行回调
    type rock interface {
        Close()
        Start() error

        Write( interface{} ) error
        Read() ([]byte , error)
        Type() string
        Json() []byte

        SetField(*LState   , LValue, LValue )
        GetField(*LState   , LValue)  LValue

        Index(*LState      , string) LValue
        NewIndex(*LState   , string , LValue)

        LCheck(interface{} , LCallBack) bool //check(obj interface{}, set func) bool
        ToLightUserData(*LState) *LightUserData
    }
    
    type LightUserData struct {
        Value   rock 
        //other
    } 
    
    type Super struct {}
    //防止过多的方法定义
	//为了防止过多方法定义可以先继承super
    //定义rock所有的方法   
```

- 示例
```go
    import (
    	"github.com/edunx/lua"
    )

    type Lud struct {
        lua.Super	
        name string
        val string 
    }
    
    func(ud *Lud) Index(L *lua.LState , key string) lua.LValue {
        if key == "name"   { return lua.LString( ud.name) }
        if key == "val"    { return lua.LString( ud.val ) }
        if key == ud.name  { return lua.LString( ud.val ) }
        return lua.LNil
    }
    
    func(ud *Lud) NewIndex(L *lua.LState , key string , value lua.LValue) {
        if key == "name"  { ud.name = value.String() }	
        if key == "val"   { ud.val = value.String()  }
        if key == ud.name { ud.val = value.String()  }
        return lua.LNil
    }
    //GetField 和 SetField 原理Index 和NewIndex 类似 只是key的类型不一样
    //就不继续说
    
    func(ud *Lud) debug(L *lua.LState , args *lua.Args) lua.LValue {
    	prefix := args.CheckString(L , 1)
    	
    	return lua.LString(fmt.Sprintf("prefix: %s , name: %s , val: %s" 
    	    , prefix , ud.name , ud.val ))
    }
    
    func createLud(L *lua.LState) {
    	ud := &Lud{ name: "edunx" , val: "git"}
		L.SetGlobal("udata" , lua.NewLightUserData(ud) )
    }
```
- 脚本是示例
```lua
    print(udata.debug("helo-"))
    udata.name = "helo"
    udata.val = "yes"
```
