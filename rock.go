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
		ret = gn.fn(L, nil)
		reg.SetTop(RA)
		return
	}

	for i := 1; i <= nargs; i++ {
		args.Set( reg.Get(RA + i) )
	}

	if nret != MultRet {
		reg.Set(RA, ret)
	}

	args.reset()
	argsPool.Put(args)
}

//防止过多的方法定义
type TLightUserData struct {}
func (ud *TLightUserData) SetField(L *LState , key LValue, val LValue )  { }
func (ud *TLightUserData) GetField(L *LState , key LValue) LValue        { return nil }
func (ud *TLightUserData) Index(L *LState    ,key string ) LValue        { return nil }

func (ud *TLightUserData) NewIndex(L *LState , key string , val LValue)  { }
func (ud *TLightUserData) LCheck(obj interface{} , set LCallBack)  bool  { return false }
