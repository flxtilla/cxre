package log

import (
	"bytes"
	"fmt"
	"time"
)

type Entry interface {
	Logger
	Fielder
	Reader
	Created() time.Time
}

type Fielder interface {
	Fields() []Field
}

type Field struct {
	Key   string
	Value interface{}
}

func mkFields(v ...interface{}) []Field {
	var ret []Field
	for i, vv := range v {
		ret = append(ret, Field{fmt.Sprintf("Field%d", i), vv})
	}
	return ret
}

func mkFormatFields(format string, v ...interface{}) []Field {
	var ret []Field
	ret = append(ret, Field{"Format", format})
	ret = append(ret, mkFields(v...)...)
	return ret
}

type Reader interface {
	Read() (*bytes.Buffer, error)
}

type entry struct {
	created time.Time
	Reader
	Logger
	fields []Field
}

func newEntry(l Logger, f ...Field) Entry {
	return &entry{
		created: time.Now(),
		Logger:  l,
		fields:  f,
	}
}

func (e *entry) Read() (*bytes.Buffer, error) {
	s, err := e.Format(e)
	return bytes.NewBuffer(s), err
}

func (e *entry) Fields() []Field {
	return e.fields
}

func (e *entry) Created() time.Time {
	return e.created
}
