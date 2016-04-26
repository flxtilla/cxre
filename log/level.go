package log

import "strings"

type Level int

const (
	LUnrecognized Level = iota
	LPanic
	LFatal
	LError
	LWarn
	LInfo
	LDebug
)

var stringToLevel = map[string]Level{
	"panic": LPanic,
	"fatal": LFatal,
	"error": LError,
	"warn":  LWarn,
	"info":  LInfo,
	"debug": LDebug,
}

func StringToLevel(lv string) Level {
	if level, ok := stringToLevel[strings.ToLower(lv)]; ok {
		return level
	}
	return LUnrecognized
}

func (l Level) String() string {
	switch l {
	case LPanic:
		return "panic"
	case LFatal:
		return "fatal"
	case LError:
		return "error"
	case LWarn:
		return "warn"
	case LInfo:
		return "info"
	case LDebug:
		return "debug"
	}
	return "unrecognized"
}

func (lv Level) Color() string {
	switch lv {
	case LPanic:
		return red
	case LFatal:
		return magenta
	case LError:
		return cyan
	case LWarn:
		return yellow
	case LInfo:
		return green
	case LDebug:
		return blue
	}
	return white
}
