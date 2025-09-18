package main

import (
	"flag"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"io"
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
	path := fs.String("path", "", `api path to generate stubs for (optional) - e.g. "/api/pets"`)
	godoc := fs.Bool("godoc", false, `include godoc comment for each handler func (optional, default: false)`)
	noFmt := fs.Bool("no-fmt", false, `suppress go formatting of generated code (optional, default: false)`)
	overwrite := fs.Bool("overwrite", false, `allow overwriting existing file (optional, default: false)`)
	help := fs.Bool("help", false, `show help`)
	egs := []string{
		"-in <filename>",
		"-outdir <dir>",
		"[-outf <filename>]",
		"[-pkg <name>]",
		"[-public-funcs]",
		"[-path-params]",
		"[-receiver <receiver-prefix>]",
		"[-naming <0|1|2>]",
		"[-path </api/foo>]",
		"[-godoc]",
		"[-no-fmt]",
		"[-overwrite]",
	}
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fail(2, err)
	}
	if *help {
		usageDetailed("", fs, egs, cmdGen, subCmdStubs, subCmdStubsDesc)
	}
	if *in == "" {
		usageDetailed("missing -in", fs, egs, cmdGen, subCmdStubs, subCmdStubsDesc)
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
		Format:      !*noFmt,
	}

	if *path != "" {
		pathDef := getPath(*path, def)
		if pathDef == nil {
			fail(1, fmt.Errorf("unknown path: %q", *path))
		}
		err = generatePathStubs(pathDef, options, *outDir, *outFn, *overwrite)
	} else {
		err = generateDefinitionStubs(def, options, *outDir, *outFn, *overwrite)
	}
	if err != nil {
		fail(1, fmt.Errorf("generate stubs: %w", err))
	}
}

func generateDefinitionStubs(def *chioas.Definition, options codegen.HandlerStubOptions, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "handlers.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateHandlerStubs(def, f, options)
	}
	return err
}

func generatePathStubs(def *chioas.Path, options codegen.HandlerStubOptions, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "handlers.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateHandlerStubs(def, f, options)
	}
	return err
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
