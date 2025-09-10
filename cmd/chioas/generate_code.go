package main

import (
	"flag"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"github.com/go-andiamo/splitter"
	"io"
	"os"
	"strings"
)

const (
	subCmdCode     = "code"
	subCmdCodeDesc = `Generate chioas definition code (e.g. "var definition = chioas.Definition{...}") from existing OAS yaml/json`
)

func generateCode(args []string) {
	fs := flag.NewFlagSet("gen code", flag.ContinueOnError)
	in := fs.String("in", "", `input definition file (.yaml|.json) or '-' for stdin (required)`)
	outDir := fs.String("outdir", "", `output directory for generated code (optional, defaults to current dir)`)
	outFn := fs.String("outf", "", `output filename for generated code (optional, default: definition.go)`)
	pkg := fs.String("pkg", "", `Go package name for generated code (optional, default: api)`)
	varName := fs.String("var", "", `name for the top-level variable (optional, default: definition or Definition with -public-vars)`)
	importAlias := fs.String("import-alias", "", `import alias for chioas (optional)`)
	omitZero := fs.Bool("omit-zero", false, `omit zero-valued fields (optional, default: false)`)
	hoistPaths := fs.Bool("hoist-paths", false, `hoist top-level vars for paths (optional, default: false)`)
	hoistComponents := fs.Bool("hoist-components", false, `hoist top-level vars for named components (optional, default: false)`)
	publicVars := fs.Bool("public-vars", false, `export top-level vars (optional, default: false)`)
	useHTTPConsts := fs.Bool("http-consts", false, `use http.MethodGet, http.Status (optional, default: false)`)
	path := fs.String("path", "", `api path to generate code for (optional) - e.g. "/api/pets"`)
	overwrite := fs.Bool("overwrite", false, `allow overwriting existing file (optional, default: false)`)
	help := fs.Bool("help", false, `show help`)
	egs := []string{
		"-in <spec.{yaml|json}>",
		"-outdir <dir>",
		"[-outf <filename>]",
		"[-pkg <name>]",
		"[-var <name>]",
		"[-import-alias <name>|.]",
		"[-omit-zero]",
		"[-hoist-paths]",
		"[-hoist-components]",
		"[-public-vars]",
		"[-http-consts]",
		"[-path </api/foo>]",
		"[-overwrite]",
	}
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fail(2, err)
	}
	if *help {
		usageGenCode("", fs, egs)
	}
	if *in == "" {
		usageGenCode("missing -in", fs, egs)
	}
	def, err := readDefinition(*in)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}
	options := codegen.Options{
		Package:         *pkg,
		VarName:         *varName,
		ImportAlias:     *importAlias,
		OmitZeroValues:  *omitZero,
		HoistPaths:      *hoistPaths,
		HoistComponents: *hoistComponents,
		PublicVars:      *publicVars,
		UseHttpConsts:   *useHTTPConsts,
		Format:          true,
	}
	if *path != "" {
		s, _ := splitter.NewSplitter('/', splitter.CurlyBrackets)
		parts, err := s.Split(*path, splitter.IgnoreEmptyFirst, splitter.IgnoreEmptyLast)
		if err != nil {
			fail(1, fmt.Errorf("invalid path: %w", err))
		}
		var pathDef *chioas.Path
		paths := def.Paths
		for i, p := range parts {
			if i == len(parts)-1 {
				if pv, ok := paths["/"+p]; ok {
					pathDef = &pv
				}
			} else if sub, ok := paths["/"+p]; ok {
				paths = sub.Paths
			} else {
				break
			}
		}
		if pathDef == nil {
			fail(1, fmt.Errorf("unknown path: %q", *path))
		}
		if err = generatePathCode(pathDef, options, *outDir, *outFn, *overwrite); err != nil {
			fail(1, fmt.Errorf("generate code: %w", err))
		}
	} else if err = generateDefinitionCode(def, options, *outDir, *outFn, *overwrite); err != nil {
		fail(1, fmt.Errorf("generate code: %w", err))
	}
}

func generateDefinitionCode(def *chioas.Definition, options codegen.Options, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "definition.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateCode(def, f, options)
	}
	return err
}

func generatePathCode(def *chioas.Path, options codegen.Options, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "definition.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateCode(def, f, options)
	}
	return err
}

func usageGenCode(msg string, fs *flag.FlagSet, egs []string) {
	out := os.Stdout
	if msg != "" {
		out = os.Stderr
		_, _ = fmt.Fprintf(out, "error: %s\n\n", msg)
	}
	_, _ = fmt.Fprintln(out, "Description: "+subCmdCodeDesc)
	_, _ = fmt.Fprintf(out, "Usage:\n  %s %s %s %s\n\n", cmdChioas, cmdGen, subCmdCode, strings.Join(egs, " "))
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
