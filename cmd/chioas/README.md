# Chioas CLI

## Installation

Install for CLI:

    go install github.com/go-andiamo/chioas/cmd/chioas@latest

## Description

The Chioas CLI provides bootstrapping commands that enable you to generate code from an existing OAS yaml/json.

Note: This is a bootstrap tool, not production codegen: intended as a starting point, not for CI/CD code generation.

## Basic commands

To check the current version:

    chioas -version

To get CLI help

    chioas -help

## Usage

The generation is broken down into three sub-commands:

1. `gen code` -
   Generate chioas definition code (e.g. `var definition = chioas.Definition{...}`) from existing OAS yaml/json
2. `gen stubs` -
   Generate handler func stubs  (e.g. `func GetRoot(w http.ResponseWriter, r *http.Request) {...}`) from existing OAS yaml/json
3. `gen structs` -
   Generate schema/request/response structs from existing OAS yaml/json

### Usage: `gen code`

Generate chioas definition code (e.g. `var definition = chioas.Definition{...}`) from existing OAS yaml/json

    chioas gen code -in <filename> -outdir <dir> [-outf <filename>] [-pkg <name>] [-var <name>] [-import-alias <name>|.] [-omit-zero] [-hoist-paths] [-hoist-components] [-public-vars] [-http-consts] [-inline-handlers] [-path </api/foo>] [-no-fmt] [-overwrite]

Flags:
- `-help`

  show help
- `-hoist-components`

  hoist top-level vars for named components (optional, default: false)
- `-hoist-paths`

  hoist top-level vars for paths (optional, default: false)
- `-http-consts`

  use http.MethodGet, http.Status (optional, default: false)
- `-import-alias`

  import alias for chioas (optional)
- `-in`

  input definition file (.yaml|.json) or '-' for stdin (required)
- `-inline-handlers`

  generate stub inline handler funcs within the definition (optional, default: false)
- `-no-fmt`

  suppress go formatting of generated code (optional, default: false)
- `-omit-zero`

  omit zero-valued fields (optional, default: false)
- `-outdir`

  output directory for generated code (optional, defaults to current dir)
- `-outf`

  output filename for generated code (optional, default: definition.go)
- `-overwrite`

  allow overwriting existing file (optional, default: false)
- `-path`

  api path to generate code for (optional) - e.g. "/api/pets"
- `-pkg`

  Go package name for generated code (optional, default: api)
- `-public-vars`

  export top-level vars (optional, default: false)
- `-var`

  name for the top-level variable (optional, default: definition or Definition with -public-vars)

### Usage: `gen stubs`

Generate handler func stubs  (e.g. `func GetRoot(w http.ResponseWriter, r *http.Request) {...}`) from existing OAS yaml/json

    chioas gen stubs -in <filename> -outdir <dir> [-outf <filename>] [-pkg <name>] [-public-funcs] [-path-params] [-receiver <receiver-prefix>] [-naming <0|1|2>] [-path </api/foo>] [-godoc] [-no-fmt] [-overwrite]

Flags:
- `-help`

  show help
- `-godoc`

  include godoc comment for each handler func (optional, default: false)
- `-in`

  input definition file (.yaml|.json) or '-' for stdin (required)
- `-naming`

  handler func naming strategy (optional, default: 0)

  0: default naming strategy

  1: try to use OAS operationId

  2: try to use existing handler or x-handler
- `-no-fmt`

  suppress go formatting of generated code (optional, default: false)
- `-outdir`

  output directory for generated code (optional, defaults to current dir)
- `-outf`

  output filename for generated code (optional, default: handlers.go)
- `-overwrite`

  allow overwriting existing file (optional, default: false)
- `-path`

  api path to generate stubs for (optional) - e.g. "/api/pets"
- `-path-params`

  include path params in handler funcs (optional, default: false)
- `-pkg`

  Go package name for generated code (optional, default: api)
- `-public-funcs`

  make handler funcs public (optional, default: false)
- `-receiver`

  make handler funcs with receiver - e.g. "(a *MyApi)" (optional, default: no receiver)

### Usage: `gen structs`

Generate schema/request/response structs from existing OAS yaml/json

    chioas gen structs -in <filename> -outdir <dir> [-outf <filename>] [-pkg <name>] [-public-structs] [-no-requests] [-noResponses] [-oasTags] [-keep] [-godoc] [-no-fmt] [-overwrite]

Flags:
- `-help`

  show help
- `-godoc`

  include godoc comment for each struct (optional, default: false)
- `-in`

  input definition file (.yaml|.json) or '-' for stdin (required)
- `-keep`

  keep references to components as separate structs (optional, default: false)
- `-no-fmt`

  suppress go formatting of generated code (optional, default: false)
- `-no-requests`

  suppress schemas for requests (optional, default: false)
- `-no-responses`

  suppress schemas for responses (optional, default: false)
- `-oas-tags`

  oas tags for struct fields (optional, default: false)
- `-outdir`

  output directory for generated code (optional, defaults to current dir)
- `-outf`

  output filename for generated code (optional, default: structs.go)
- `-overwrite`

  allow overwriting existing file (optional, default: false)
- `-pkg`

  Go package name for generated code (optional, default: api)
- `-public-structs`

  make structs public (optional, default: false)
