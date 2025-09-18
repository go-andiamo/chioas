package main

import (
	"flag"
	"fmt"
	"github.com/go-andiamo/chioas"
	"io"
	"os"
)

const (
	subCmdRefs     = "refs"
	subCmdRefsDesc = "Check OAS yaml/json $refs"
)

func checkRefs(args []string) {
	fs := flag.NewFlagSet("gen code", flag.ContinueOnError)
	in := fs.String("in", "", `input definition file (.yaml|.json) or '-' for stdin (required)`)
	help := fs.Bool("help", false, `show help`)
	egs := []string{
		"-in <filename>",
	}
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args); err != nil {
		fail(2, err)
	}
	if *help {
		usageDetailed("", fs, egs, cmdCheck, subCmdRefs, subCmdRefsDesc)
	}
	if *in == "" {
		usageDetailed("missing -in", fs, egs, cmdCheck, subCmdRefs, subCmdRefsDesc)
	}
	def, err := readDefinition(*in)
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
