package codegen

import (
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"sort"
	"strconv"
)

type ItemType interface {
	chioas.Definition | *chioas.Definition |
		chioas.Path | *chioas.Path |
		chioas.Paths |
		chioas.Method | *chioas.Method |
		chioas.Schema | *chioas.Schema |
		chioas.Components | *chioas.Components
}

// GenerateCode writes Go source that reconstructs the given item (a chioas.Definition
// or chioas.Path) to w using the supplied Options.
//
// Typical use: unmarshal an existing OpenAPI (YAML/JSON) into a chioas.Definition value, then
// emit equivalent Go code so the spec can live on as code.
//
// Notes:
//   - DocOptions are never emitted.
//   - Output is deterministic (map keys & slices are sorted)
//   - Internal $refs are normalized to bare component names
//     (e.g., "#/components/schemas/User" → "User"). External refs
//     or refs in other areas are left as-is. No dereferencing is performed.
//   - This is a bootstrap tool, not production codegen: intended as a starting point,
//     not for CI/CD code generation.
//   - If Options.SkipPrologue is false, a package clause and a single chioas import (optionally
//     aliased) are emitted. When true, the caller is responsible for imports.
//
// Errors:
//   - Returns the first write error encountered. It does not close w.
func GenerateCode[T ItemType](item T, w io.Writer, opts Options) error {
	cw := newCodeWriter(w, opts)
	cw.writePrologue()
	switch it := any(item).(type) {
	case chioas.Definition:
		generateDefinition(it, false, cw)
	case *chioas.Definition:
		generateDefinition(*it, true, cw)
	case chioas.Path:
		cw.writeVarStart(cw.opts.topVarName(), typePath, false)
		generatePath(0, it, cw)
		cw.writeLine(0, "}", true)
	case *chioas.Path:
		cw.writeVarStart(cw.opts.topVarName(), typePath, true)
		generatePath(0, *it, cw)
		cw.writeLine(0, "}", true)
	case chioas.Paths:
		cw.writeVarStart(cw.opts.topVarName(), typePaths, false)
		generatePathsInner(0, it, cw)
		cw.writeLine(0, "}", true)
	case chioas.Method:
		cw.writeVarStart(cw.opts.topVarName(), typeMethod, false)
		generateMethodInner(0, it, cw)
		cw.writeLine(0, "}", true)
	case *chioas.Method:
		cw.writeVarStart(cw.opts.topVarName(), typeMethod, true)
		generateMethodInner(0, *it, cw)
		cw.writeLine(0, "}", true)
	case chioas.Schema:
		cw.writeVarStart(cw.opts.topVarName(), typeSchema, false)
		generateSchema(0, &it, cw)
		cw.writeLine(0, "}", true)
	case *chioas.Schema:
		cw.writeVarStart(cw.opts.topVarName(), typeSchema, true)
		generateSchema(0, it, cw)
		cw.writeLine(0, "}", true)
	case chioas.Components:
		if !opts.HoistComponents {
			cw.writeVarStart(cw.opts.topVarName(), typeComponents, false)
			generateComponentsInner(0, &it, cw)
			cw.writeLine(0, "}", true)
		} else {
			generateComponentsVars(&it, cw, true, false)
		}
	case *chioas.Components:
		if !opts.HoistComponents {
			cw.writeVarStart(cw.opts.topVarName(), typeComponents, true)
			generateComponentsInner(0, it, cw)
			cw.writeLine(0, "}", true)
		} else {
			generateComponentsVars(it, cw, true, true)
		}
	}
	return cw.format()
}

func generateDefinition(def chioas.Definition, ptr bool, cw *codeWriter) {
	cw.writeVarStart(cw.opts.topVarName(), typeDefinition, ptr)
	generateInfo(1, def.Info, cw)
	if len(def.Servers) > 0 {
		cw.writeLine(1, "Servers: "+cw.opts.alias()+typeServers+"{", false)
		ks := maps.Keys(def.Servers)
		sort.Strings(ks)
		for _, k := range ks {
			cw.writeKey(2, k)
			s := def.Servers[k]
			writeZeroField(cw, 3, "Description", s.Description)
			cw.writeExtensions(3, s.Extensions)
			writeZeroField(cw, 3, "Comment", s.Comment)
			cw.writeEnd(2, "},")
		}
		cw.writeEnd(1, "},")
	}
	if len(def.Tags) > 0 {
		cw.writeLine(1, "Tags: "+cw.opts.alias()+typeTags+"{", false)
		for _, tg := range def.Tags {
			cw.writeLine(2, "{", false)
			writeZeroField(cw, 3, "Name", tg.Name)
			writeZeroField(cw, 3, "Description", tg.Description)
			if tg.ExternalDocs != nil {
				cw.writeLine(3, "ExternalDocs: &"+cw.opts.alias()+typeExternalDocs+"{", false)
				writeZeroField(cw, 4, "Description", tg.ExternalDocs.Description)
				writeZeroField(cw, 4, "Url", tg.ExternalDocs.Url)
				cw.writeExtensions(4, tg.ExternalDocs.Extensions)
				writeZeroField(cw, 4, "Comment", tg.ExternalDocs.Comment)
				cw.writeEnd(3, "},")
			}
			cw.writeExtensions(3, tg.Extensions)
			writeZeroField(cw, 3, "Comment", tg.Comment)
			cw.writeEnd(2, "},")
		}
		cw.writeEnd(1, "},")
	}
	generateMethods(1, def.Methods, cw)
	var paths []string
	if !cw.opts.HoistPaths {
		generatePaths(1, def.Paths, cw)
	} else {
		cw.writeCollectionFieldStart(1, typePaths, typePaths)
		paths = maps.Keys(def.Paths)
		sort.Strings(paths)
		deduper := newNameDeDuper()
		for _, k := range paths {
			useName := k
			if useName == "" || useName == "/" {
				useName = "Root"
			}
			cw.writeLine(2, strconv.Quote(k)+": "+deduper.take(cw.opts.varName("Path", useName))+",", false)
		}
		cw.writeLine(1, "},", false)
	}
	if len(def.Security) > 0 {
		cw.writeLine(1, "Security: "+cw.opts.alias()+typeSecuritySchemes+"{", false)
		for _, ss := range def.Security {
			cw.writeLine(2, "{", false)
			generateSecurityScheme(2, ss, cw)
			cw.writeEnd(2, "},")
		}
		cw.writeEnd(1, "},")
	}
	if def.Components != nil {
		generateComponents(1, def.Components, cw)
	}
	cw.writeExtensions(1, def.Extensions)
	writeZeroField(cw, 1, "Comment", def.Comment)
	writeZeroField(cw, 1, "AutoHeadMethods", def.AutoHeadMethods)
	writeZeroField(cw, 1, "AutoOptionsMethods", def.AutoOptionsMethods)
	writeZeroField(cw, 1, "RootAutoOptionsMethod", def.RootAutoOptionsMethod)
	writeZeroField(cw, 1, "AutoMethodNotAllowed", def.AutoMethodNotAllowed)
	cw.writeLine(0, "}", true)
	if cw.opts.HoistPaths {
		cw.writeLine(0, "var (", false)
		// write out the actual path vars...
		deduper := newNameDeDuper()
		for _, k := range paths {
			useName := k
			if useName == "" || useName == "/" {
				useName = "Root"
			}
			cw.writeLine(1, deduper.take(cw.opts.varName("Path", useName))+" = "+cw.opts.alias()+typePath+"{", false)
			generatePath(1, def.Paths[k], cw)
			cw.writeLine(1, "}", false)
		}
		cw.writeLine(0, ")", true)
	}
	if cw.opts.HoistComponents && def.Components != nil {
		generateComponentsVars(def.Components, cw, false, false)
	}
}

func generateComponentsVars(def *chioas.Components, cw *codeWriter, topVar bool, topPtr bool) {
	deduper := newNameDeDuper()
	deDupe := func(kind, name string) string {
		return deduper.take(cw.opts.varName(kind, name))
	}
	const (
		varSchema         = "Schema"
		varRequest        = "Request"
		varResponse       = "Response"
		varExample        = "Example"
		varParameter      = "Parameter"
		varSecurityScheme = "SecurityScheme"
	)
	// start vars...
	cw.writeLine(0, "var (", false)
	// write components var...
	if topVar {
		amp := ""
		if topPtr {
			amp = "&"
		}
		cw.writeLine(1, cw.opts.topVarName()+" = "+amp+cw.opts.alias()+typeComponents+"{", false)
	} else {
		cw.writeLine(1, cw.opts.varName("", "components")+" = &"+cw.opts.alias()+typeComponents+"{", false)
	}
	var schemas chioas.Schemas
	if len(def.Schemas) > 0 {
		cw.writeCollectionFieldStart(2, typeSchemas, typeSchemas)
		schemas = append(schemas, def.Schemas...)
		slices.SortStableFunc(schemas, func(a, b chioas.Schema) bool {
			return a.Name < b.Name
		})
		for _, s := range schemas {
			cw.writeLine(3, deDupe(varSchema, s.Name)+",", false)
		}
		cw.writeEnd(2, "},")
	}
	var requests []string
	if len(def.Requests) > 0 {
		cw.writeCollectionFieldStart(2, "Requests", typeCommonRequests)
		requests = maps.Keys(def.Requests)
		sort.Strings(requests)
		for _, k := range requests {
			cw.writeLine(3, strconv.Quote(k)+": "+deDupe(varRequest, k)+",", false)
		}
		cw.writeEnd(2, "},")
	}
	var responses []string
	if len(def.Responses) > 0 {
		cw.writeCollectionFieldStart(2, "Responses", typeCommonResponses)
		responses = maps.Keys(def.Responses)
		sort.Strings(responses)
		for _, k := range responses {
			cw.writeLine(3, strconv.Quote(k)+": "+deDupe(varResponse, k)+",", false)
		}
		cw.writeEnd(2, "},")
	}
	var examples chioas.Examples
	if len(def.Examples) > 0 {
		cw.writeCollectionFieldStart(2, typeExamples, typeExamples)
		examples = append(examples, def.Examples...)
		slices.SortStableFunc(examples, func(a, b chioas.Example) bool {
			return a.Name < b.Name
		})
		for _, eg := range examples {
			cw.writeLine(3, deDupe(varExample, eg.Name)+",", false)
		}
		cw.writeEnd(2, "},")
	}
	var params []string
	if len(def.Parameters) > 0 {
		cw.writeCollectionFieldStart(2, "Parameters", typeCommonParameters)
		params = maps.Keys(def.Parameters)
		sort.Strings(params)
		for _, k := range params {
			cw.writeLine(3, strconv.Quote(k)+": "+deDupe(varParameter, k)+",", false)
		}
		cw.writeEnd(2, "},")
	}
	var secSchemes chioas.SecuritySchemes
	if len(def.SecuritySchemes) > 0 {
		cw.writeCollectionFieldStart(2, typeSecuritySchemes, typeSecuritySchemes)
		secSchemes = append(secSchemes, def.SecuritySchemes...)
		slices.SortStableFunc(secSchemes, func(a, b chioas.SecurityScheme) bool {
			return a.Name < b.Name
		})
		for _, s := range secSchemes {
			cw.writeLine(3, deDupe(varSecurityScheme, s.Name)+",", false)
		}
		cw.writeEnd(2, "},")
	}
	cw.writeExtensions(2, def.Extensions)
	writeZeroField(cw, 2, "Comment", def.Comment)
	cw.writeEnd(1, "}")
	// write components parts vars...
	deduper.clear()
	// schemas...
	for _, s := range schemas {
		cw.writeLine(1, deDupe(varSchema, s.Name)+" = "+cw.opts.alias()+typeSchema+"{", false)
		generateSchema(1, &s, cw)
		cw.writeEnd(1, "}")
	}
	// requests...
	for _, k := range requests {
		cw.writeLine(1, deDupe(varRequest, k)+" = "+cw.opts.alias()+typeRequest+"{", false)
		r := def.Requests[k]
		generateRequest(1, &r, cw)
		cw.writeEnd(1, "}")
	}
	// responses...
	for _, k := range responses {
		cw.writeLine(1, deDupe(varResponse, k)+" = "+cw.opts.alias()+typeResponse+"{", false)
		generateResponse(1, def.Responses[k], cw)
		cw.writeEnd(1, "}")
	}
	// examples...
	for _, eg := range examples {
		cw.writeLine(1, deDupe(varExample, eg.Name)+" = "+cw.opts.alias()+typeExample+"{", false)
		generateExample(1, eg, cw)
		cw.writeEnd(1, "}")
	}
	// parameters...
	for _, k := range params {
		cw.writeLine(1, deDupe(varParameter, k)+" = "+cw.opts.alias()+typeCommonParameter+"{", false)
		generateCommonParam(1, def.Parameters[k], cw)
		cw.writeEnd(1, "}")
	}
	// security schemes...
	for _, s := range secSchemes {
		cw.writeLine(1, deDupe(varSecurityScheme, s.Name)+" = "+cw.opts.alias()+typeSecurityScheme+"{", false)
		generateSecurityScheme(1, s, cw)
		cw.writeEnd(1, "}")
	}
	// end vars...
	cw.writeLine(0, ")", true)
}

func generateComponents(indent int, def *chioas.Components, cw *codeWriter) {
	if cw.opts.HoistComponents {
		cw.writeLine(indent, "Components: "+cw.opts.varName("", "Components")+",", false)
		return
	}
	cw.writeLine(indent, "Components: &"+cw.opts.alias()+typeComponents+"{", false)
	generateComponentsInner(indent, def, cw)
	cw.writeEnd(indent, "},")
}

func generateComponentsInner(indent int, def *chioas.Components, cw *codeWriter) {
	if len(def.Schemas) > 0 {
		cw.writeCollectionFieldStart(indent+1, typeSchemas, typeSchemas)
		ss := append(chioas.Schemas{}, def.Schemas...)
		slices.SortStableFunc(ss, func(a, b chioas.Schema) bool {
			return a.Name < b.Name
		})
		for _, s := range ss {
			cw.writeLine(indent+2, "{", false)
			generateSchema(indent+2, &s, cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	if len(def.Requests) > 0 {
		cw.writeCollectionFieldStart(indent+1, "Requests", typeCommonRequests)
		ks := maps.Keys(def.Requests)
		sort.Strings(ks)
		for _, k := range ks {
			cw.writeKey(indent+2, k)
			r := def.Requests[k]
			generateRequest(indent+2, &r, cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	if len(def.Responses) > 0 {
		cw.writeCollectionFieldStart(indent+1, "Responses", typeCommonResponses)
		ks := maps.Keys(def.Responses)
		sort.Strings(ks)
		for _, k := range ks {
			cw.writeKey(indent+2, k)
			generateResponse(indent+2, def.Responses[k], cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	if len(def.Examples) > 0 {
		cw.writeCollectionFieldStart(indent+1, typeExamples, typeExamples)
		egs := append(chioas.Examples{}, def.Examples...)
		slices.SortStableFunc(egs, func(a, b chioas.Example) bool {
			return a.Name < b.Name
		})
		for _, eg := range egs {
			cw.writeLine(indent+2, "{", false)
			generateExample(indent+2, eg, cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	if len(def.Parameters) > 0 {
		cw.writeCollectionFieldStart(indent+1, "Parameters", typeCommonParameters)
		ks := maps.Keys(def.Parameters)
		sort.Strings(ks)
		for _, k := range ks {
			cw.writeKey(indent+2, k)
			generateCommonParam(indent+2, def.Parameters[k], cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	if len(def.SecuritySchemes) > 0 {
		cw.writeCollectionFieldStart(indent+1, typeSecuritySchemes, typeSecuritySchemes)
		ss := append(chioas.SecuritySchemes{}, def.SecuritySchemes...)
		slices.SortStableFunc(ss, func(a, b chioas.SecurityScheme) bool {
			return a.Name < b.Name
		})
		for _, s := range ss {
			cw.writeLine(indent+2, "{", false)
			generateSecurityScheme(indent+3, s, cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	cw.writeExtensions(indent+2, def.Extensions)
	writeZeroField(cw, indent+2, "Comment", def.Comment)
}

func generateInfo(indent int, def chioas.Info, cw *codeWriter) {
	if hasNonZeroValues(def.Title, def.Description, def.Version, def.TermsOfService, def.Comment,
		def.Contact, def.License, def.Extensions, def.ExternalDocs) {
		cw.writeLine(indent+1, "Info: "+cw.opts.alias()+typeInfo+"{", false)
		writeZeroField(cw, indent+2, "Title", def.Title)
		writeZeroField(cw, indent+2, "Description", def.Description)
		writeZeroField(cw, indent+2, "Version", def.Version)
		writeZeroField(cw, indent+2, "TermsOfService", def.TermsOfService)
		if def.Contact != nil {
			cw.writeLine(indent+2, "Contact: &"+cw.opts.alias()+typeContact+"{", false)
			writeZeroField(cw, indent+3, "Name", def.Contact.Name)
			writeZeroField(cw, indent+3, "Url", def.Contact.Url)
			writeZeroField(cw, indent+3, "Email", def.Contact.Email)
			cw.writeExtensions(indent+3, def.Contact.Extensions)
			writeZeroField(cw, indent+3, "Comment", def.Contact.Comment)
			cw.writeEnd(indent+2, "},")
		}
		if def.License != nil {
			cw.writeLine(indent+2, "License: &"+cw.opts.alias()+typeLicense+"{", false)
			writeZeroField(cw, indent+3, "Name", def.License.Name)
			writeZeroField(cw, indent+3, "Url", def.License.Url)
			cw.writeExtensions(indent+3, def.License.Extensions)
			writeZeroField(cw, indent+3, "Comment", def.License.Comment)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeExtensions(indent+2, def.Extensions)
		if def.ExternalDocs != nil {
			cw.writeLine(indent+2, "ExternalDocs: &"+cw.opts.alias()+typeExternalDocs+"{", false)
			writeZeroField(cw, indent+3, "Description", def.ExternalDocs.Description)
			writeZeroField(cw, indent+3, "Url", def.ExternalDocs.Url)
			cw.writeExtensions(indent+3, def.ExternalDocs.Extensions)
			writeZeroField(cw, indent+3, "Comment", def.ExternalDocs.Comment)
			cw.writeEnd(indent+2, "},")
		}
		writeZeroField(cw, indent+2, "Comment", def.Comment)
		cw.writeEnd(indent+1, "},")
	}
}

func generatePath(indent int, def chioas.Path, cw *codeWriter) {
	generateMethods(indent+1, def.Methods, cw)
	generatePaths(indent+1, def.Paths, cw)
	writeZeroField(cw, indent+1, "Tag", def.Tag)
	if len(def.PathParams) > 0 {
		cw.writeLine(indent+1, "PathParams: "+cw.opts.alias()+typePathParams+"{", false)
		ks := maps.Keys(def.PathParams)
		sort.Strings(ks)
		for _, k := range ks {
			cw.writeKey(indent+2, k)
			generatePathParam(indent+2, def.PathParams[k], cw)
			cw.writeLine(indent+2, "},", false)
		}
		cw.writeEnd(indent+1, "},")
	}
	writeZeroField(cw, indent+1, "HideDocs", def.HideDocs)
	cw.writeExtensions(indent+1, def.Extensions)
	writeZeroField(cw, indent+1, "Comment", def.Comment)
	writeZeroField(cw, indent+1, "AutoOptionsMethod", def.AutoOptionsMethod)
}

func generatePaths(indent int, paths chioas.Paths, cw *codeWriter) {
	if len(paths) > 0 {
		cw.writeCollectionFieldStart(indent, typePaths, typePaths)
		generatePathsInner(indent, paths, cw)
		cw.writeLine(indent, "},", false)
	}
}

func generatePathsInner(indent int, paths chioas.Paths, cw *codeWriter) {
	ks := maps.Keys(paths)
	sort.Strings(ks)
	for _, k := range ks {
		cw.writeKey(indent+1, k)
		generatePath(indent+1, paths[k], cw)
		cw.writeLine(indent+1, "},", false)
	}
}

func generateMethods(indent int, methods chioas.Methods, cw *codeWriter) {
	if len(methods) > 0 {
		cw.writeCollectionFieldStart(indent, typeMethods, typeMethods)
		sms := maps.Keys(methods)
		sort.Slice(sms, func(i, j int) bool {
			return compareMethods(sms[i], sms[j])
		})
		for _, m := range sms {
			generateMethod(indent+1, m, methods[m], cw)
		}
		cw.writeLine(indent, "},", false)
	}
}

func generateMethod(indent int, method string, def chioas.Method, cw *codeWriter) {
	cw.writeLine(indent, cw.opts.translateMethod(method)+": {", false)
	if cw.opts.InlineHandlers {
		generateMethodHandler(indent, cw)
	}
	generateMethodInner(indent, def, cw)
	cw.writeLine(indent, "},", false)
}

func generateMethodHandler(indent int, cw *codeWriter) {
	cw.writeLine(indent+1, "Handler: func"+string(handlerSignature), false)
	cw.writeLine(indent+2, `// TODO implement me`, false)
	cw.writeLine(indent+2, `panic("implement me!")`, false)
	cw.writeLine(indent+1, "},", false)
}

func generateMethodInner(indent int, def chioas.Method, cw *codeWriter) {
	writeZeroField(cw, indent+1, "Description", def.Description)
	writeZeroField(cw, indent+1, "Summary", def.Summary)
	if hs, ok := def.Handler.(string); ok {
		// only output handler if a string
		writeZeroField(cw, indent+1, "Handler", hs)
	}
	writeZeroField(cw, indent+1, "OperationId", def.OperationId)
	writeZeroField(cw, indent+1, "Tag", def.Tag)
	if len(def.QueryParams) > 0 {
		cw.writeLine(indent+1, "QueryParams: "+cw.opts.alias()+typeQueryParams+"{", false)
		for _, qp := range def.QueryParams {
			cw.writeLine(indent+2, "{", false)
			generateQueryParam(indent+2, qp, cw)
			cw.writeLine(indent+2, "},", false)
		}
		cw.writeEnd(indent+1, "},")
	}
	if def.Request != nil {
		cw.writeLine(indent+1, "Request: &"+cw.opts.alias()+typeRequest+"{", false)
		generateRequest(indent+1, def.Request, cw)
		cw.writeEnd(indent+1, "},")
	}
	if len(def.Responses) > 0 {
		cw.writeLine(indent+1, "Responses: "+cw.opts.alias()+typeResponses+"{", false)
		ks := maps.Keys(def.Responses)
		slices.Sort(ks)
		for _, k := range ks {
			cw.writeLine(indent+2, cw.opts.translateStatus(k)+": {", false)
			generateResponse(indent+2, def.Responses[k], cw)
			cw.writeEnd(indent+2, "},")
		}
		cw.writeEnd(indent+1, "},")
	}
	writeZeroField(cw, indent+1, "Deprecated", def.Deprecated)
	if len(def.Security) > 0 {
		cw.writeLine(indent+1, "Security: "+cw.opts.alias()+typeSecuritySchemes+"{", false)
		for _, s := range def.Security {
			cw.writeLine(indent+2, "{", false)
			cw.writeLine(indent+3, "Name: "+strconv.Quote(s.Name)+",", false)
			cw.writeLine(indent+2, "},", false)
		}
		cw.writeEnd(indent+1, "},")
	}
	writeZeroField(cw, indent+1, "OptionalSecurity", def.OptionalSecurity)
	cw.writeExtensions(indent+1, def.Extensions)
	writeZeroField(cw, indent+1, "Comment", def.Comment)
	writeZeroField(cw, indent+1, "HideDocs", def.HideDocs)
}

func generateRequest(indent int, def *chioas.Request, cw *codeWriter) {
	if def.Ref != "" {
		cw.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.RequestBodies, def.Ref))+",", false)
	} else {
		writeZeroField(cw, indent+1, "Description", def.Description)
		writeZeroField(cw, indent+1, "Required", def.Required)
		writeZeroField(cw, indent+1, "ContentType", def.ContentType)
		generateAlternativeContentTypes(indent, def.AlternativeContentTypes, cw)
		if s := def.Schema; s != nil {
			generateVaryingSchema(indent, def.Schema, cw)
		} else if def.SchemaRef != "" {
			cw.writeSchemaRef(indent+1, def.SchemaRef)
		}
		writeZeroField(cw, indent+1, "IsArray", def.IsArray)
		generateExamples(indent, def.Examples, cw)
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generateResponse(indent int, def chioas.Response, cw *codeWriter) {
	if def.Ref != "" {
		cw.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.Responses, def.Ref))+",", false)
	} else {
		writeZeroField(cw, indent+1, "Description", def.Description)
		writeZeroField(cw, indent+1, "NoContent", def.NoContent)
		writeZeroField(cw, indent+1, "ContentType", def.ContentType)
		generateAlternativeContentTypes(indent, def.AlternativeContentTypes, cw)
		if s := def.Schema; s != nil {
			generateVaryingSchema(indent, def.Schema, cw)
		} else if def.SchemaRef != "" {
			cw.writeSchemaRef(indent+1, def.SchemaRef)
		}
		writeZeroField(cw, indent+1, "IsArray", def.IsArray)
		generateExamples(indent, def.Examples, cw)
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generateAlternativeContentTypes(indent int, cts chioas.ContentTypes, cw *codeWriter) {
	if len(cts) > 0 {
		cw.writeLine(indent+1, "AlternativeContentTypes: "+cw.opts.alias()+typeContentTypes+"{", false)
		ks := maps.Keys(cts)
		sort.Strings(ks)
		for _, k := range ks {
			cw.writeKey(indent+2, k)
			generateContentType(indent+2, cts[k], cw)
			cw.writeLine(indent+2, "},", false)
		}
		cw.writeEnd(indent+1, "},")
	}
}

func generateContentType(indent int, def chioas.ContentType, cw *codeWriter) {
	if s := def.Schema; s != nil {
		generateVaryingSchema(indent, def.Schema, cw)
	} else if def.SchemaRef != "" {
		cw.writeSchemaRef(indent+1, def.SchemaRef)
	}
	writeZeroField(cw, indent+1, "IsArray", def.IsArray)
	generateExamples(indent, def.Examples, cw)
	cw.writeExtensions(indent+1, def.Extensions)
	writeZeroField(cw, indent+1, "Comment", def.Comment)
}

func generateVaryingSchema(indent int, s any, cw *codeWriter) {
	switch schema := s.(type) {
	case chioas.Schema:
		cw.writeLine(indent+1, "Schema: "+cw.opts.alias()+typeSchema+"{", false)
		generateSchema(indent+1, &schema, cw)
		cw.writeLine(indent+1, "},", false)
	case *chioas.Schema:
		if schema != nil {
			cw.writeLine(indent+1, "Schema: &"+cw.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, schema, cw)
			cw.writeLine(indent+1, "},", false)
		}
	case chioas.SchemaConverter:
		cw.writeLine(indent+1, "Schema: &"+cw.opts.alias()+typeSchema+"{", false)
		ts := schema.ToSchema()
		if ts != nil {
			generateSchema(indent+1, ts, cw)
		}
		cw.writeLine(indent+1, "},", false)
	default:
		if fs, err := chioas.SchemaFrom(s); err == nil {
			cw.writeLine(indent+1, "Schema: &"+cw.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, fs, cw)
			cw.writeLine(indent+1, "},", false)
		} else {
			// note: this will produce bad code - but indicates needs manual attention
			cw.writeLine(indent+1, fmt.Sprintf("Schema: %T,", s), false)
		}
	}
}

func generateExamples(indent int, egs chioas.Examples, cw *codeWriter) {
	if len(egs) > 0 {
		cw.writeLine(indent+1, "Examples: "+cw.opts.alias()+typeExamples+"{", false)
		for _, eg := range egs {
			cw.writeLine(indent+2, "{", false)
			generateExample(indent+2, eg, cw)
			cw.writeLine(indent+2, "},", false)
		}
		cw.writeEnd(indent+1, "},")
	}
}

func generateExample(indent int, def chioas.Example, cw *codeWriter) {
	if def.ExampleRef != "" {
		cw.writeLine(indent+1, "ExampleRef: "+strconv.Quote(refs.Normalize(tags.Examples, def.ExampleRef))+",", false)
	} else {
		writeZeroField(cw, indent+1, "Name", def.Name)
		writeZeroField(cw, indent+1, "Summary", def.Summary)
		writeZeroField(cw, indent+1, "Description", def.Description)
		if def.Value != nil {
			cw.writeStart(indent+1, "Value: ")
			cw.writeValue(indent+1, def.Value)
		}
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generatePathParam(indent int, def chioas.PathParam, cw *codeWriter) {
	if def.Ref != "" {
		cw.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.Parameters, def.Ref))+",", false)
	} else {
		writeZeroField(cw, indent+1, "Description", def.Description)
		if def.Example != nil {
			cw.writeStart(indent+1, "Example: ")
			cw.writeValue(indent+1, def.Example)
		}
		if def.Schema != nil {
			cw.writeLine(indent+1, "Schema: &"+cw.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, def.Schema, cw)
			cw.writeEnd(indent+1, "},")
		} else if def.SchemaRef != "" {
			cw.writeSchemaRef(indent+1, def.SchemaRef)
		}
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generateQueryParam(indent int, def chioas.QueryParam, cw *codeWriter) {
	if def.Ref != "" {
		cw.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.Parameters, def.Ref))+",", false)
	} else {
		writeZeroField(cw, indent+1, "Name", def.Name)
		writeZeroField(cw, indent+1, "Description", def.Description)
		writeZeroField(cw, indent+1, "Required", def.Required)
		writeZeroField(cw, indent+1, "In", def.In)
		if def.Example != nil {
			cw.writeStart(indent+1, "Example: ")
			cw.writeValue(indent+1, def.Example)
		}
		if def.Schema != nil {
			cw.writeLine(indent+1, "Schema: &"+cw.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, def.Schema, cw)
			cw.writeEnd(indent+1, "},")
		} else if def.SchemaRef != "" {
			cw.writeSchemaRef(indent+1, def.SchemaRef)
		}
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generateCommonParam(indent int, def chioas.CommonParameter, cw *codeWriter) {
	writeZeroField(cw, indent+1, "Name", def.Name)
	writeZeroField(cw, indent+1, "Description", def.Description)
	writeZeroField(cw, indent+1, "Required", def.Required)
	writeZeroField(cw, indent+1, "In", def.In)
	if def.Example != nil {
		cw.writeStart(indent+1, "Example: ")
		cw.writeValue(indent+1, def.Example)
	}
	if def.Schema != nil {
		cw.writeLine(indent+1, "Schema: &"+cw.opts.alias()+typeSchema+"{", false)
		generateSchema(indent+1, def.Schema, cw)
		cw.writeEnd(indent+1, "},")
	} else if def.SchemaRef != "" {
		cw.writeSchemaRef(indent+1, def.SchemaRef)
	}
	cw.writeExtensions(indent+1, def.Extensions)
	writeZeroField(cw, indent+1, "Comment", def.Comment)
}

func generateSchema(indent int, def *chioas.Schema, cw *codeWriter) {
	if def.SchemaRef != "" {
		cw.writeSchemaRef(indent+1, def.SchemaRef)
	} else {
		writeZeroField(cw, indent+1, "Name", def.Name)
		writeZeroField(cw, indent+1, "Description", def.Description)
		writeZeroField(cw, indent+1, "Type", def.Type)
		writeZeroField(cw, indent+1, "Format", def.Format)
		if len(def.RequiredProperties) > 0 {
			cw.writeStart(indent+1, "RequiredProperties: ")
			cw.writeValue(indent+1, def.RequiredProperties)
		}
		if len(def.Properties) > 0 {
			cw.writeLine(indent+1, "Properties: "+cw.opts.alias()+typeProperties+"{", false)
			for _, p := range def.Properties {
				cw.writeLine(indent+2, "{", false)
				generateProperty(indent+2, p, cw)
				cw.writeLine(indent+2, "},", false)
			}
			cw.writeEnd(indent+1, "},")
		}
		if def.Default != nil {
			cw.writeStart(indent+1, "Default: ")
			cw.writeValue(indent+1, def.Default)
		}
		if def.Example != nil {
			cw.writeStart(indent+1, "Example: ")
			cw.writeValue(indent+1, def.Example)
		}
		if len(def.Enum) > 0 {
			cw.writeStart(indent+1, "Enum: ")
			cw.writeValue(indent+1, def.Enum)
		}
		if def.Discriminator != nil {
			cw.writeLine(indent+1, "Discriminator: &"+cw.opts.alias()+typeDiscriminator+"{", false)
			generateDiscriminator(indent+1, def.Discriminator, cw)
			cw.writeEnd(indent+1, "},")
		}
		if def.Ofs != nil {
			cw.writeLine(indent+1, "Ofs: &"+cw.opts.alias()+typeOfs+"{", false)
			generateOfs(indent+1, def.Ofs, cw)
			cw.writeEnd(indent+1, "},")
		}
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generateDiscriminator(indent int, def *chioas.Discriminator, cw *codeWriter) {
	writeZeroField(cw, indent+1, "PropertyName", def.PropertyName)
	cw.writeLine(indent+1, "Mapping: map[string]string{", false)
	ks := maps.Keys(def.Mapping)
	sort.Strings(ks)
	for _, k := range ks {
		cw.writeLine(indent+2, strconv.Quote(k)+": "+strconv.Quote(refs.Normalize(tags.Schemas, def.Mapping[k]))+",", false)
	}
	cw.writeEnd(indent+1, "},")
	cw.writeExtensions(indent+1, def.Extensions)
	writeZeroField(cw, indent+1, "Comment", def.Comment)
}

func generateOfs(indent int, def *chioas.Ofs, cw *codeWriter) {
	switch def.OfType {
	case chioas.AllOf:
		cw.writeLine(indent+1, "OfType: "+cw.opts.alias()+"AllOf,", false)
	case chioas.AnyOf:
		cw.writeLine(indent+1, "OfType: "+cw.opts.alias()+"AnyOf,", false)
	case chioas.OneOf:
		cw.writeLine(indent+1, "OfType: "+cw.opts.alias()+"OneOf,", false)
	default:
		cw.writeLine(indent+1, "OfType: "+strconv.Itoa(int(def.OfType))+",", false)
	}
	cw.writeLine(indent+1, "Of: []"+cw.opts.alias()+typeOfSchema+"{", false)
	for _, ofs := range def.Of {
		cw.writeLine(indent+2, "&"+cw.opts.alias()+typeOf+"{", false)
		if ofs.IsRef() {
			cw.writeSchemaRef(indent+3, ofs.Ref())
		} else {
			s := ofs.Schema()
			if s == nil {
				s = &chioas.Schema{}
			}
			cw.writeLine(indent+3, "SchemaDef: &"+cw.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+3, s, cw)
			cw.writeEnd(indent+3, "},")
		}
		cw.writeEnd(indent+2, "},")
	}
	cw.writeEnd(indent+1, "},")
}

func generateProperty(indent int, def chioas.Property, cw *codeWriter) {
	if def.SchemaRef != "" {
		cw.writeSchemaRef(indent+1, def.SchemaRef)
	} else {
		writeZeroField(cw, indent+1, "Name", def.Name)
		writeZeroField(cw, indent+1, "Description", def.Description)
		writeZeroField(cw, indent+1, "Type", def.Type)
		writeZeroField(cw, indent+1, "ItemType", def.ItemType)
		if len(def.Properties) > 0 {
			cw.writeLine(indent+1, "Properties: "+cw.opts.alias()+typeProperties+"{", false)
			for _, p := range def.Properties {
				cw.writeLine(indent+2, "{", false)
				generateProperty(indent+2, p, cw)
				cw.writeLine(indent+2, "},", false)
			}
			cw.writeEnd(indent+1, "},")
		}
		writeZeroField(cw, indent+1, "Required", def.Required)
		writeZeroField(cw, indent+1, "Format", def.Format)
		if def.Example != nil {
			cw.writeStart(indent+1, "Example: ")
			cw.writeValue(indent+1, def.Example)
		}
		if len(def.Enum) > 0 {
			cw.writeStart(indent+1, "Enum: ")
			cw.writeValue(indent+1, def.Enum)
		}
		writeZeroField(cw, indent+1, "Deprecated", def.Deprecated)
		if hasNonZeroValues(def.Constraints.Pattern, def.Constraints.Maximum, def.Constraints.Minimum,
			def.Constraints.ExclusiveMaximum, def.Constraints.ExclusiveMinimum, def.Constraints.Nullable,
			def.Constraints.MultipleOf, def.Constraints.MaxLength, def.Constraints.MinLength,
			def.Constraints.MaxItems, def.Constraints.MinItems, def.Constraints.UniqueItems,
			def.Constraints.MaxProperties, def.Constraints.MinProperties) || len(def.Constraints.Additional) > 0 {
			cw.writeLine(indent+1, "Constraints: "+cw.opts.alias()+typeConstraints+"{", false)
			writeZeroField(cw, indent+2, "Pattern", def.Constraints.Pattern)
			writeZeroField(cw, indent+2, "Maximum", def.Constraints.Maximum)
			writeZeroField(cw, indent+2, "Minimum", def.Constraints.Minimum)
			writeZeroField(cw, indent+2, "ExclusiveMaximum", def.Constraints.ExclusiveMaximum)
			writeZeroField(cw, indent+2, "ExclusiveMinimum", def.Constraints.ExclusiveMinimum)
			writeZeroField(cw, indent+2, "Nullable", def.Constraints.Nullable)
			writeZeroField(cw, indent+2, "MultipleOf", def.Constraints.MultipleOf)
			writeZeroField(cw, indent+2, "MaxLength", def.Constraints.MaxLength)
			writeZeroField(cw, indent+2, "MinLength", def.Constraints.MinLength)
			writeZeroField(cw, indent+2, "MaxItems", def.Constraints.MaxItems)
			writeZeroField(cw, indent+2, "MinItems", def.Constraints.MinItems)
			writeZeroField(cw, indent+2, "UniqueItems", def.Constraints.UniqueItems)
			writeZeroField(cw, indent+2, "MaxProperties", def.Constraints.MaxProperties)
			writeZeroField(cw, indent+2, "MinProperties", def.Constraints.MinProperties)
			if len(def.Constraints.Additional) > 0 {
				cw.writeStart(indent+2, "Additional: ")
				cw.writeValue(indent+2, def.Constraints.Additional)
			}
			cw.writeEnd(indent+1, "},")
		}
		cw.writeExtensions(indent+1, def.Extensions)
		writeZeroField(cw, indent+1, "Comment", def.Comment)
	}
}

func generateSecurityScheme(indent int, def chioas.SecurityScheme, cw *codeWriter) {
	writeZeroField(cw, indent+1, "Name", def.Name)
	writeZeroField(cw, indent+1, "Description", def.Description)
	writeZeroField(cw, indent+1, "Type", def.Type)
	writeZeroField(cw, indent+1, "Scheme", def.Scheme)
	writeZeroField(cw, indent+1, "ParamName", def.ParamName)
	writeZeroField(cw, indent+1, "In", def.In)
	if len(def.Scopes) > 0 {
		cw.writeStart(indent+1, "Scopes: ")
		cw.writeValue(indent+1, def.Scopes)
	}
	cw.writeExtensions(indent+1, def.Extensions)
	writeZeroField(cw, indent+1, "Comment", def.Comment)
}

func compareMethods(ma, mb string) bool {
	a := slices.Index(chioas.MethodsOrder, ma)
	b := slices.Index(chioas.MethodsOrder, mb)
	if a == -1 && b == -1 {
		return ma < mb
	} else if a == -1 {
		return false
	} else if b == -1 {
		return true
	}
	return a < b
}

const (
	typeDefinition       = "Definition"
	typePath             = "Path"
	typePaths            = "Paths"
	typeMethod           = "Method"
	typeMethods          = "Methods"
	typeComponents       = "Components"
	typeRequest          = "Request"
	typeResponses        = "Responses"
	typeResponse         = "Response"
	typeExtensions       = "Extensions"
	typeProperties       = "Properties"
	typeSchema           = "Schema"
	typeSchemas          = "Schemas"
	typeSecurityScheme   = "SecurityScheme"
	typeSecuritySchemes  = "SecuritySchemes"
	typePathParams       = "PathParams"
	typeQueryParams      = "QueryParams"
	typeConstraints      = "Constraints"
	typeDiscriminator    = "Discriminator"
	typeOfs              = "Ofs"
	typeOfSchema         = "OfSchema"
	typeOf               = "Of"
	typeContentTypes     = "ContentTypes"
	typeExample          = "Example"
	typeExamples         = "Examples"
	typeInfo             = "Info"
	typeContact          = "Contact"
	typeLicense          = "License"
	typeExternalDocs     = "ExternalDocs"
	typeServers          = "Servers"
	typeTags             = "Tags"
	typeCommonRequests   = "CommonRequests"
	typeCommonResponses  = "CommonResponses"
	typeCommonParameter  = "CommonParameter"
	typeCommonParameters = "CommonParameters"
)
