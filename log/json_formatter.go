package log

type JsonFormatter struct {
	TimestampFormat string
}

func (j *JsonFormatter) Format(e Entry) ([]byte, error) {
	//data := make([]Field, len(e.Fields())+3)
	//for i, f := range e.Fields() {
	//	switch fv := f.Value.(type) {
	//	case error:
	//		// Otherwise errors are ignored by `encoding/json`
	//		// https://github.com/Sirupsen/logrus/issues/137
	//		data[i] = Field{f.Key, fv.Error()}
	//	default:
	//		data[i] = f
	//	}
	//}
	//prefixFieldClashes(data)

	//timestampFormat := f.TimestampFormat
	//if timestampFormat == "" {
	//	timestampFormat = DefaultTimestampFormat
	//}

	//data["time"] = entry.Time.Format(timestampFormat)
	//data["msg"] = entry.Message
	//data["level"] = entry.Level.String()

	//serialized, err := json.Marshal(data)
	//if err != nil {
	//	return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	//}
	//return append(serialized, '\n'), nil
	return nil, nil
}
