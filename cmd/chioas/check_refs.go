package main

import (
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/flagpole"
	"os"
)

const (
	subCmdRefs     = "refs"
	subCmdRefsDesc = "Check OAS yaml/json $refs"
)

type checkRefsFlags struct {
	CommonFlags
}

var checkRefsFlagsParser = flagpole.MustNewParser[checkRefsFlags](flagpole.StopOnHelp(true), flagpole.DefaultedOptionals(true), flagpole.IgnoreUnknownFlags(true))

func checkRefs(args []string) {
	flags, err := checkRefsFlagsParser.Parse(args)
	if err != nil || (flags.Help != nil && *flags.Help) {
		out := os.Stdout
		code := 0
		if err != nil {
			out = os.Stderr
			code = 2
		}
		checkRefsFlagsParser.Usage(out, err, cmdCheck, subCmdRefs)
		os.Exit(code)
	}

	def, err := readDefinition(flags.In)
	if err != nil {
		fail(1, fmt.Errorf("read definition: %w", err))
	}
	errs := def.CheckRefs()
	if len(errs) > 0 {
		for _, err = range errs {
			printError(err)
		}
		os.Exit(2)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "ok, no unresolved or cyclic refs found")
		os.Exit(0)
	}
}

func printError(err any) {
	switch et := err.(type) {
	case *chioas.RefError:
		printRefError(et)
	case error:
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", et.Error())
	default:
		_, _ = fmt.Fprint(os.Stderr, "error: unknown\n")
	}
}

func printRefError(err *chioas.RefError) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
	_, _ = fmt.Fprintf(os.Stderr, "     ref: %s\n", err.Ref)
	if err.Path != "" {
		_, _ = fmt.Fprintf(os.Stderr, "    path: %s\n", err.Path)
	}
	if err.Method != "" {
		_, _ = fmt.Fprintf(os.Stderr, "  method: %s\n", err.Method)
	}
	if err.Item != nil {
		_, _ = fmt.Fprintf(os.Stderr, "    type: %T\n", err.Item)
	}
}
