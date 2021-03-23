package lua

import (
	"context"
	"fmt"
	"os"
)

type LValueType int

const (
	LTNil LValueType = iota
	LTBool
	LTNumber
	LTString
	LTFunction
	LTUserData
	LTThread
	LTTable
	LTChannel
	LTLightUserData
	LTGFunction
)

var lValueNames = [12]string{"nil", "boolean", "number", "string", "function", "userdata", "thread", "table", "channel" , "ligthuserdata"  , "GFunction"}

func (vt LValueType) String() string {
	return lValueNames[int(vt)]
}

type LValue interface {
	String() string
	Type() LValueType
	// to reduce `runtime.assertI2T2` costs, this method should be used instead of the type assertion in heavy paths(typically inside the VM).
	assertFloat64() (float64, bool)
	// to reduce `runtime.assertI2T2` costs, this method should be used instead of the type assertion in heavy paths(typically inside the VM).
	assertString() (string, bool)
	// to reduce `runtime.assertI2T2` costs, this method should be used instead of the type assertion in heavy paths(typically inside the VM).
	assertFunction() (*LFunction, bool)
}

type Logger interface {
    Err(format string, v ...interface{})
	Info(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

// LVIsFalse returns true if a given LValue is a nil or false otherwise false.
func LVIsFalse(v LValue) bool { return v == LNil || v == LFalse }

// LVIsFalse returns false if a given LValue is a nil or false otherwise true.
func LVAsBool(v LValue) bool { return v != LNil && v != LFalse }

// LVAsString returns string representation of a given LValue
// if the LValue is a string or number, otherwise an empty string.
func LVAsString(v LValue) string {
	switch sn := v.(type) {
	case LString, LNumber:
		return sn.String()
	default:
		return ""
	}
}

// LVCanConvToString returns true if a given LValue is a string or number
// otherwise false.
func LVCanConvToString(v LValue) bool {
	switch v.(type) {
	case LString, LNumber:
		return true
	default:
		return false
	}
}

// LVAsNumber tries to convert a given LValue to a number.
func LVAsNumber(v LValue) LNumber {
	switch lv := v.(type) {
	case LNumber:
		return lv
	case LString:
		if num, err := parseNumber(string(lv)); err == nil {
			return num
		}
	}
	return LNumber(0)
}

type LNilType struct{}

func (nl *LNilType) String() string                     { return "nil" }
func (nl *LNilType) Type() LValueType                   { return LTNil }
func (nl *LNilType) assertFloat64() (float64, bool)     { return 0, false }
func (nl *LNilType) assertString() (string, bool)       { return "", false }
func (nl *LNilType) assertFunction() (*LFunction, bool) { return nil, false }

var LNil = LValue(&LNilType{})

type LBool bool

func (bl LBool) String() string {
	if bool(bl) {
		return "true"
	}
	return "false"
}
func (bl LBool) Type() LValueType                   { return LTBool }
func (bl LBool) assertFloat64() (float64, bool)     { return 0, false }
func (bl LBool) assertString() (string, bool)       { return "", false }
func (bl LBool) assertFunction() (*LFunction, bool) { return nil, false }

var LTrue = LBool(true)
var LFalse = LBool(false)

type LString string

func (st LString) String() string                     { return string(st) }
func (st LString) Type() LValueType                   { return LTString }
func (st LString) assertFloat64() (float64, bool)     { return 0, false }
func (st LString) assertString() (string, bool)       { return string(st), true }
func (st LString) assertFunction() (*LFunction, bool) { return nil, false }

// fmt.Formatter interface
func (st LString) Format(f fmt.State, c rune) {
	switch c {
	case 'd', 'i':
		if nm, err := parseNumber(string(st)); err != nil {
			defaultFormat(nm, f, 'd')
		} else {
			defaultFormat(string(st), f, 's')
		}
	default:
		defaultFormat(string(st), f, c)
	}
}

func (nm LNumber) String() string {
	if isInteger(nm) {
		return fmt.Sprint(int64(nm))
	}
	return fmt.Sprint(float64(nm))
}

func (nm LNumber) Type() LValueType                   { return LTNumber }
func (nm LNumber) assertFloat64() (float64, bool)     { return float64(nm), true }
func (nm LNumber) assertString() (string, bool)       { return "", false }
func (nm LNumber) assertFunction() (*LFunction, bool) { return nil, false }

// fmt.Formatter interface
func (nm LNumber) Format(f fmt.State, c rune) {
	switch c {
	case 'q', 's':
		defaultFormat(nm.String(), f, c)
	case 'b', 'c', 'd', 'o', 'x', 'X', 'U':
		defaultFormat(int64(nm), f, c)
	case 'e', 'E', 'f', 'F', 'g', 'G':
		defaultFormat(float64(nm), f, c)
	case 'i':
		defaultFormat(int64(nm), f, 'd')
	default:
		if isInteger(nm) {
			defaultFormat(int64(nm), f, c)
		} else {
			defaultFormat(float64(nm), f, c)
		}
	}
}

type LTable struct {
	Metatable LValue

	array   []LValue
	dict    map[LValue]LValue
	strdict map[string]LValue
	keys    []LValue
	k2i     map[LValue]int
}

func (tb *LTable) String() string                     { return fmt.Sprintf("table: %p", tb) }
func (tb *LTable) Type() LValueType                   { return LTTable }
func (tb *LTable) assertFloat64() (float64, bool)     { return 0, false }
func (tb *LTable) assertString() (string, bool)       { return "", false }
func (tb *LTable) assertFunction() (*LFunction, bool) { return nil, false }

type LFunction struct {
	IsG       bool
	Env       *LTable
	Proto     *FunctionProto
	GFunction LGFunction
	Upvalues  []*Upvalue
}

type LGFunction func(*LState) int
func (fn *LFunction) String() string                     { return fmt.Sprintf("function: %p", fn) }
func (fn *LFunction) Type() LValueType                   { return LTFunction }
func (fn *LFunction) assertFloat64() (float64, bool)     { return 0, false }
func (fn *LFunction) assertString() (string, bool)       { return "", false }
func (fn *LFunction) assertFunction() (*LFunction, bool) { return fn, true}

type GFunction struct {
	fn    func(*LState , []LValue) LValue
}

func NewGFunction(fn func(*LState, []LValue ) LValue ) *GFunction {
	return &GFunction{fn}
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
	if nargs <= 0 {
		ret = gn.fn(L , nil)
		reg.SetTop( RA )
		return
	}

	args := make([]LValue , nargs)
	for i := 1;i < nargs ; i++ {
		args[i] = reg.Get(RA + i)
	}

	ret = gn.fn(L , args)
	if nret != MultRet {
		reg.Set(RA , ret)
	}
}

type Global struct {
	MainThread    *LState
	CurrentThread *LState
	Registry      *LTable
	Global        *LTable

	builtinMts map[int]LValue
	tempFiles  []*os.File
	gccount    int32
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

type LState struct {
	G       *Global
	Parent  *LState
	Env     *LTable
	Panic   func(*LState)
	Dead    bool
	Options Options

	stop         int32
	Reg          *registry
	stack        callFrameStack
	alloc        *allocator
	currentFrame *callFrame
	wrapped      bool
	uvcache      *Upvalue
	hasErrorFunc bool
	mainLoop     func(*LState, *callFrame)
	ctx          context.Context

	ExData   ExData
}

func (ls *LState) String() string                     { return fmt.Sprintf("thread: %p", ls) }
func (ls *LState) Type() LValueType                   { return LTThread }
func (ls *LState) assertFloat64() (float64, bool)     { return 0, false }
func (ls *LState) assertString() (string, bool)       { return "", false }
func (ls *LState) assertFunction() (*LFunction, bool) { return nil, false }

type LUserData struct {
	Value     interface{}
	Env       *LTable
	Metatable LValue
}

func (ud *LUserData) String() string                     { return fmt.Sprintf("userdata: %p", ud) }
func (ud *LUserData) Type() LValueType                   { return LTUserData }
func (ud *LUserData) assertFloat64() (float64, bool)     { return 0, false }
func (ud *LUserData) assertString() (string, bool)       { return "", false }
func (ud *LUserData) assertFunction() (*LFunction, bool) { return nil, false }


type LCallBack func( obj interface{} ) //用来回调方法
type luaSetGetFunc  interface {
	LSet(*LState , LValue, LValue )
	LGet(*LState , LValue)  LValue
	LCheck(interface{} , LCallBack) bool //check(obj interface{}, set func) bool
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


type LChannel chan LValue

func (ch LChannel) String() string                     { return fmt.Sprintf("channel: %p", ch) }
func (ch LChannel) Type() LValueType                   { return LTChannel }
func (ch LChannel) assertFloat64() (float64, bool)     { return 0, false }
func (ch LChannel) assertString() (string, bool)       { return "", false }
func (ch LChannel) assertFunction() (*LFunction, bool) { return nil, false }