package main

import (
	"flag"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"github.com/go-andiamo/chioas/internal/refs"
	"io"
)

const (
	subCmdStructs     = "structs"
	subCmdStructsDesc = `Generate schema/request/response structs from existing OAS yaml/json`
)

func generateStructs(args []string) {
	fs := flag.NewFlagSet("gen structs", flag.ContinueOnError)
	in := fs.String("in", "", `input definition file (.yaml|.json) or '-' for stdin (required)`)
	outDir := fs.String("outdir", "", `output directory for generated code (optional, defaults to current dir)`)
	outFn := fs.String("outf", "", `output filename for generated code (optional, default: structs.go)`)
	pkg := fs.String("pkg", "", `Go package name for generated code (optional, default: api)`)
	publicStructs := fs.Bool("public-structs", false, `make structs public (optional, default: false)`)
	noRequests := fs.Bool("no-requests", false, `suppress schemas for requests (optional, default: false)`)
	noResponses := fs.Bool("no-responses", false, `suppress schemas for responses (optional, default: false)`)
	oasTags := fs.Bool("oas-tags", false, `oas tags for struct fields (optional, default: false)`)
	keep := fs.Bool("keep", false, `keep references to components as separate structs (optional, default: false)`)
	path := fs.String("path", "", `api path to generate structs for (optional) - e.g. "/api/pets"`)
	godoc := fs.Bool("godoc", false, `include godoc comment for each struct (optional, default: false)`)
	noFmt := fs.Bool("no-fmt", false, `suppress go formatting of generated code (optional, default: false)`)
	overwrite := fs.Bool("overwrite", false, `allow overwriting existing file (optional, default: false)`)
	help := fs.Bool("help", false, `show help`)
	egs := []string{
		"-in <filename>",
		"-outdir <dir>",
		"[-outf <filename>]",
		"[-pkg <name>]",
		"[-public-structs]",
		"[-no-requests]",
		"[-noResponses]",
		"[-oasTags]",
		"[-keep]",
		"[-godoc]",
		"[-no-fmt]",
		"[-overwrite]",
	}
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fail(2, err)
	}
	if *help {
		usageDetailed("", fs, egs, cmdGen, subCmdStructs, subCmdStructsDesc)
	}
	if *in == "" {
		usageDetailed("missing -in", fs, egs, cmdGen, subCmdStructs, subCmdStructsDesc)
	}
	def, err := readDefinition(*in)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}

	options := codegen.SchemaStructOptions{
		Package:                 *pkg,
		PublicStructs:           *publicStructs,
		NoRequests:              *noRequests,
		NoResponses:             *noResponses,
		OASTags:                 *oasTags,
		GoDoc:                   *godoc,
		KeepComponentProperties: *keep,
		Format:                  !*noFmt,
	}

	if *path != "" {
		if *path == refs.ComponentsPrefix {
			if def.Components == nil {
				fail(1, fmt.Errorf("definition does not have components: %q", *path))
			}
			err = generateComponentsStructs(def.Components, options, *outDir, *outFn, *overwrite)
		} else {
			pathDef := getPath(*path, def)
			if pathDef == nil {
				fail(1, fmt.Errorf("unknown path: %q", *path))
			}
			err = generatePathStructs(pathDef, options, *outDir, *outFn, *overwrite)
		}
	} else {
		err = generateDefinitionStructs(def, options, *outDir, *outFn, *overwrite)
	}
	if err != nil {
		fail(1, fmt.Errorf("generate structs: %w", err))
	}
}

func generateDefinitionStructs(def *chioas.Definition, options codegen.SchemaStructOptions, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "structs.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateSchemaStructs(def, f, options)
	}
	return err
}

func generatePathStructs(def *chioas.Path, options codegen.SchemaStructOptions, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "structs.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateSchemaStructs(def, f, options)
	}
	return err
}

func generateComponentsStructs(def *chioas.Components, options codegen.SchemaStructOptions, outDir string, outFn string, overwrite bool) (err error) {
	var f io.WriteCloser
	if f, err = createFile(outFn, outDir, overwrite, "structs.go"); err == nil {
		defer func() {
			_ = f.Close()
		}()
		err = codegen.GenerateSchemaStructs(def, f, options)
	}
	return err
}
