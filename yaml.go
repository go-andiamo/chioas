package chioas

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type yamlWriter struct {
	buffer bytes.Buffer
	w      *bufio.Writer
	indent []byte
	err    error
}

func newYamlWriter(w *bufio.Writer) *yamlWriter {
	res := &yamlWriter{}
	if w != nil {
		res.w = w
	} else {
		res.w = bufio.NewWriter(&res.buffer)
	}
	return res
}

func (y *yamlWriter) bytes() ([]byte, error) {
	y.err = y.w.Flush()
	return y.buffer.Bytes(), y.err
}

func (y *yamlWriter) incIndent() {
	y.indent = append(y.indent, ' ', ' ')
}

func (y *yamlWriter) decIndent() {
	if len(y.indent) > 1 {
		y.indent = y.indent[2:]
	} else {
		y.err = errors.New("attempt to end un-started indent")
	}
}

func (y *yamlWriter) writeIndent() bool {
	if y.err == nil {
		_, y.err = y.w.Write(y.indent)
	}
	return y.err == nil
}

func yamlValue(value any, allowEmpty bool) string {
	result := ""
	switch vt := value.(type) {
	case string:
		if vt != "" || allowEmpty {
			result = `"` + strings.ReplaceAll(vt, `"`, `\"`) + `"`
		}
	default:
		result = fmt.Sprintf("%v", value)
	}
	return result
}

func (y *yamlWriter) writeTagValue(name string, value any) {
	if y.err == nil && value != nil {
		wv := yamlValue(value, false)
		if wv != "" && y.writeIndent() {
			_, y.err = y.w.WriteString(name + ": " + wv + "\n")
		}
	}
}

func (y *yamlWriter) writeTagStart(name string) {
	if y.writeIndent() {
		_, y.err = y.w.WriteString(name + ":\n")
		y.incIndent()
	}
}

func (y *yamlWriter) writeTagEnd() {
	y.decIndent()
}

func (y *yamlWriter) writePathStart(context string, path string) {
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
}

func (y *yamlWriter) writeItem(value any) {
	if y.err == nil && value != nil {
		wv := yamlValue(value, false)
		if wv != "" && y.writeIndent() {
			_, y.err = y.w.WriteString("- " + wv + "\n")
		}
	}
}

func (y *yamlWriter) writeItemStart(name string, value any) {
	if y.err == nil {
		if y.writeIndent() {
			_, y.err = y.w.WriteString(`- ` + name + `: ` + yamlValue(value, true) + "\n")
			y.incIndent()
		}
	}
}
