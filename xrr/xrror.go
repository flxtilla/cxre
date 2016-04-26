package xrr

import (
	"bytes"
	"fmt"
)

type Xrror struct {
	Err        string      `json:"error"`
	Type       uint32      `json:"-"`
	Meta       interface{} `json:"meta"`
	parameters []interface{}
}

func (x *Xrror) Error() string {
	return fmt.Sprintf(x.Err, x.parameters...)
}

func (x *Xrror) Out(p ...interface{}) *Xrror {
	x.parameters = p
	return x
}

func NewXrror(err string, params ...interface{}) *Xrror {
	return &Xrror{Err: err, parameters: params, Type: ErrorTypeFlotilla}
}

type Xrrors []*Xrror

func (a Xrrors) ByType(typ uint32) Xrrors {
	if len(a) == 0 {
		return a
	}
	result := make(Xrrors, 0, len(a))
	for _, msg := range a {
		if msg.Type&typ > 0 {
			result = append(result, msg)
		}
	}
	return result
}

func (a Xrrors) String() string {
	if len(a) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	for i, msg := range a {
		text := fmt.Sprintf("Error #%02d: %s\nMeta: %v\n", (i + 1), msg.Error(), msg.Meta)
		buffer.WriteString(text)
	}
	return buffer.String()
}
