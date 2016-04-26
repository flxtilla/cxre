package log

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"
)

type TextFormatter struct {
	Color           bool
	TimestampFormat string
	Sort            bool
}

func DefaultTextFormatter() Formatter {
	return &TextFormatter{
		true,
		time.StampNano,
		false,
	}
}

func (t *TextFormatter) Format(e Entry) ([]byte, error) {
	fs := e.Fields()
	var keys []string = make([]string, 0, len(fs))
	for _, k := range fs {
		keys = append(keys, k.Key)
	}

	if t.Sort {
		sort.Strings(keys)
	}

	timestampFormat := t.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampNano
	}

	b := &bytes.Buffer{}

	if t.Color {
		formatColorFields(b, e, keys, timestampFormat)
	} else {
		formatFields(b, e, keys, timestampFormat)
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func formatColorFields(b *bytes.Buffer, e Entry, keys []string, timestampFormat string) {
	lvl := e.Level()
	lvlColor := lvl.Color()
	lvlText := strings.ToUpper(lvl.String())
	fmt.Fprintf(b, "%s %s %s ", lvlColor, lvlText, reset)

	timestamp := time.Now().Format(timestampFormat)
	fmt.Fprintf(b, "%s| %v |%s ", red, timestamp, reset)

	fds := e.Fields()
	for _, v := range fds {
		for _, vv := range keys {
			if v.Key == vv {
				fmt.Fprintf(b, "%v", v.Value)
			}
		}
	}
}

func formatFields(b *bytes.Buffer, e Entry, keys []string, timestampFormat string) {
	lvl := e.Level()
	lvlText := strings.ToUpper(lvl.String())
	fmt.Fprintf(b, "[%s]", lvlText)

	timestamp := time.Now().Format(timestampFormat)
	fmt.Fprintf(b, " | %v | ", timestamp)

	fds := e.Fields()
	for _, v := range fds {
		for _, vv := range keys {
			if v.Key == vv {
				fmt.Fprintf(b, "%v", v.Value)
			}
		}
	}
}

//func LogFmt(s State) string {
//st := s.RStatus
//md := s.RMethod
//return fmt.Sprintf("%v |%s %3d %s| %12v | %s |%s %s %-7s %s",
//s.RStop.Format("2006/01/02 - 15:04:05"),
//StatusColor(st), st, reset,
//s.RLatency,
//s.RRequester,
//MethodColor(md), reset, md,
//s.RPath,
//)
//}

//func StatusColor(code int) (color string) {
//	switch {
//	case code >= 200 && code <= 299:
//		color = green
//	case code >= 300 && code <= 399:
//		color = white
//	case code >= 400 && code <= 499:
//		color = yellow
//	default:
//		color = red
//	}
//	return color
//}

//func MethodColor(method string) (color string) {
//	switch {
//	case method == "GET":
//		color = blue
//	case method == "POST":
//		color = cyan
//	case method == "PUT":
//		color = yellow
//	case method == "DELETE":
//		color = red
//	case method == "PATCH":
//		color = green
//	case method == "HEAD":
//		color = magenta
//	case method == "OPTIONS":
//		color = white
//	}
//	return color
//}
