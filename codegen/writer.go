package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
)

var (
	lf   = []byte{'\n'}
	crlf = []byte{'\r', '\n'}
)

var (
	tabs = [...][]byte{
		0:  {},
		1:  {'\t'},
		2:  {'\t', '\t'},
		3:  {'\t', '\t', '\t'},
		4:  {'\t', '\t', '\t', '\t'},
		5:  {'\t', '\t', '\t', '\t', '\t'},
		6:  {'\t', '\t', '\t', '\t', '\t', '\t'},
		7:  {'\t', '\t', '\t', '\t', '\t', '\t', '\t'},
		8:  {'\t', '\t', '\t', '\t', '\t', '\t', '\t', '\t'},
		9:  {'\t', '\t', '\t', '\t', '\t', '\t', '\t', '\t', '\t'},
		10: {'\t', '\t', '\t', '\t', '\t', '\t', '\t', '\t', '\t', '\t'},
	}
	maxT = len(tabs) - 1
)

func newWriter(w io.Writer, formatted bool, useCRLF bool) *writer {
	if !formatted {
		return &writer{
			w:       w,
			useCRLF: useCRLF,
		}
	}
	buf := &bytes.Buffer{}
	return &writer{
		w:             buf,
		formatted:     true,
		buf:           buf,
		origW:         w,
		useCRLF:       false,
		formattedCRLF: useCRLF,
	}
}

type writer struct {
	w             io.Writer
	formatted     bool
	buf           *bytes.Buffer
	origW         io.Writer
	useCRLF       bool
	formattedCRLF bool
	err           error
}

func (w *writer) format() error {
	if w.err == nil && w.formatted {
		var out []byte
		if out, w.err = format.Source(w.buf.Bytes()); w.err == nil {
			if w.formattedCRLF {
				out = bytes.ReplaceAll(out, lf, crlf)
			}
			_, w.err = w.origW.Write(out)
		} else {
			w.err = fmt.Errorf("formatting failed: %w", w.err)
		}
	}
	return w.err
}

func (w *writer) writeLine(indent int, line string, extra bool) {
	if w.err == nil {
		if w.writeIndent(indent) {
			_, w.err = w.w.Write([]byte(line))
			w.writeLf(extra)
		}
	}
}

func (w *writer) writeLf(extra bool) {
	if w.err == nil {
		if w.useCRLF && !w.formatted {
			if _, w.err = w.w.Write(crlf); w.err == nil && extra {
				_, w.err = w.w.Write(crlf)
			}
		} else if _, w.err = w.w.Write(lf); w.err == nil && extra {
			_, w.err = w.w.Write(lf)
		}
	}
}

func (w *writer) writeIndent(indent int) bool {
	if indent > 0 {
		for indent >= maxT && w.err == nil {
			_, w.err = w.w.Write(tabs[maxT])
			indent -= maxT
		}
		if w.err == nil && indent > 0 {
			_, w.err = w.w.Write(tabs[indent])
		}
	}
	return w.err == nil
}
