package main

import (
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"github.com/go-andiamo/flagpole"
	"io"
	"os"
)

const (
	subCmdStubs     = "stubs"
	subCmdStubsDesc = `Generate handler func stubs  (e.g. "func GetRoot(w http.ResponseWriter, r *http.Request) {...}") from existing OAS yaml/json`
)

type genStubsFlags struct {
	CommonFlags
	OutDir      *string `name:"outdir"       alias:"od" usage:"output directory for generated code"                           default:""            example:"[-outdir <dir>]"`
	OutFn       *string `name:"outf"         alias:"of" usage:"output filename for generated code (default: \"handlers.go\")" default:"handlers.go" example:"[-outf <filename>]"`
	Pkg         *string `name:"pkg"          alias:"pk" usage:"package for generated code (default: \"api\")"                 default:"api"         example:"[-pkg <name>]"`
	PublicFuncs *bool   `name:"public-funcs" alias:"pf" usage:"make handler funcs public (default: false)"                    default:"false"       example:"[-public-funcs]"`
	PathParams  *bool   `name:"path-params"  alias:"pp" usage:"include path params in handler funcs (default: false)"         default:"false"       example:"[-path-params]"`
	Receiver    *string `name:"receiver"     alias:"r"  usage:"make handler funcs with receiver - e.g. \"(a *MyApi)\" (default: no receiver)" default:"" example:"[-receiver <receiver-prefix>]"`
	Naming      *int    `name:"naming"       alias:"n"  usage:"handler func naming strategy (default: 0)\n        0: default naming strategy\n        1: try to use OAS operationId\n        2: try to use existing handler or x-handler" default:"0" example:"[-naming <0|1|2>]"`
	Path        *string `name:"path"                    usage:"api path to generate stubs for (optional) - e.g. \"/api/pets\""                      example:"[-path </api/foo>]"`
	GoDoc       *bool   `name:"godoc"        alias:"gd" usage:"include godoc comment for each handler (default: false)"       default:"false"       example:"[-godoc]"`
	CommonSupplementaryFlags
}

var genStubsFlagsParser = flagpole.MustNewParser[genStubsFlags](flagpole.StopOnHelp(true), flagpole.DefaultedOptionals(true), flagpole.IgnoreUnknownFlags(true))

func generateStubs(args []string) {
	flags, err := genStubsFlagsParser.Parse(args)
	if err != nil || (flags.Help != nil && *flags.Help) {
		out := os.Stdout
		code := 0
		if err != nil {
			out = os.Stderr
			code = 2
		}
		genStubsFlagsParser.Usage(out, err, cmdGen, subCmdStubs)
		os.Exit(code)
	}

	def, err := readDefinition(flags.In)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}
	stubNamer := &stubNaming{strategy: *flags.Naming}
	options := codegen.HandlerStubOptions{
		Package:     *flags.Pkg,
		PublicFuncs: *flags.PublicFuncs,
		Receiver:    *flags.Receiver,
		PathParams:  *flags.PathParams,
		StubNaming:  stubNamer,
		GoDoc:       *flags.GoDoc,
		Format:      !*flags.NoFormat,
	}
	if flags.Path != nil {
		pathDef := getPath(*flags.Path, def)
		if pathDef == nil {
			fail(1, fmt.Errorf("unknown path: %q", *flags.Path))
		}
		err = generatePathStubs(pathDef, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite)
	} else {
		err = generateDefinitionStubs(def, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite)
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
