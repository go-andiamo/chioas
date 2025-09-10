package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const (
	cliVersion  = "1.0.0"
	cmdChioas   = "chioas"
	cmdGen      = "gen"
	flagHelp    = "-help"
	flagVersion = "-version"
)

func main() {
	if len(os.Args) < 2 {
		usage("")
	}
	switch os.Args[1] {
	case cmdGen:
		generate(os.Args[2:])
	case flagVersion, "-" + flagVersion, "version", "-v", "--v":
		fmt.Println("CLI version: " + cliVersion)
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(cmdChioas + " version: " + info.Main.Version)
		}
	case flagHelp, "-" + flagHelp, "help", "-h", "--h":
		usage("")
	default:
		usage(fmt.Sprintf("unknown command %q", os.Args[1]))
	}
}

func generate(args []string) {
	if len(args) == 0 {
		usage("")
		return
	}
	switch args[0] {
	case subCmdCode:
		generateCode(args[1:])
	case subCmdStubs:
		generateStubs(args[1:])
	case flagHelp:
		usageGen("")
	default:
		usageGen(fmt.Sprintf("unknown subcommand %q", args[0]))
	}
}

func createFile(filename string, outDir string, overwrite bool, defaultFilename string) (f io.WriteCloser, err error) {
	if outDir == "-" {
		return os.Stdout, nil
	}
	if filename == "" {
		filename = defaultFilename
	}
	if !strings.HasSuffix(filename, ".go") {
		filename += ".go"
	}
	if outDir != "" {
		err = os.MkdirAll(outDir, 0o755)
	} else {
		outDir = "."
	}
	if err == nil {
		path := filepath.Join(outDir, filename)
		if _, err = os.Stat(path); err == nil {
			// file exists
			if !overwrite {
				return nil, fmt.Errorf("file %s already exists (use -overwrite to replace)", path)
			}
		} else if !os.IsNotExist(err) {
			// unexpected error (e.g. permission issue)
			return nil, err
		}
		f, err = os.Create(path)
	}
	return f, err
}

func readDefinition(path string) (result *chioas.Definition, err error) {
	var data []byte
	var ext string
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
		ext = sniff(data)
	} else {
		data, err = os.ReadFile(path)
		ext = strings.ToLower(filepath.Ext(path))
	}
	if err != nil {
		return nil, err
	}
	result = &chioas.Definition{}
	switch ext {
	case ".json", "json":
		err = json.Unmarshal(data, result)
		return result, err
	case ".yml", ".yaml", "yaml":
		err = yaml.Unmarshal(data, result)
		return result, err
	default:
		return nil, errors.New("unable to detect input format (expected JSON or YAML)")
	}
}

func sniff(b []byte) string {
	trim := strings.TrimLeftFunc(string(b), func(r rune) bool { return r == ' ' || r == '\n' || r == '\r' || r == '\t' })
	if strings.HasPrefix(trim, "{") {
		return "json"
	}
	return "yaml"
}

func usageGen(msg string) {
	out := os.Stdout
	if msg != "" {
		out = os.Stderr
		_, _ = fmt.Fprintf(out, "error: %s\n\n", msg)
	}
	_, _ = fmt.Fprintln(out, "Description: "+subCmdCodeDesc)
	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintln(out, "    "+cmdChioas+" "+cmdGen+" "+subCmdCode+" "+flagHelp)
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Description: "+subCmdStubsDesc)
	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintln(out, "    "+cmdChioas+" "+cmdGen+" "+subCmdStubs+" "+flagHelp)
	if msg != "" {
		os.Exit(2)
	} else {
		os.Exit(0)
	}
}

func usage(msg string) {
	out := os.Stdout
	if msg != "" {
		out = os.Stderr
		_, _ = fmt.Fprintf(out, "error: %s\n\n", msg)
	}
	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintln(out, "    "+cmdChioas+" "+cmdGen+" "+flagHelp)
	_, _ = fmt.Fprintln(out, "        Show help for generate commands")
	_, _ = fmt.Fprintln(out, "    "+cmdChioas+" "+flagVersion)
	_, _ = fmt.Fprintln(out, "        Show the current CLI version")
	_, _ = fmt.Fprintln(out, "    "+cmdChioas+" "+flagHelp)
	_, _ = fmt.Fprintln(out, "        Show CLI help")
	if msg != "" {
		os.Exit(2)
	} else {
		os.Exit(0)
	}
}

func fail(code int, err error) {
	_ = json.NewEncoder(os.Stderr).Encode(map[string]any{
		"error": err.Error(),
	})
	_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(code)
}
