package main

import (
	"flag"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"io"
	"os"
	"strings"
)

const (
	subCmdStubs     = "stubs"
	subCmdStubsDesc = `Generate handler func stubs  (e.g. "func GetRoot(w http.ResponseWriter, r *http.Request) {...}") from existing OAS yaml/json`
)

func generateStubs(args []string) {
	fs := flag.NewFlagSet("gen stubs", flag.ContinueOnError)
	in := fs.String("in", "", `input definition file (.yaml|.json) or '-' for stdin (required)`)
	outDir := fs.String("outdir", "", `output directory for generated code (optional, defaults to current dir)`)
	outFn := fs.String("outf", "", `output filename for generated code (optional, default: handlers.go)`)
	pkg := fs.String("pkg", "", `Go package name for generated code (optional, default: api)`)
	publicFuncs := fs.Bool("public-funcs", false, `make handler funcs public (optional, default: false)`)
	pathParams := fs.Bool("path-params", false, `include path params in handler funcs (optional, default: false)`)
	receiver := fs.String("receiver", "", `make handler funcs with receiver - e.g. "(a *MyApi)" (optional, default: no receiver)`)
	naming := fs.Int("naming", 0, `handler func naming strategy (optional, default: 0)
        0: default naming strategy
		1: try to use OAS operationId
        2: try to use existing handler or x-handler`)
	godoc := fs.Bool("godoc", false, `include godoc comment for each handler func (optional, default: false)`)
	overwrite := fs.Bool("overwrite", false, `allow overwriting existing file (optional, default: false)`)
	help := fs.Bool("help", false, `show help`)
	egs := []string{
		"-in <spec.{yaml|json}>",
		"-outdir <dir>",
		"[-outf <filename>]",
		"[-pkg <name>]",
		"[-public-funcs]",
		"[-path-params]",
		"[-receiver <receiver-prefix>]",
		"[-naming <0|1|2>]",
		"[-godoc]",
		"[-overwrite]",
	}
	_ = naming
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fail(2, err)
	}
	if *help {
		usageGenStubs("", fs, egs)
	}
	if *in == "" {
		usageGenStubs("missing -in", fs, egs)
	}
	def, err := readDefinition(*in)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}
	stubNamer := &stubNaming{strategy: *naming}
	options := codegen.HandlerStubOptions{
		Package:     *pkg,
		PublicFuncs: *publicFuncs,
		Receiver:    *receiver,
		PathParams:  *pathParams,
		StubNaming:  stubNamer,
		GoDoc:       *godoc,
		Format:      true,
	}

	var f io.WriteCloser
	if f, err = createFile(*outFn, *outDir, *overwrite, "handlers.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateHandlerStubs(def, f, options)
	}
	if err != nil {
		fail(1, err)
	}
}

type stubNaming struct {
	strategy int
}

func (s *stubNaming) Name(path string, method string, def chioas.Method) (result string) {
	if s.strategy > 1 {
		if h, ok := def.Handler.(string); ok {
			result = h
		} else if xh, ok := def.Extensions["handler"]; ok {
			if h, ok = xh.(string); ok {
				result = h
			}
		}
	}
	if result == "" && s.strategy > 0 {
		result = def.OperationId
	}
	return result
}

func usageGenStubs(msg string, fs *flag.FlagSet, egs []string) {
	out := os.Stdout
	if msg != "" {
		out = os.Stderr
		_, _ = fmt.Fprintf(out, "error: %s\n\n", msg)
	}
	_, _ = fmt.Fprintln(out, "Description: "+subCmdStubsDesc)
	_, _ = fmt.Fprintf(out, "Usage:\n  %s %s %s %s\n\n", cmdChioas, cmdGen, subCmdStubs, strings.Join(egs, " "))
	_, _ = fmt.Fprintln(out, `Flags:`)
	fs.VisitAll(func(f *flag.Flag) {
		_, _ = fmt.Fprintf(out, "    -%s\n", f.Name)
		_, _ = fmt.Fprintf(out, "        %s\n", f.Usage)
	})
	if msg != "" {
		os.Exit(2)
	} else {
		os.Exit(0)
	}
}
