package lua


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
