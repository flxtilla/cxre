package pack

import (
	"fmt"
	"io"
)

var (
	newline    = []byte{'\n'}
	dataindent = []byte{'\t', '\t'}
	space      = []byte{' '}
)

type ByteWriter struct {
	io.Writer
	c int
}

func (w *ByteWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	for n = range p {
		if w.c%12 == 0 {
			w.Writer.Write(newline)
			w.Writer.Write(dataindent)
			w.c = 0
		} else {
			w.Writer.Write(space)
		}

		fmt.Fprintf(w.Writer, "0x%02x,", p[n])
		w.c++
	}

	n++

	return
}
