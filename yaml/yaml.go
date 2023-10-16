package yaml

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v3"
	"reflect"
	"strconv"
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
	WriteItemValue(name string, value any) Writer
	WriteItemStart(name string, value any) Writer
	// WriteLines writes the provided lines (current indent is added to each line)
	WriteLines(lines ...string) Writer
	// Write writes the provided data (no indent is added - and data must include indents!)
	Write(data []byte) Writer
	WriteComments(lines ...string) Writer
	CurrentIndent() string
	SetError(err error)
	Errored() error
	RefChecker(rc RefChecker) RefChecker
}

// RefChecker is an interface optionally used by Writer.RefChecker so that refs can be checked for existence
type RefChecker interface {
	RefCheck(area, ref string) error
}

var _ Writer = &writer{}

type writer struct {
	buffer     bytes.Buffer
	w          *bufio.Writer
	indent     []byte
	err        error
	refChecker RefChecker
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
	} else if y.err == nil {
		y.err = errors.New("attempt to end un-started indent")
	}
}

func (y *writer) writeIndent() bool {
	if y.err == nil {
		_, y.err = y.w.Write(y.indent)
	}
	return y.err == nil
}

type LiteralValue struct {
	Value string
}

func (y *writer) yamlValue(value any, allowEmpty bool) []string {
	result := make([]string, 0)
	switch vt := value.(type) {
	case LiteralValue:
		result = append(result, vt.Value)
	case *LiteralValue:
		result = append(result, vt.Value)
	case json.Number:
		result = append(result, vt.String())
	case string:
		if vt != "" || allowEmpty {
			if strings.Contains(vt, "\n") {
				result = append(result, y.formattedString(vt)...)
			} else {
				result = append(result, `"`+strings.ReplaceAll(vt, `"`, `\"`)+`"`)
			}
		}
	case *string:
		vo := reflect.ValueOf(vt)
		if !vo.IsZero() {
			result = y.yamlValue(vo.Elem().Interface(), allowEmpty)
		} else if allowEmpty {
			result = append(result, `""`)
		}
	case bool:
		result = append(result, fmt.Sprintf("%t", vt))
	case *bool:
		if vt == nil && allowEmpty {
			result = append(result, "false")
		} else if vt != nil {
			result = append(result, fmt.Sprintf("%t", *vt))
		}
	case int, int8, int16, int32, int64:
		result = append(result, fmt.Sprintf("%d", vt))
	case uint, uint8, uint16, uint32, uint64:
		result = append(result, fmt.Sprintf("%d", vt))
	case float32:
		result = append(result, strconv.FormatFloat(float64(vt), 'f', -1, 32))
	case float64:
		result = append(result, strconv.FormatFloat(vt, 'f', -1, 64))
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
		vo := reflect.ValueOf(vt)
		if !vo.IsZero() {
			result = y.yamlValue(vo.Elem().Interface(), allowEmpty)
		}
	default:
		if value != nil {
			result = append(result, y.marshalYaml(value)...)
		}
	}
	return result
}

func (y *writer) marshalYaml(v any) []string {
	result := []string{""}
	var buffer bytes.Buffer
	enc := goyaml.NewEncoder(&buffer)
	enc.SetIndent(2)
	if y.err = enc.Encode(v); y.err == nil {
		lines := strings.Split(buffer.String(), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		result = append(result, lines...)
	}
	return result
}

func (y *writer) formattedString(s string) []string {
	lines := strings.Split(s, "\n")
	result := make([]string, len(lines)+1)
	result[0] = `>-`
	for i, line := range lines {
		result[i+1] = line
	}
	return result
}

func (y *writer) WriteTagValue(name string, value any) Writer {
	if y.err == nil && value != nil {
		wv := y.yamlValue(value, false)
		if len(wv) > 0 && y.writeIndent() {
			_, y.err = y.w.WriteString(name + ":" + padFirst(wv[0]) + "\n")
			for i := 1; i < len(wv); i++ {
				_, y.err = y.w.WriteString(string(y.indent) + "  " + wv[i] + "\n")
			}
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
		if len(wv) > 0 && y.writeIndent() {
			_, y.err = y.w.WriteString("-" + padFirst(wv[0]) + "\n")
			for i := 1; i < len(wv); i++ {
				_, y.err = y.w.WriteString(string(y.indent) + "  " + wv[i] + "\n")
			}
		}
	}
	return y
}

func (y *writer) WriteItemValue(name string, value any) Writer {
	if y.err == nil {
		if value == nil {
			if y.writeIndent() {
				_, y.err = y.w.WriteString("- " + name + ":\n")
			}
		} else {
			wv := y.yamlValue(value, true)
			if len(wv) > 0 && y.writeIndent() {
				_, y.err = y.w.WriteString("- " + name + ":" + padFirst(wv[0]) + "\n")
				for i := 1; i < len(wv); i++ {
					_, y.err = y.w.WriteString(string(y.indent) + "  " + wv[i] + "\n")
				}
			}
		}
	}
	return y
}

func (y *writer) WriteItemStart(name string, value any) Writer {
	if y.err == nil {
		if y.writeIndent() {
			wv := y.yamlValue(value, true)
			if len(wv) > 0 {
				_, y.err = y.w.WriteString("- " + name + ":" + padFirst(wv[0]) + "\n")
				for i := 1; i < len(wv); i++ {
					_, y.err = y.w.WriteString(string(y.indent) + "  " + wv[i] + "\n")
				}
			} else {
				_, y.err = y.w.WriteString(`- ` + name + ":\n")
			}
			y.incIndent()
		}
	}
	return y
}

func padFirst(first string) string {
	if first != "" {
		return " " + first
	}
	return ""
}

func (y *writer) WriteLines(lines ...string) Writer {
	for _, ln := range lines {
		if strings.Contains(ln, "\n") {
			y.WriteLines(strings.Split(ln, "\n")...)
		} else if y.writeIndent() {
			_, y.err = y.w.WriteString(ln + "\n")
		}
	}
	return y
}

func (y *writer) Write(data []byte) Writer {
	_, y.err = y.w.Write(data)
	return y
}

func (y *writer) WriteComments(lines ...string) Writer {
	if len(lines) > 1 || (len(lines) == 1 && lines[0] != "") {
		for _, ln := range lines {
			if strings.Contains(ln, "\n") {
				y.WriteComments(strings.Split(ln, "\n")...)
			} else if y.writeIndent() {
				_, y.err = y.w.WriteString("#" + ln + "\n")
			}
		}
	}
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

func (y *writer) RefChecker(rc RefChecker) RefChecker {
	if rc != nil {
		y.refChecker = rc
		return rc
	} else if y.refChecker != nil {
		return y.refChecker
	}
	return nullRefChecker
}

var nullRefChecker RefChecker = &refChecker{}

type refChecker struct {
}

func (r *refChecker) RefCheck(area, ref string) error {
	return nil
}
