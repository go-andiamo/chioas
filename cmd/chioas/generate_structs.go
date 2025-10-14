package main

import (
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/flagpole"
	"io"
	"os"
)

const (
	subCmdStructs     = "structs"
	subCmdStructsDesc = `Generate schema/request/response structs from existing OAS yaml/json`
)

type genStructsFlags struct {
	CommonFlags
	OutDir        *string `name:"outdir"           alias:"od" usage:"output directory for generated code"                           default:""            example:"[-outdir <dir>]"`
	OutFn         *string `name:"outf"             alias:"of" usage:"output filename for generated code (default: \"structs.go\")"  default:"structs.go"  example:"[-outf <filename>]"`
	Pkg           *string `name:"pkg"              alias:"pk" usage:"package for generated code (default: \"api\")"                 default:"api"         example:"[-pkg <name>]"`
	PublicStructs *bool   `name:"public-structs"   alias:"ps" usage:"make structs public (default: false)"                          default:"false"       example:"[-public-structs]"`
	NoRequests    *bool   `name:"no-requests"                 usage:"suppress schemas for requests (default: false)"                default:"false"       example:"[-no-requests]"`
	NoResponses   *bool   `name:"no-responses"                usage:"suppress schemas for responses (default: false)"               default:"false"       example:"[-no-responses]"`
	OASTags       *bool   `name:"oas-tags"         alias:"oas" usage:"oas tags for struct fields (default: false)"                  default:"false"       example:"[-oas-tags]"`
	Keep          *bool   `name:"keep"             alias:"k"  usage:"keep references to components as separate structs (default: false)" default:"false"  example:"[-keep]"`
	Path          *string `name:"path"                        usage:"api path to generate stubs for (optional) - e.g. \"/api/pets\""                      example:"[-path </api/foo>]"`
	GoDoc         *bool   `name:"godoc"            alias:"gd" usage:"include godoc comment for each struct (default: false)"        default:"false"       example:"[-godoc]"`
	CommonSupplementaryFlags
}

var genStructsFlagsParser = flagpole.MustNewParser[genStructsFlags](flagpole.StopOnHelp(true), flagpole.DefaultedOptionals(true), flagpole.IgnoreUnknownFlags(true))

func generateStructs(args []string) {
	flags, err := genStructsFlagsParser.Parse(args)
	if err != nil || (flags.Help != nil && *flags.Help) {
		out := os.Stdout
		code := 0
		if err != nil {
			out = os.Stderr
			code = 2
		}
		genStructsFlagsParser.Usage(out, err, cmdGen, subCmdStructs)
		os.Exit(code)
	}

	def, err := readDefinition(flags.In)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}
	options := codegen.SchemaStructOptions{
		Package:                 *flags.Pkg,
		PublicStructs:           *flags.PublicStructs,
		NoRequests:              *flags.NoRequests,
		NoResponses:             *flags.NoResponses,
		OASTags:                 *flags.OASTags,
		GoDoc:                   *flags.GoDoc,
		KeepComponentProperties: *flags.Keep,
		Format:                  !*flags.NoFormat,
	}
	if flags.Path != nil {
		if *flags.Path == refs.ComponentsPrefix {
			if def.Components == nil {
				fail(1, fmt.Errorf("definition does not have components: %q", *flags.Path))
			}
			err = generateComponentsStructs(def.Components, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite)
		} else {
			pathDef := getPath(*flags.Path, def)
			if pathDef == nil {
				fail(1, fmt.Errorf("unknown path: %q", *flags.Path))
			}
			err = generatePathStructs(pathDef, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite)
		}
	} else {
		err = generateDefinitionStructs(def, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite)
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
