package asset

import "fmt"

type assetError struct {
	err  string
	vals []interface{}
}

func (m *assetError) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(m.err, m.vals...))
}

func (m *assetError) Out(vals ...interface{}) *assetError {
	m.vals = vals
	return m
}

func Xrror(err string) *assetError {
	return &assetError{err: err}
}
