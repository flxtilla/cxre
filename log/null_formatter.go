package log

type NullFormatter struct{}

func DefaultNullFormatter() Formatter {
	return &NullFormatter{}
}

func (n *NullFormatter) Format(e Entry) ([]byte, error) {
	return nil, nil
}
