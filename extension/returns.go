package extension

type Returns interface {
	CallString(string, ...interface{}) string
	CallInteger(string, ...interface{}) int
	CallBoolean(string, ...interface{}) bool
}

type returns struct {
	e Extension
}

func (r *returns) CallString(name string, arg ...interface{}) string {
	var ret string
	var ok bool
	res := r.e.MustCall(name, arg...)
	if ret, ok = res.(string); !ok {
		panic(NotExpectedReturn(ret, "string").Error())
	}
	return ret
}

func (r *returns) CallInteger(name string, arg ...interface{}) int {
	var ret int
	var ok bool
	res := r.e.MustCall(name, arg...)
	if ret, ok = res.(int); !ok {
		panic(NotExpectedReturn(ret, "integer").Error())
	}
	return ret
}

func (r *returns) CallBoolean(name string, arg ...interface{}) bool {
	var ret bool
	var ok bool
	res := r.e.MustCall(name, arg...)
	if ret, ok = res.(bool); !ok {
		panic(NotExpectedReturn(ret, "boolean").Error())
	}
	return ret
}
