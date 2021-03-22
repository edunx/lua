package lua

type TLightUserData struct {}
func (tud *TLightUserData) LSet(L *LState , key LValue, val LValue )        { }
func (tud *TLightUserData) LGet(L *LState , key LValue) LValue              { return nil }
func (tud *TLightUserData) LCheck(obj interface{} , set LCallBack)  bool    { return false }

