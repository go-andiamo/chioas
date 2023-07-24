package yaml

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Writer interface {
	Flush() error
	Bytes() ([]byte, error)
	WriteTagValue(name string, value any) Writer
	WriteTagStart(name string) Writer
	WriteTagEnd() Writer
	WritePathStart(context string, path string) Writer
	WriteItem(value any) Writer
	WriteItemStart(name string, value any) Writer
	// WriteLines writes the provided lines (current indent is added to each line)
	WriteLines(lines ...string) Writer
	// Write writes the provided data (no indent is added - and data must include indents!)
	Write(data []byte) Writer
	CurrentIndent() string
	SetError(err error)
	Errored() error
}

var _ Writer = &writer{}

type writer struct {
	buffer bytes.Buffer
	w      *bufio.Writer
	indent []byte
	err    error
}

func newWriter(w *bufio.Writer) *writer {
	res := &writer{}
	if w != nil {
		res.w = w
	} else {
		res.w = bufio.NewWriter(&res.buffer)
	}
	return res
}

func NewWriter(w *bufio.Writer) Writer {
	return newWriter(w)
}

func (y *writer) Flush() error {
	if y.err == nil {
		y.err = y.w.Flush()
	}
	return y.err
}

func (y *writer) Bytes() ([]byte, error) {
	if y.err != nil {
		return nil, y.err
	}
	y.err = y.w.Flush()
	return y.buffer.Bytes(), y.err
}

func (y *writer) incIndent() {
	y.indent = append(y.indent, ' ', ' ')
}

func (y *writer) decIndent() {
	if len(y.indent) > 1 {
		y.indent = y.indent[2:]
	} else {
		y.err = errors.New("attempt to end un-started indent")
	}
}

func (y *writer) writeIndent() bool {
	if y.err == nil {
		_, y.err = y.w.Write(y.indent)
	}
	return y.err == nil
}

func (y *writer) formattedString(s string) string {
	var builder strings.Builder
	lines := strings.Split(s, "\n")
	l := len(lines)
	indent := append(y.indent, ' ', ' ')
	builder.Grow(3 + len(s) + (l * (len(indent) + 1)))
	builder.WriteString(">-\n")
	l--
	for i, line := range lines {
		builder.Write(indent)
		builder.WriteString(line)
		if i < l {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

func (y *writer) yamlValue(value any, allowEmpty bool) string {
	result := ""
	switch vt := value.(type) {
	case string:
		if vt != "" || allowEmpty {
			if strings.Contains(vt, "\n") {
				result = y.formattedString(vt)
			} else {
				result = `"` + strings.ReplaceAll(vt, `"`, `\"`) + `"`
			}
		}
	case *string:
		if vt == nil && allowEmpty {
			result = `""`
		} else if vt != nil && (*vt != "" || allowEmpty) {
			result = `"` + strings.ReplaceAll(*vt, `"`, `\"`) + `"`
		}
	case bool:
		result = fmt.Sprintf("%t", vt)
	case *bool:
		if vt == nil && allowEmpty {
			result = "false"
		} else if vt != nil {
			result = fmt.Sprintf("%t", *vt)
		}
	case int, int8, int16, int32, int64:
		result = fmt.Sprintf("%d", vt)
	case uint, uint8, uint16, uint32, uint64:
		result = fmt.Sprintf("%d", vt)
	case float32, float64:
		result = fmt.Sprintf("%f", vt)
	case *int:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *int8:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *int16:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *int32:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *int64:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *uint:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *uint8:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *uint16:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *uint32:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *uint64:
		if vt != nil {
			result = fmt.Sprintf("%d", *vt)
		}
	case *float32:
		if vt != nil {
			result = fmt.Sprintf("%f", *vt)
		}
	case *float64:
		if vt != nil {
			result = fmt.Sprintf("%f", *vt)
		}
	}
	return result
}

func (y *writer) WriteTagValue(name string, value any) Writer {
	if y.err == nil && value != nil {
		wv := y.yamlValue(value, false)
		if wv != "" && y.writeIndent() {
			_, y.err = y.w.WriteString(name + ": " + wv + "\n")
		}
	}
	return y
}

func (y *writer) WriteTagStart(name string) Writer {
	if y.writeIndent() {
		_, y.err = y.w.WriteString(name + ":\n")
		y.incIndent()
	}
	return y
}

func (y *writer) WriteTagEnd() Writer {
	y.decIndent()
	return y
}

func (y *writer) WritePathStart(context string, path string) Writer {
	if y.err == nil {
		if y.writeIndent() {
			if context != "" {
				_, y.err = y.w.WriteString(`"/` + context + path + "\":\n")
			} else {
				_, y.err = y.w.WriteString(`"` + path + "\":\n")
			}
			y.incIndent()
		}
	}
	return y
}

func (y *writer) WriteItem(value any) Writer {
	if y.err == nil && value != nil {
		wv := y.yamlValue(value, false)
		if wv != "" && y.writeIndent() {
			_, y.err = y.w.WriteString("- " + wv + "\n")
		}
	}
	return y
}

func (y *writer) WriteItemStart(name string, value any) Writer {
	if y.err == nil {
		if y.writeIndent() {
			_, y.err = y.w.WriteString(`- ` + name + `: ` + y.yamlValue(value, true) + "\n")
			y.incIndent()
		}
	}
	return y
}

func (y *writer) WriteLines(lines ...string) Writer {
	for _, ln := range lines {
		if y.writeIndent() {
			_, y.err = y.w.WriteString(ln + "\n")
		}
	}
	return y
}

func (y *writer) Write(data []byte) Writer {
	_, y.err = y.w.Write(data)
	return y
}

func (y *writer) CurrentIndent() string {
	return string(y.indent)
}

func (y *writer) SetError(err error) {
	if y.err == nil {
		y.err = err
	}
}

func (y *writer) Errored() error {
	return y.err
}
