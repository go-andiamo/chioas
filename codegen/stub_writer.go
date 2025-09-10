package codegen

import (
	"bytes"
	"go/format"
	"io"
	"strings"
)

func newStubsWriter(w io.Writer, opts HandlerStubOptions) *stubsWriter {
	if !opts.Format {
		return &stubsWriter{
			opts:    opts,
			w:       w,
			useCRLF: opts.UseCRLF,
			deduper: newNameDeDuper(),
		}
	}
	buf := &bytes.Buffer{}
	return &stubsWriter{
		opts:      opts,
		w:         buf,
		formatted: true,
		buf:       buf,
		origW:     w,
		useCRLF:   false,
		deduper:   newNameDeDuper(),
	}
}

type stubsWriter struct {
	opts      HandlerStubOptions
	w         io.Writer
	formatted bool
	buf       *bytes.Buffer
	origW     io.Writer
	useCRLF   bool
	err       error
	deduper   *nameDeDuper
}

func (w *stubsWriter) format() error {
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

func (w *stubsWriter) writeLine(indent int, line string, extra bool) {
	if w.err == nil {
		if w.writeIndent(indent) {
			_, w.err = w.w.Write([]byte(line))
			w.writeLf(extra)
		}
	}
}

func (w *stubsWriter) writeIndent(indent int) bool {
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

func (w *stubsWriter) writeLf(extra bool) {
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
