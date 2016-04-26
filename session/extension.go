package session

/*import (
	"github.com/thrisp/flotilla/extension"
	"github.com/thrisp/flotilla/session"
	"github.com/thrisp/flotilla/state"
)

func mkFunction(k string, v interface{}) extension.Function {
	return extension.NewFunction(k, v)
}

var sessionFns = []extension.Function{
	mkFunction("delete_session", deleteSession),
	mkFunction("get_session", getSession),
	mkFunction("session", returnSession),
	mkFunction("set_session", setSession),
}

var SessionFxtension extension.Fxtension = extension.New("Session_Fxtension", sessionFns...)

func deleteSession(s state.State, key string) error {
	return s.Delete(key)
}

func getSession(s state.State, key string) interface{} {
	return s.Get(key)
}

func SessionStore(s state.State) SessionStore {
	ss, _ := s.Call("session")
	return ss.(session.SessionStore)
}

func returnSession(s state.State) SessionStore {
	return s
}

func setSession(s state.State, key string, value interface{}) error {
	return s.Set(key, value)
}
*/
