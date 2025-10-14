package main

import (
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/codegen"
	"github.com/go-andiamo/flagpole"
	"github.com/go-andiamo/splitter"
	"io"
	"os"
)

const (
	subCmdCode     = "code"
	subCmdCodeDesc = `Generate chioas definition code (e.g. "var definition = chioas.Definition{...}") from existing OAS yaml/json`
)

type genCodeFlags struct {
	CommonFlags
	OutDir          *string `name:"outdir"           alias:"od" usage:"output directory for generated code"                           default:""              example:"[-outdir <dir>]"`
	OutFn           *string `name:"outf"             alias:"of" usage:"output filename for generated code (default: \"definition.go\")" default:"definition.go" example:"[-outf <filename>]"`
	Pkg             *string `name:"pkg"              alias:"pk" usage:"package for generated code (default: \"api\")"                 default:"api"           example:"[-pkg <name>]"`
	VarName         *string `name:"var"              alias:"v"  usage:"name for the top-level variable (default: definition | Definition" default:""          example:"[-var <name>]"`
	ImportAlias     *string `name:"import-alias"     alias:"ia" usage:"import alias for chioas (optional)"                            default:""              example:"[-import-alias <name>|.]"`
	OmitZero        *bool   `name:"omit-zero"        alias:"oz" usage:"omit zero-valued fields (default: false)"                      default:"false"         example:"[-omit-zero]"`
	HoistPaths      *bool   `name:"hoist-paths"      alias:"hp" usage:"hoist top-level vars for paths (default: false)"               default:"false"         example:"[-hoist-paths]"`
	HoistComponents *bool   `name:"hoist-components" alias:"hc" usage:"hoist top-level vars for named components (default: false)"    default:"false"         example:"[-hoist-components]"`
	PublicVars      *bool   `name:"public-vars"      alias:"pv" usage:"export top-level vars (default: false)"                        default:"false"         example:"[-public-vars]"`
	HTTPConsts      *bool   `name:"http-consts"      alias:"ht" usage:"use http.MethodGet, http.Status etc. (default: false)"         default:"false"         example:"[-http-consts]"`
	InlineHandlers  *bool   `name:"inline-handlers"  alias:"ih" usage:"generate stub inline handler funcs within definition (default: false)" default:"false" example:"[-inline-handlers]"`
	Path            *string `name:"path"                        usage:"api path to generate stubs for (optional) - e.g. \"/api/pets\""                        example:"[-path </api/foo>]"`
	CommonSupplementaryFlags
}

var genCodeFlagsParser = flagpole.MustNewParser[genCodeFlags](flagpole.StopOnHelp(true), flagpole.DefaultedOptionals(true), flagpole.IgnoreUnknownFlags(true))

func generateCode(args []string) {
	flags, err := genCodeFlagsParser.Parse(args)
	if err != nil || (flags.Help != nil && *flags.Help) {
		out := os.Stdout
		code := 0
		if err != nil {
			out = os.Stderr
			code = 2
		}
		genCodeFlagsParser.Usage(out, err, cmdGen, subCmdCode)
		os.Exit(code)
	}

	def, err := readDefinition(flags.In)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}
	options := codegen.Options{
		Package:         *flags.Pkg,
		VarName:         *flags.VarName,
		ImportAlias:     *flags.ImportAlias,
		OmitZeroValues:  *flags.OmitZero,
		HoistPaths:      *flags.HoistPaths,
		HoistComponents: *flags.HoistComponents,
		PublicVars:      *flags.PublicVars,
		UseHttpConsts:   *flags.HTTPConsts,
		InlineHandlers:  *flags.InlineHandlers,
		Format:          !*flags.NoFormat,
	}
	if flags.Path != nil {
		pathDef := getPath(*flags.Path, def)
		if pathDef == nil {
			fail(1, fmt.Errorf("unknown path: %q", *flags.Path))
		}
		if err = generatePathCode(pathDef, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite); err != nil {
			fail(1, fmt.Errorf("generate code: %w", err))
		}
	} else if err = generateDefinitionCode(def, options, *flags.OutDir, *flags.OutFn, *flags.Overwrite); err != nil {
		fail(1, fmt.Errorf("generate code: %w", err))
	}
}

func getPath(path string, def *chioas.Definition) (pathDef *chioas.Path) {
	s, _ := splitter.NewSplitter('/', splitter.CurlyBrackets)
	parts, err := s.Split(path, splitter.IgnoreEmptyFirst, splitter.IgnoreEmptyLast)
	if err != nil {
		fail(1, fmt.Errorf("invalid path: %w", err))
	}
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
	return pathDef
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
