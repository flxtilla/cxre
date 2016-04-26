package store

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/flxtilla/cxre/xrr"
)

type Store interface {
	Load(string) error
	LoadByte([]byte, string) error
	Add(string, string)
	Returnr
}

type Returnr interface {
	String(string) string
	List(string) []string
	Bool(string) bool
	Float(string) float64
	Int(string) int
	Int64(string) int64
}

type store map[string]map[string]StoreItem

func New() Store {
	return make(store)
}

func (s store) query(key string) StoreItem {
	var sec, seckey string
	base := strings.Split(key, "_")
	if len(base) == 1 {
		sec, seckey = "", strings.ToUpper(base[0])
	} else {
		sec, seckey = strings.ToUpper(base[0]), strings.ToUpper(base[1])
	}
	if k, ok := s[strings.ToUpper(sec)]; ok {
		if i, ok := k[seckey]; ok {
			return i
		}
	}
	return &storeItem{}
}

func (s store) String(key string) string {
	i := s.query(key)
	return i.String()
}

func (s store) Bool(key string) bool {
	i := s.query(key)
	return i.Bool()
}

func (s store) Float(key string) float64 {
	i := s.query(key)
	return i.Float()
}

func (s store) Int(key string) int {
	i := s.query(key)
	return i.Int()
}

func (s store) Int64(key string) int64 {
	i := s.query(key)
	return i.Int64()
}

func (s store) List(key string) []string {
	i := s.query(key)
	return i.List()
}

func (s store) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	return s.parse(reader, filename)
}

func (s store) LoadByte(b []byte, name string) error {
	reader := bufio.NewReader(bytes.NewReader(b))
	return s.parse(reader, name)
}

var StoreParseError = xrr.NewXrror("Store configuration parsing: syntax error at '%s:%d'.").Out

func (s store) parse(reader *bufio.Reader, filename string) (err error) {
	lineno := 0
	section := ""
	for err == nil {
		l, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineno++
		if len(l) == 0 {
			continue
		}
		line := strings.TrimFunc(string(l), unicode.IsSpace)
		for line[len(line)-1] == '\\' {
			line = line[:len(line)-1]
			l, _, err := reader.ReadLine()
			if err != nil {
				return err
			}
			line += strings.TrimFunc(string(l), unicode.IsSpace)
		}
		section, err = s.parseLine(section, line)
		if err != nil {
			return StoreParseError(filename, lineno)
		}
	}
	return err
}

var (
	regDoubleQuote = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*\"([^\"]*)\"$")
	regSingleQuote = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*'([^']*)'$")
	regNoQuote     = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*([^#;]+)")
	regNoValue     = regexp.MustCompile("^([^= \t]+)[ \t]*=[ \t]*([#;].*)?")
)

func (s store) parseLine(section, line string) (string, error) {
	if line[0] == '#' || line[0] == ';' {
		return section, nil
	}

	if line[0] == '[' && line[len(line)-1] == ']' {
		section := strings.TrimFunc(line[1:len(line)-1], unicode.IsSpace)
		section = strings.ToLower(section)
		return section, nil
	}

	if m := regDoubleQuote.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], m[0][2])
		return section, nil
	} else if m = regSingleQuote.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], m[0][2])
		return section, nil
	} else if m = regNoQuote.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], strings.TrimFunc(m[0][2], unicode.IsSpace))
		return section, nil
	} else if m = regNoValue.FindAllStringSubmatch(line, 1); m != nil {
		s.add(section, m[0][1], "")
		return section, nil
	}
	return section, errors.New("line parse error")
}

func (s store) Add(key, value string) {
	sl := strings.Split(key, "_")
	if len(sl) > 1 {
		section, label := sl[0], strings.Join(sl[1:], "_")
		s.add(section, label, value)
	} else {
		s.add("", sl[0], value)
	}
}

func (s store) add(section, key, value string) {
	sec, seckey := strings.ToUpper(section), strings.ToUpper(key)
	if _, ok := s[sec]; !ok {
		s[sec] = make(map[string]StoreItem)
	}
	s[sec][seckey] = newItem(seckey, value)
}

type StoreItem interface {
	String() string
	Bool() bool
	Float() float64
	Int() int
	Int64() int64
	List() []string
}

type storeItem struct {
	Key   string
	Value string
}

func newItem(key, value string) *storeItem {
	return &storeItem{key, value}
}

func (i *storeItem) String() string {
	return i.Value
}

var boolString = map[string]bool{
	"t":     true,
	"true":  true,
	"y":     true,
	"yes":   true,
	"on":    true,
	"1":     true,
	"f":     false,
	"false": false,
	"n":     false,
	"no":    false,
	"off":   false,
	"0":     false,
}

func (i *storeItem) Bool() bool {
	if value, ok := boolString[strings.ToLower(i.Value)]; ok {
		return value
	}
	return false
}

func (i *storeItem) Float() float64 {
	if value, err := strconv.ParseFloat(i.Value, 64); err == nil {
		return value
	}
	return 0.0
}

func (i *storeItem) Int() int {
	if value, err := strconv.Atoi(i.Value); err == nil {
		return value
	}
	return 0
}

func (i *storeItem) Int64() int64 {
	if value, err := strconv.ParseInt(i.Value, 10, 64); err == nil {
		return value
	}
	return -1
}

func (i *storeItem) List() []string {
	if i.Value == "" {
		return nil
	}
	return strings.Split(i.Value, ",")
}
