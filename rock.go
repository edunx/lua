package lua

import (
	"fmt"
	"sync"
)
type ExDataKV struct {
	key       string
	value     interface{}
}

type ExData  []ExDataKV

func (ed *ExData) Set(key string , value interface{}) {
	args := *ed

	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if key == kv.key {
			kv.value = value
			return
		}
		if kv.key == "" {
			kv.value = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = key

		kv.value = value
		*ed = args
		return
	}

	kv := ExDataKV{}
	kv.key = key
	kv.value = value
	*ed = append(args, kv)
}

func (ed *ExData) Get(key string )  interface{} {
	args := *ed
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if key == kv.key {
			return kv.value
		}
	}

	return nil
}

func (ed *ExData) Del( key string ) {
	args := *ed
	n := len(args)
	for i := 0 ; i < n ; i++ {
		kv := &args[i]
		if kv.key == key {
			kv.key = ""
			kv.value = nil
			goto DONE
		}
	}

DONE:
	*ed = args
}

func (ed *ExData) Reset() {
	*ed =(*ed)[:0]
}


type LCallBack func( obj interface{} ) //用来回调方法
type luaSetGetFunc  interface {
	SetField(*LState   , LValue, LValue )
	GetField(*LState   , LValue)  LValue
	Index(*LState      , string) LValue
	NewIndex(*LState   , string , LValue)
	LCheck(interface{} , LCallBack) bool //check(obj interface{}, set func) bool
	ToLightUserData(*LState) *LightUserData
}

type LightUserData struct {
	Value    luaSetGetFunc
	fnExData ExData
}

func (ud *LightUserData) String() string                     { return fmt.Sprintf("userdata: %p", ud) }
func (ud *LightUserData) Type() LValueType                   { return LTLightUserData}
func (ud *LightUserData) assertFloat64() (float64, bool)     { return 0, false }
func (ud *LightUserData) assertString() (string, bool)       { return "", false }
func (ud *LightUserData) assertFunction() (*LFunction, bool) { return nil, false }

type Args []LValue
var argsPool = &sync.Pool{
	New: func() interface {} { return &Args{} },
}

func (a *Args) Get( idx int) LValue {
	args := *a
	id := idx - 1
	n := len(args)
	if id < 0 || id >= n  {
		return LNil
	}
	return args[id]
}

func (a *Args) Set( val LValue) {
	args := *a
	*a = append(args, val)
}

func (a *Args) Len() int {
	return len(*a)
}

func (a *Args) reset() {
	*a = (*a)[:0]
}

func (a *Args) LGet(L *LState , idx int) LValue {
	id := idx - 1
	if id < 0 {
		L.RaiseError("#%d not found")
		return LNil
	}

	if id >= a.Len() {
		L.RaiseError("#%d overflower" , idx)
		return LNil
	}

	return (*a)[id]
}

func (a *Args) Int(L *LState , n int ) int {
	v := L.Get(n)
	if intv, ok := v.(LNumber); ok {
		return int(intv)
	}
	L.TypeError(n, LTNumber)
	return 0
}

func (a *Args) CheckAny(L *LState , n int) LValue {
	return a.LGet(L ,n)
}

func (a *Args) CheckInt(L *LState , n int) int {
	v := a.LGet(L , n)
	if intv, ok := v.(LNumber); ok {
		return int(intv)
	}
	L.TypeError(n, LTNumber)
	return 0
}

func (a *Args) CheckIntOrDefault(L *LState , n int , d int) int {
	v := a.LGet(L , n)
	if intv, ok := v.(LNumber); ok {
		return int(intv)
	}

	L.TypeError(n, LTNumber)
	return d
}

func (a *Args) CheckInt64(L *LState , n int) int64 {
	v := a.LGet(L , n)
	if intv, ok := v.(LNumber); ok {
		return int64(intv)
	}
	L.TypeError(n, LTNumber)
	return 0
}

func (a *Args) CheckNumber(L *LState , n int) LNumber {
	v := a.LGet(L , n)
	if lv, ok := v.(LNumber); ok {
		return lv
	}
	L.TypeError(n, LTNumber)
	return 0
}

func (a *Args) CheckString(L *LState , n int) string {
	v := a.LGet(L , n)
	if lv, ok := v.(LString); ok {
		return string(lv)
	} else if LVCanConvToString(v) {
		return LVAsString( v )
	}
	L.TypeError(n, LTString)
	return ""
}

func (a *Args) CheckBool(L *LState , n int) bool {
	v := a.LGet(L , n)
	if lv, ok := v.(LBool); ok {
		return bool(lv)
	}
	L.TypeError(n, LTBool)
	return false
}

func (a *Args) CheckTable(L *LState , n int) *LTable {
	v := a.LGet(L , n)
	if lv, ok := v.(*LTable); ok {
		return lv
	}
	L.TypeError(n, LTTable)
	return nil
}

func (a *Args) CheckFunction(L *LState , n int) *LFunction {
	v := a.LGet(L , n)
	if lv, ok := v.(*LFunction); ok {
		return lv
	}
	L.TypeError(n, LTFunction)
	return nil
}

func (a *Args) CheckUserData(L *LState , n int) *LUserData {
	v := a.LGet(L , n)
	if lv, ok := v.(*LUserData); ok {
		return lv
	}
	L.TypeError(n, LTUserData)
	return nil
}

func (a *Args) CheckLightUserData(L *LState , n int) *LightUserData {
	v := a.LGet(L , n)
	if lv, ok := v.(*LightUserData); ok {
		return lv
	}
	L.TypeError(n, LTLightUserData)
	return nil
}

func (a *Args) CheckThread(L *LState , n int) *LState {
	v := a.LGet(L , n)
	if lv, ok := v.(*LState); ok {
		return lv
	}
	L.TypeError(n, LTThread)
	return nil
}

func (a *Args) CheckType(L *LState , n int, typ LValueType) {
	v := a.LGet(L ,n)
	if v.Type() != typ {
		L.TypeError(n, typ)
	}
}

type GFunction struct {
	fn    func(*LState , *Args ) LValue
}

func NewGFunction(fn func(*LState, *Args ) LValue ) *GFunction {
	return &GFunction{fn }
}

func (gn *GFunction) String() string                     { return fmt.Sprintf("function: %p", gn) }
func (gn *GFunction) Type() LValueType                   { return LTGFunction}
func (gn *GFunction) assertFloat64() (float64, bool)     { return 0, false   }
func (gn *GFunction) assertString() (string, bool)       { return "", false  }
func (gn *GFunction) assertFunction() (*LFunction, bool) { return nil, false }
func (gn *GFunction) pcall(L *LState , reg *registry , RA int , nargs int , nret int) {

	if gn.fn == nil {
		L.RaiseError("invalid GFunction , got nil")
		return
	}

	var ret LValue
	args := argsPool.Get().(*Args)

	if nargs <= 0 {
		goto SET
	}

	for i := 1; i <= nargs; i++ {
		args.Set( reg.Get(RA + i) )
	}

SET:
	ret = gn.fn(L , args)
	args.reset()
	argsPool.Put(args)

	if nret != MultRet {
		reg.Set(RA, ret)
		reg.SetTop(RA + 1)
	}
}

//防止过多的方法定义
type TLightUserData struct {}
func (ud *TLightUserData) SetField(L *LState , key LValue, val LValue )  { }
func (ud *TLightUserData) GetField(L *LState , key LValue) LValue        { return nil }
func (ud *TLightUserData) Index(L *LState    ,key string ) LValue        { return nil }

func (ud *TLightUserData) NewIndex(L *LState , key string , val LValue)  { }
func (ud *TLightUserData) LCheck(obj interface{} , set LCallBack)  bool  { return false }
