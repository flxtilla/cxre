package status

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"github.com/flxtilla/cxre/state"
	"github.com/flxtilla/cxre/xrr"
)

const statusText = `%d %s`

const panicBlock = `<h1>%s</h1>
<pre style="font-weight: bold;">%s</pre>
`
const panicHtml = `<html>
<head><title>Internal Server Error</title>
<style type="text/css">
html, body {
font-family: "Roboto", sans-serif;
color: #333333;
margin: 0px;
}
h1 {
color: #2b3848;
background-color: #ffffff;
padding: 20px;
border-bottom: 1px dashed #2b3848;
}
pre {
font-size: 1.1em;
margin: 20px;
padding: 20px;
border: 2px solid #2b3848;
background-color: #ffffff;
}
pre p:nth-child(odd){margin:0;}
pre p:nth-child(even){background-color: rgba(216,216,216,0.25); margin: 0;}
</style>
</head>
<body>
%s
</body>
</html>
`

type Status interface {
	Code() int
	Managers() []state.Manage
}

type status struct {
	code     int
	managers []state.Manage
}

func newStatus(code int, m ...state.Manage) *status {
	st := &status{code: code}
	st.managers = []state.Manage{st.first}
	st.managers = append(st.managers, m...)
	st.managers = append(st.managers, st.panics, st.last)
	return st
}

func (s *status) Code() int {
	return s.code
}

func (s *status) Managers() []state.Manage {
	return s.managers
}

func (st status) first(s state.State) {
	s.Call("header_write", st.code)
}

func panicServe(s state.State, b bytes.Buffer) {
	servePanic := fmt.Sprintf(panicHtml, b.String())
	_, _ = s.Call("header_modify", "set", []string{"Content-Type", "text/html"})
	_, _ = s.Call("write_to_response", servePanic)
}

func panics(s state.State) xrr.Xrrors {
	return s.Errors().ByType(xrr.ErrorTypePanic)
}

func panicLogToBuffer(s state.State) bytes.Buffer {
	var auffer bytes.Buffer
	for _, p := range panics(s) {
		reader := bufio.NewReader(bytes.NewReader([]byte(fmt.Sprintf("%s", p.Meta))))
		lineno := 0
		var buffer bytes.Buffer
		var cuffer bytes.Buffer
		cuffer.WriteString(fmt.Sprintf("%s\n", p.Error()))
		var err error
		for err == nil {
			lineno++
			l, _, err := reader.ReadLine()
			cuffer.WriteString(fmt.Sprintf("%s\n", l))
			if lineno%2 == 0 {
				buffer.WriteString(fmt.Sprintf("\n%s</p>\n", l))
			} else {
				buffer.WriteString(fmt.Sprintf("<p>%s\n", l))
			}
			if err != nil {
				break
			}
		}
		s.Println(cuffer.String())
		pb := fmt.Sprintf(panicBlock, p.Error(), buffer.String())
		auffer.WriteString(pb)
	}
	return auffer
}

func isWritten(s state.State) bool {
	return s.RWriter().Written()
}

func modeIs(s state.State, is string) bool {
	m, _ := s.Call("mode_is", is)
	return m.(bool)
}

func (st status) panics(s state.State) {
	if st.code == 500 && !isWritten(s) {
		if !modeIs(s, "production") {
			panicServe(s, panicLogToBuffer(s))
		}
	}
}

func (st status) last(s state.State) {
	s.Push(func(ps state.State) {
		if !isWritten(ps) {
			code := st.Code()
			_, _ = ps.Call("write_to_response", fmt.Sprintf(statusText, code, http.StatusText(code)))
		}
	})
}
