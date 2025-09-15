package codegen

import (
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/splitter"
	"golang.org/x/exp/maps"
	"io"
	"sort"
	"strings"
)

// StubNaming is an interface that can be used for HandlerStubOptions.StubNaming
type StubNaming interface {
	Name(path string, method string, def chioas.Method) string
}

// HandlerStubOptions is the options for GenerateHandlerStubs
type HandlerStubOptions struct {
	Package      string // e.g. "api" (default "api")
	SkipPrologue bool   // don't write package & imports
	// StubNaming is an optional handler func naming (if nil, the default naming is used)
	//
	// If StubNaming is provided but returns an empty string, the default naming is used
	StubNaming  StubNaming
	PublicFuncs bool   // pubic handler funcs
	Receiver    string // sets the receiver for handler funcs (e.g. "(api *MyApi)" - empty string for no receiver)
	PathParams  bool   // if true, writes vars for path params
	GoDoc       bool   // if true, writes a godoc comment for each handler
	// Format if set, formats output in canonical gofmt style (and checks syntax)
	//
	// Note: using this option means the output will be buffered before writing to the final writer
	Format  bool
	UseCRLF bool // true to use \r\n as the line terminator
}

type stubNaming struct{}

func (d *stubNaming) Name(path string, method string, def chioas.Method) string {
	if parts, err := pathSplitter.Split(path); err == nil && len(parts) > 1 {
		if last := parts[len(parts)-1]; strings.HasPrefix(last, "{") && strings.HasSuffix(last, "}") {
			last = last[1 : len(last)-1]
			if cAt := strings.IndexByte(last, ':'); cAt != -1 {
				last = last[:cAt]
			}
			path = parts[len(parts)-2] + " " + last
		} else {
			path = last
		}
	}
	return strings.ToLower(method) + "" + toPascal(path)
}

var defaultStubNaming StubNaming = &stubNaming{}

type StubItemType interface {
	chioas.Definition | *chioas.Definition |
		chioas.Path | *chioas.Path |
		chioas.Method | *chioas.Method |
		chioas.Paths
}

// GenerateHandlerStubs writes Go source for the http handlers for the specified definition (a chioas.Definition,
// chioas.Path, chioas.Paths or chioas.Method) to w using the supplied Options.
//
// Typical use: unmarshal an existing OpenAPI (YAML/JSON) into a chioas.Definition value, then
// emit http handler stubs
//
// Errors:
//   - Returns the first write error encountered. It does not close w.
func GenerateHandlerStubs[T StubItemType](item T, w io.Writer, opts HandlerStubOptions) error {
	if opts.StubNaming == nil {
		opts.StubNaming = &stubNaming{}
	}
	sw := newStubsWriter(w, opts)
	sw.writePrologue()
	switch it := any(item).(type) {
	case chioas.Definition:
		generateDefinitionStubs(it, sw)
	case *chioas.Definition:
		generateDefinitionStubs(*it, sw)
	case chioas.Paths:
		generatePathsStubs("", it, sw)
	case chioas.Path:
		generatePathStubs("", it, sw)
	case *chioas.Path:
		generatePathStubs("", *it, sw)
	case chioas.Method:
		generateMethodStub("", "", it, sw)
	case *chioas.Method:
		generateMethodStub("", "", *it, sw)
	}
	return sw.format()
}

func generateDefinitionStubs(def chioas.Definition, sw *stubsWriter) {
	generateMethodsStub("Root", def.Methods, sw)
	generatePathsStubs("", def.Paths, sw)
}

func generatePathStubs(path string, def chioas.Path, sw *stubsWriter) {
	generateMethodsStub(path, def.Methods, sw)
	generatePathsStubs(path, def.Paths, sw)
}

func generatePathsStubs(path string, def chioas.Paths, sw *stubsWriter) {
	ks := maps.Keys(def)
	sort.Strings(ks)
	for _, k := range ks {
		generatePathStubs(path+k, def[k], sw)
	}
}

func generateMethodsStub(path string, def chioas.Methods, sw *stubsWriter) {
	sms := maps.Keys(def)
	sort.Slice(sms, func(i, j int) bool {
		return compareMethods(sms[i], sms[j])
	})
	for _, m := range sms {
		generateMethodStub(path, m, def[m], sw)
	}
}

var pathSplitter = splitter.MustCreateSplitter('/', splitter.CurlyBrackets).
	AddDefaultOptions(splitter.IgnoreEmptyFirst, splitter.IgnoreEmptyLast)

func generateMethodStub(path string, method string, def chioas.Method, sw *stubsWriter) {
	var name string
	if path == "" && method == "" {
		name = "handler"
	} else {
		name = sw.opts.StubNaming.Name(path, method, def)
		if name == "" {
			name = defaultStubNaming.Name(path, method, def)
		}
	}
	name = toPascal(name)
	sw.writeFuncStart(name, method, path)
	if sw.opts.PathParams {
		if parts, err := pathSplitter.Split(path); err == nil {
			vNames := make([]string, 0)
			uScores := make([]string, 0)
			for _, part := range parts {
				if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
					name := part[1 : len(part)-1]
					if cAt := strings.Index(name, ":"); cAt != -1 {
						name = name[:cAt]
					}
					vNames = append(vNames, name)
					uScores = append(uScores, "_")
					sw.writeLine(1, name+` := chi.URLParam(r, "`+name+`")`, false)
				}
			}
			if len(vNames) > 0 {
				sw.writeLine(1, strings.Join(uScores, ", ")+" = "+strings.Join(vNames, ", "), false)
			}
		}
	}
	sw.writeLine(1, `// TODO implement me`, false)
	sw.writeLine(1, `panic("implement me!")`, false)
	sw.writeLine(0, "}", true)
}

func newStubsWriter(w io.Writer, opts HandlerStubOptions) *stubsWriter {
	return &stubsWriter{
		writer:  newWriter(w, opts.Format, opts.UseCRLF),
		opts:    opts,
		deduper: newNameDeDuper(),
	}
}

type stubsWriter struct {
	*writer
	opts    HandlerStubOptions
	deduper *nameDeDuper
}

func (w *stubsWriter) writePrologue() {
	const chiPkg = `"github.com/go-chi/chi/v5"`
	if w.err == nil && !w.opts.SkipPrologue {
		pkg := w.opts.Package
		if pkg == "" {
			pkg = defaultPackage
		}
		w.writeLine(0, "package "+pkg, true)
		w.writeLine(0, "import (", false)
		w.writeLine(1, `"net/http"`, false)
		if w.opts.PathParams {
			w.writeLf(false)
			w.writeLine(1, chiPkg, false)
		}
		w.writeLine(0, ")", true)
	}
}

var (
	handlerSignature = []byte("(w http.ResponseWriter, r *http.Request) {")
	handlerFunc      = []byte("func ")
)

func (w *stubsWriter) writeFuncGoDoc(name, method, path string) bool {
	if w.err == nil && w.opts.GoDoc {
		_, w.err = w.w.Write([]byte("// " + name + " " + method + " " + path))
		w.writeLf(false)
	}
	return w.err == nil
}

func (w *stubsWriter) writeFuncStart(name, method, path string) {
	if w.err == nil {
		if w.opts.PublicFuncs {
			name = strings.ToUpper(name[:1]) + name[1:]
		} else {
			name = strings.ToLower(name[:1]) + name[1:]
		}
		name = w.deduper.take(name)
		if w.writeFuncGoDoc(name, method, path) {
			if _, w.err = w.w.Write(handlerFunc); w.err == nil {
				if w.opts.Receiver != "" {
					_, w.err = w.w.Write([]byte(w.opts.Receiver + " "))
				}
				if w.err == nil {
					if _, w.err = w.w.Write([]byte(name)); w.err == nil {
						if _, w.err = w.w.Write(handlerSignature); w.err == nil {
							w.writeLf(false)
						}
					}
				}
			}
		}
	}
}
