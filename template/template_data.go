package template

import "github.com/flxtilla/cxre/state"

type TemplateData interface {
	state.State
	//HTML(string) template.HTML
	//CALL(string) interface{}
}

type templateData struct {
	state.State
	Data map[string]interface{}
}

func NewTemplateData(s state.State, in interface{}) TemplateData {
	ret := &templateData{
		State: s,
		Data:  make(map[string]interface{}),
	}
	if rcvd, ok := in.(map[string]interface{}); ok {
		for k, v := range rcvd {
			ret.Data[k] = v
		}
	} else {
		ret.Data["Any"] = in
	}
	return ret
}

//func (t templateData) HTML(name string) template.HTML {
//	res, err := t.State.Call(name)
//
//	if err != nil {
//		return template.HTML(err.Error())
//	}
//
//	if result, ok := res.(template.HTML); ok {
//		return result
//	}
//
//	return template.HTML(fmt.Sprintf("!: %s unprocessable by HTML</p>", name))
//}

//func (t templateData) CALL(name string) interface{} {
//	res, err := t.State.Call(name)
//	if err != nil {
//		return err.Error()
//	}
//	return res
//}
