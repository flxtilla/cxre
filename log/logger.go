package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type StdLogger interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

type Mutex interface {
	Lock()
	Unlock()
}

type Logger interface {
	io.Writer
	StdLogger
	Log(Level, Entry)
	Mutex
	Level() Level
	Formatter
	Hooks
}

type Formatter interface {
	Format(Entry) ([]byte, error)
}

type logger struct {
	io.Writer
	level Level
	Formatter
	Hooks
	sync.Mutex
}

func New(w io.Writer, l Level, f Formatter) Logger {
	return &logger{
		Writer:    w,
		level:     l,
		Formatter: f,
		Hooks:     &hooks{},
	}
}

func (l *logger) Level() Level {
	return l.level
}

func (l *logger) Log(lv Level, e Entry) {
	log(lv, e)
}

func log(lv Level, e Entry) {
	if err := e.Fire(lv, e); err != nil {
		e.Lock()
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
		e.Unlock()
	}

	reader, err := e.Read()
	if err != nil {
		e.Lock()
		fmt.Fprintf(os.Stderr, "Failed to obtain reader: %v\n", err)
		e.Unlock()
	}

	e.Lock()
	defer e.Unlock()

	_, err = io.Copy(e, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}

	if lv <= LPanic {
		panic(&e)
	}
}

func (l *logger) Fatal(v ...interface{}) {
	if l.level == LFatal {
		log(LFatal, newEntry(l, mkFields(v...)...))
		os.Exit(1)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	if l.level == LFatal {
		log(LFatal, newEntry(l, mkFormatFields(format, v...)...))
		os.Exit(1)
	}
}

func (l *logger) Fatalln(v ...interface{}) {
	if l.level == LFatal {
		log(LFatal, newEntry(l, mkFields(v...)...))
		os.Exit(1)
	}
}

func (l *logger) Panic(v ...interface{}) {
	if l.level == LPanic {
		log(LPanic, newEntry(l, mkFields(v...)...))
	}
	panic(fmt.Sprint(v...))
}

func (l *logger) Panicf(format string, v ...interface{}) {
	if l.level == LPanic {
		log(LPanic, newEntry(l, mkFormatFields(format, v...)...))
	}
	panic(fmt.Sprintf(format, v...))
}

func (l *logger) Panicln(v ...interface{}) {
	l.Panic(v...)
}

func (l *logger) Print(v ...interface{}) {
	if l.level >= LError {
		log(LInfo, newEntry(l, mkFields(v...)...))
	}
}

func (l *logger) Printf(format string, v ...interface{}) {
	if l.level >= LError {
		log(LInfo, newEntry(l, mkFormatFields(format, v...)...))
	}
}

func (l *logger) Println(v ...interface{}) {
	if l.level >= LError {
		log(LInfo, newEntry(l, mkFields(v...)...))
	}
}
