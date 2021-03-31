package lua

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

const (
	INIT  LightUserDataStatus = iota
	RUNNING
	CLOSE
	PANIC
)

var ERR = errors.New("not function")
var NULL = []byte("")

//type userdata Status

type LightUserDataStatus int
var LightUserDataStatusValue = [4]string{"INIT" , "RUNNING" , "CLOSE" , "PANIC"}

func (us LightUserDataStatus) String() string {
	return LightUserDataStatusValue[int(us)]
}

//JSON
type jsonBuffer struct {
	buff bytes.Buffer
}

func (jb *jsonBuffer) Write( v []byte ) {
	len := len(v)
	var ch byte
	for i := 0;i<len;i++ {
		ch = v[i]
		switch ch {
		case '"':
			jb.buff.WriteByte('\\')
			jb.buff.WriteByte('"')
		case '\\':
			jb.buff.WriteByte('\\')
			jb.buff.WriteByte('\\')
		case '\r':
			jb.buff.WriteByte('\\')
			jb.buff.WriteByte('r')
		case '\n':
			jb.buff.WriteByte('\\')
			jb.buff.WriteByte('n')
		case '\t':
			jb.buff.WriteByte('\\')
			jb.buff.WriteByte('t')
		default:
			jb.buff.WriteByte(ch)
		}
	}
}

func (jb *jsonBuffer) WriteKey(key string) {
	jb.buff.WriteByte('"')
	jb.Write(S2B(key))
	jb.buff.WriteByte('"')
	jb.buff.WriteByte(':')
}

func (jb *jsonBuffer) WriteVal(val string) {
	jb.buff.WriteByte('"')
	jb.Write(S2B(val))
	jb.buff.WriteByte('"')
}

func (jb *jsonBuffer) WriteInt(val int) {
	jb.buff.WriteString(strconv.Itoa(val))
}

//写 key: val
func (jb *jsonBuffer) WriteKV(key string , val string , end bool) {
	jb.WriteKey(key)
	jb.WriteVal(val)

	if end { return }

	jb.buff.WriteByte(',')
}

//写 key: int
func (jb *jsonBuffer) WriteKI(key string , val int , end bool) {
	jb.WriteKey(key)
	jb.WriteInt(val)
	if end { return }
	jb.buff.WriteByte(',')
}

//写 key: OBJ
func (jb *jsonBuffer) WriteKO(key string , obj interface{ Name() string } , end bool) {
	jb.WriteKey( key )

	jb.buff.WriteByte('{')
	jb.buff.WriteString(`"type":"userdata","value":`)
	jb.WriteVal( obj.Name() )
	jb.buff.WriteByte('}')
	if end { return }
	jb.buff.WriteByte(',')
}

func (jb *jsonBuffer) Bytes() []byte {
	return jb.buff.Bytes()
}

func (jb *jsonBuffer) Start( name string ) {
	jb.buff.WriteByte('{')
	jb.WriteKey(name)
	jb.buff.WriteByte('{')
}

func (jb *jsonBuffer) End() {
	jb.buff.WriteByte('}')
	jb.buff.WriteByte('}')
}

func NewJsonBuffer() *jsonBuffer {
	return &jsonBuffer{buff: bytes.Buffer{}}
}

//防止过多的方法定义
type Super struct {}
func (s *Super) SetField(L *LState , key LValue, val LValue )  { }
func (s *Super) GetField(L *LState , key LValue) LValue        { return LNil }

func (s *Super) Index(L *LState    ,key string ) LValue        { return LNil }
func (s *Super) NewIndex(L *LState , key string , val LValue)  { }

func (s *Super) LCheck(obj interface{} , set LCallBack)  bool  { return false }
func (s *Super) ToLightUserData(L *LState ) *LightUserData     { return L.NewLightUserData(s) }

func (s *Super) Name() string                          { return "super"           }
func (s *Super) Type() string                          { return "super"           }
func (s *Super) Close() error                          { return ERR               }
func (s *Super) Start() error                          { return ERR               }
func (s *Super) Write( v interface{} ) error           { return ERR               }
func (s *Super) Read() ([]byte , error)                { return nil , ERR         }
func (s *Super) ToJson() ([]byte , error)              { return nil , ERR         }
func (s *Super) Status() (string , error)              { return "name:super" , ERR}

func IsNotFound( err error ) bool {
	if err.Error() == "not found" {
		return true
	}
	return false
}

func IsNull(v []byte) bool {
	if len(v) == 0 {
		return true
	}
	return false
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

func (a *Args) CheckIO(L *LState , n int) IO {
	ud := a.CheckLightUserData( L , n )

	v , ok := ud.Value.(IO)
	if ok {
		return v
	}

	L.RaiseError("#%d must be IO , got fail" , n)
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

func NewGFunction(fn func(*LState, *Args ) LValue ) *GFunction {
	return &GFunction{fn }
}


func CheckInt(L *LState , lv LValue) int {
	if intv, ok := lv.(LNumber); ok {
		return int(intv)
	}
	L.RaiseError("must be int , got %s" , lv.Type().String())
	return 0
}

func CheckIntOrDefault(L *LState , lv LValue , d int) int {
	if intv, ok := lv.(LNumber); ok {
		return int(intv)
	}
	return d
}

func CheckInt64(L *LState , lv LValue) int64 {
	if intv, ok := lv.(LNumber); ok {
		return int64(intv)
	}
	L.RaiseError("must be int64 , got %s" , lv.Type().String())
	return 0
}

func CheckNumber(L *LState , lv LValue) LNumber {
	if lv, ok := lv.(LNumber); ok {
		return lv
	}
	L.RaiseError("must be LNumber , got %s" , lv.Type().String())
	return 0
}

func CheckString(L *LState , lv LValue ) string {
	if lv, ok := lv.(LString); ok {
		return string(lv)
	} else if LVCanConvToString(lv) {
		return LVAsString( lv )
	}
	return ""
}

func CheckBool(L *LState , lv LValue) bool {
	if lv, ok := lv.(LBool); ok {
		return bool(lv)
	}

	L.RaiseError("must be bool , got %s" , lv.Type().String())
	return false
}

func CheckTable(L *LState , lv LValue ) *LTable {
	if lv, ok := lv.(*LTable); ok {
		return lv
	}
	L.RaiseError("must be LTable, got %s" , lv.Type().String())
	return nil
}

func  CheckFunction(L *LState , lv LValue ) *LFunction {
	if lv, ok := lv.(*LFunction); ok {
		return lv
	}
	L.RaiseError("must be Function, got %s" , lv.Type().String())
	return nil
}

func  CheckUserData(L *LState , lv LValue ) *LUserData {
	if lv, ok := lv.(*LUserData); ok {
		return lv
	}
	L.RaiseError("must be UserData, got %s" , lv.Type().String())
	return nil
}

func  CheckLightUserData(L *LState , lv LValue ) *LightUserData {
	if lv, ok := lv.(*LightUserData); ok {
		return lv
	}
	L.RaiseError("must be LightUserData, got %s" , lv.Type().String())
	return nil
}

func  CheckIO(L *LState , lv LValue) IO {
	if lv.Type() != LTLightUserData {
		L.RaiseError("must be LightUserData, got %s" , lv.Type().String())
		return nil
	}

	v , ok := lv.(*LightUserData).Value.(IO)
	if ok {
		return v
	}

	L.RaiseError("must be IO , got %s" , lv.Type().String())
	return nil
}

func  CheckThread(L *LState , lv LValue ) *LState {
	if lv, ok := lv.(*LState); ok {
		return lv
	}
	L.RaiseError("must be thread , got %s" , lv.Type().String())
	return nil
}

func  CheckType(L *LState , lv LValue , typ LValueType) {
	if lv.Type() != typ {
		L.RaiseError("must be %" , typ.String() , lv.Type().String())
	}
}

func S2B(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return
}

func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}