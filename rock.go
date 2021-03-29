package lua

import (
	"errors"
	"fmt"
	"sync"
)

var (
	rock_not_found error = errors.New("not found")
	rock_json_null = []byte("null")
)

type Message interface {
	Byte() []byte
}

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

type ExUserKV struct {
	key string
	val LValue
}

type UserKV []ExUserKV
func (ukv *UserKV) Set(key string , val LValue ) {
	args := *ukv

	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if key == kv.key {
			kv.val = val
			return
		}
		if kv.key == "" {
			kv.val = val
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = key

		kv.val = val
		*ukv = args
		return
	}

	kv := ExUserKV{}
	kv.key = key
	kv.val = val
	*ukv = append(args, kv)
}
func (ukv *UserKV) Get(key string) LValue {
	args := *ukv
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if key == kv.key {
			return kv.val
		}
	}
	return LNil
}

func (ukv *UserKV) String() string                     { return fmt.Sprintf("function: %p", ukv) }
func (ukv *UserKV) Type() LValueType                   { return LTKEYVAL}
func (ukv *UserKV) assertFloat64() (float64, bool)     { return 0, false   }
func (ukv *UserKV) assertString() (string, bool)       { return "", false  }
func (ukv *UserKV) assertFunction() (*LFunction, bool) { return nil, false }


type LCallBack func( obj interface{} ) //用来回调方法

type rock  interface {
	Name() string
	Type() string
	Json() []byte

	SetField(*LState   , LValue, LValue )
	GetField(*LState   , LValue)  LValue

	Index(*LState      , string) LValue
	NewIndex(*LState   , string , LValue)

	LCheck(interface{} , LCallBack) bool //check(obj interface{}, set func) bool
	ToLightUserData(*LState) *LightUserData
}

type IO interface {
	rock
	Close()
	Start() error
	Write(interface{}) error
	Read() ([]byte , error )
}

type LightUserData struct {
	Value    rock
	ctx      ExData
}

func (ud *LightUserData) String() string                     { return fmt.Sprintf("userdata: %p", ud) }
func (ud *LightUserData) Type() LValueType                   { return LTLightUserData}
func (ud *LightUserData) assertFloat64() (float64, bool)     { return 0, false         }
func (ud *LightUserData) assertString() (string, bool)       { return "", false        }
func (ud *LightUserData) assertFunction() (*LFunction, bool) { return nil, false       }
func (ud *LightUserData) Get(key string) interface{}         { return ud.ctx.Get(key)  }
func (ud *LightUserData) Set(key string , v interface{} )    { ud.ctx.Set(key , v)     }

func (ud *LightUserData) CheckIO( L *LState ) IO {
	v , ok := ud.Value.(IO)
	if ok {
		return v
	}
	L.RaiseError("%s not IO , got: %s" , ud.Value.Name() , ud.Value.Type())
	return nil
}

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



type GFunction struct {
	fn    func(*LState , *Args ) LValue
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

	if ret != nil {
		reg.Set(RA, ret)
		reg.SetTop(RA + 1)
	}
}
