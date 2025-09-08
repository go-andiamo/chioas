package codegen

import (
	"bytes"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"go/format"
	"golang.org/x/exp/maps"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var (
	lf   = []byte{'\n'}
	crlf = []byte{'\r', '\n'}
)

func newCodeWriter(w io.Writer, opts Options) *codeWriter {
	if !opts.Format {
		return &codeWriter{
			opts:    opts,
			w:       w,
			useCRLF: opts.UseCRLF,
		}
	}
	buf := &bytes.Buffer{}
	return &codeWriter{
		opts:      opts,
		w:         buf,
		formatted: true,
		buf:       buf,
		origW:     w,
		useCRLF:   false,
	}
}

type codeWriter struct {
	opts      Options
	w         io.Writer
	formatted bool
	buf       *bytes.Buffer
	origW     io.Writer
	useCRLF   bool
	err       error
}

func (w *codeWriter) format() error {
	if w.err == nil && w.formatted {
		var out []byte
		if out, w.err = format.Source(w.buf.Bytes()); w.err == nil {
			if w.opts.UseCRLF {
				out = bytes.ReplaceAll(out, lf, crlf)
			}
			_, w.err = w.origW.Write(out)
		}
	}
	return w.err
}

func (w *codeWriter) writePrologue() {
	if w.err == nil && !w.opts.SkipPrologue {
		pkg := w.opts.Package
		if pkg == "" {
			pkg = defaultPackage
		}
		w.writeLine(0, "package "+pkg, true)
		w.writeLine(0, "import (", false)
		if w.opts.UseHttpConsts {
			w.writeLine(0, "\t\"net/http\"", true)
		}
		if w.opts.ImportAlias != "" {
			w.writeLine(1, w.opts.ImportAlias+" "+chioasPkg, false)
		} else {
			w.writeLine(1, chioasPkg, false)
		}
		w.writeLine(0, ")", true)
	}
}

func (w *codeWriter) writeVarStart(name string, vtype string, ptr bool) {
	amp := ""
	if ptr {
		amp = "&"
	}
	w.writeLine(0, "var "+name+" = "+amp+w.opts.alias()+vtype+"{", false)
}

func (w *codeWriter) writeCollectionFieldStart(indent int, name string, vtype string) {
	w.writeLine(indent, name+": "+w.opts.alias()+vtype+"{", false)
}

func (w *codeWriter) writeLine(indent int, line string, extra bool) {
	if w.err == nil {
		if w.writeIndent(indent) {
			_, w.err = w.w.Write([]byte(line))
			w.writeLf(extra)
		}
	}
}

func (w *codeWriter) writeKey(indent int, k string) {
	w.writeLine(indent, strconv.Quote(k)+": {", false)
}

func (w *codeWriter) writeSchemaRef(indent int, ref string) {
	w.writeLine(indent, "SchemaRef: "+strconv.Quote(refs.Normalize(tags.Schemas, ref))+",", false)
}

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

func (w *codeWriter) writeIndent(indent int) bool {
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

func (w *codeWriter) writeLf(extra bool) {
	if w.err == nil {
		if w.useCRLF {
			if _, w.err = w.w.Write(crlf); w.err == nil && extra {
				_, w.err = w.w.Write(crlf)
			}
		} else if _, w.err = w.w.Write(lf); w.err == nil && extra {
			_, w.err = w.w.Write(lf)
		}
	}
}

func (w *codeWriter) writeStart(indent int, s string) {
	if w.err == nil {
		if w.writeIndent(indent) {
			_, w.err = w.w.Write([]byte(s))
		}
	}
}

func (w *codeWriter) writeEnd(indent int, s string) {
	if w.err == nil {
		if w.writeIndent(indent) {
			if _, w.err = w.w.Write([]byte(s)); w.err == nil {
				w.writeLf(false)
			}
		}
	}
}

func (w *codeWriter) writeValue(indent int, value any) {
	if w.err == nil {
		switch v := value.(type) {
		case nil:
			if _, w.err = w.w.Write([]byte("nil,")); w.err == nil {
				w.writeLf(false)
			}
		case string:
			if _, w.err = w.w.Write([]byte(strconv.Quote(v) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case bool:
			if _, w.err = w.w.Write([]byte(strconv.FormatBool(v) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case int:
			if _, w.err = w.w.Write([]byte(strconv.FormatInt(int64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case int8:
			if _, w.err = w.w.Write([]byte(strconv.FormatInt(int64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case int16:
			if _, w.err = w.w.Write([]byte(strconv.FormatInt(int64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case int32:
			if _, w.err = w.w.Write([]byte(strconv.FormatInt(int64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case int64:
			if _, w.err = w.w.Write([]byte(strconv.FormatInt(v, 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case uint:
			if _, w.err = w.w.Write([]byte(strconv.FormatUint(uint64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case uint8:
			if _, w.err = w.w.Write([]byte(strconv.FormatUint(uint64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case uint16:
			if _, w.err = w.w.Write([]byte(strconv.FormatUint(uint64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case uint32:
			if _, w.err = w.w.Write([]byte(strconv.FormatUint(uint64(v), 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case uint64:
			if _, w.err = w.w.Write([]byte(strconv.FormatUint(v, 10) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case float32:
			if _, w.err = w.w.Write([]byte(strconv.FormatFloat(float64(v), 'f', -1, 32) + ",")); w.err == nil {
				w.writeLf(false)
			}
		case float64:
			if _, w.err = w.w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64) + ",")); w.err == nil {
				w.writeLf(false)
			}
		default:
			w.writeExtendedValue(indent, value)
		}
	}
}

func (w *codeWriter) writeValueOnly(indent int, value any) {
	if w.err == nil {
		switch v := value.(type) {
		case nil:
			if w.writeIndent(indent) {
				if _, w.err = w.w.Write([]byte("nil,")); w.err == nil {
					w.writeLf(false)
				}
			}
		case string:
			if w.writeIndent(indent) {
				if _, w.err = w.w.Write([]byte(strconv.Quote(v) + ",")); w.err == nil {
					w.writeLf(false)
				}
			}
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			if w.writeIndent(indent) {
				if _, w.err = w.w.Write([]byte(fmt.Sprintf("%v,", value))); w.err == nil {
					w.writeLf(false)
				}
			}
		default:
			if w.writeIndent(indent) {
				w.writeExtendedValue(indent, value)
			}
		}
	}
}

func (w *codeWriter) writeExtendedValue(indent int, value any) {
	if w.err == nil {
		ok := false
		vo := reflect.ValueOf(value)
		switch vo.Kind() {
		case reflect.Map:
			if vo.Len() == 0 {
				ok = true
				vts := strings.ReplaceAll(fmt.Sprintf("%T", value), "interface {}", "any") + "{},"
				if _, w.err = w.w.Write([]byte(vts)); w.err == nil {
					w.writeLf(false)
				}
			} else {
				vt := reflect.TypeOf(value)
				kk := vt.Key().Kind()
				keyIsStr := false
				switch kk {
				case reflect.String:
					keyIsStr = true
					ok = true
				case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
					ok = true
				}
				if ok {
					vts := strings.ReplaceAll(fmt.Sprintf("%T", value), "interface {}", "any") + "{"
					w.writeStart(0, vts)
					w.writeLf(false)
					iter := vo.MapRange()
					for iter.Next() {
						if keyIsStr {
							w.writeStart(indent+1, fmt.Sprintf("%q: ", iter.Key().Interface()))
						} else {
							w.writeStart(indent+1, fmt.Sprintf("%v: ", iter.Key().Interface()))
						}
						w.writeValue(indent+1, iter.Value().Interface())
					}
					w.writeEnd(indent, "},")
				}
			}
		case reflect.Slice:
			ok = true
			if vo.Len() == 0 {
				vts := strings.ReplaceAll(fmt.Sprintf("%T", value), "interface {}", "any") + "{},"
				if _, w.err = w.w.Write([]byte(vts)); w.err == nil {
					w.writeLf(false)
				}
			} else {
				vts := strings.ReplaceAll(fmt.Sprintf("%T", value), "interface {}", "any") + "{"
				w.writeStart(0, vts)
				w.writeLf(false)
				for i := 0; i < vo.Len(); i++ {
					v := vo.Index(i).Interface()
					w.writeValueOnly(indent+1, v)
				}
				w.writeEnd(indent, "},")
			}
		}
		if !ok {
			if _, w.err = w.w.Write([]byte(fmt.Sprintf(`"Unknown value type: %T",`, value))); w.err == nil {
				w.writeLf(false)
			}
		}
	}
}

func (w *codeWriter) writeExtensions(indent int, extensions chioas.Extensions) {
	if w.err == nil && len(extensions) > 0 {
		w.writeLine(indent, typeExtensions+": "+w.opts.alias()+typeExtensions+"{", false)
		ks := maps.Keys(extensions)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeStart(indent+1, strconv.Quote(k)+": ")
			w.writeValue(indent+1, extensions[k])
		}
		w.writeEnd(indent, "},")
	}
}

func isZeroValue(value any) bool {
	switch value.(type) {
	case nil:
		return true
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

func hasNonZeroValues(values ...any) bool {
	for _, value := range values {
		if !isZeroValue(value) {
			return true
		}
	}
	return false
}

func writeZeroField[T any](w *codeWriter, indent int, name string, value T) {
	if w.err == nil {
		ok := true
		if w.opts.OmitZeroValues {
			ok = !isZeroValue(value)
		}
		if ok {
			switch vt := any(value).(type) {
			case nil:
				w.writeLine(indent, name+": nil,", false)
			case string:
				w.writeLine(indent, name+": "+strconv.Quote(vt)+",", false)
			default:
				w.writeLine(indent, fmt.Sprintf("%s: %v,", name, value), false)
			}
		}
	}
}

type nameDeDuper struct {
	used map[string]struct{}
	next map[string]int
}

func newNameDeDuper() *nameDeDuper {
	return &nameDeDuper{
		used: make(map[string]struct{}),
		next: make(map[string]int),
	}
}

func (d *nameDeDuper) clear() {
	d.used = make(map[string]struct{})
	d.next = make(map[string]int)
}

func (d *nameDeDuper) take(name string) string {
	i := d.next[name]
	for {
		candidate := name
		if i > 0 {
			if r := []rune(name); len(r) > 0 && unicode.IsDigit(r[len(r)-1]) {
				candidate = name + "_" + strconv.Itoa(i+1)
			} else {
				candidate = name + strconv.Itoa(i+1)
			}
		}
		if _, ok := d.used[candidate]; !ok {
			d.used[candidate] = struct{}{}
			d.next[name] = i + 1
			return candidate
		}
		i++
	}
}
