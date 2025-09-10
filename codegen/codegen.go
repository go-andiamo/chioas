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
//   - GenerateCode does not run go/format. Callers should format the result themselves
//   - If Options.SkipPrologue is false, a package clause and a single chioas import (optionally
//     aliased) are emitted. When true, the caller is responsible for imports.
//
// Errors:
//   - Returns the first write error encountered. It does not close w.
func GenerateCode[T ItemType](item T, w io.Writer, opts Options) error {
	writer := newCodeWriter(w, opts)
	writer.writePrologue()
	switch it := any(item).(type) {
	case chioas.Definition:
		generateDefinition(it, false, writer)
	case *chioas.Definition:
		generateDefinition(*it, true, writer)
	case chioas.Path:
		writer.writeVarStart(writer.opts.topVarName(), typePath, false)
		generatePath(0, it, writer)
		writer.writeLine(0, "}", true)
	case *chioas.Path:
		writer.writeVarStart(writer.opts.topVarName(), typePath, true)
		generatePath(0, *it, writer)
		writer.writeLine(0, "}", true)
	case chioas.Method:
		writer.writeVarStart(writer.opts.topVarName(), typeMethod, false)
		generateMethodInner(0, it, writer)
		writer.writeLine(0, "}", true)
	case *chioas.Method:
		writer.writeVarStart(writer.opts.topVarName(), typeMethod, true)
		generateMethodInner(0, *it, writer)
		writer.writeLine(0, "}", true)
	case chioas.Schema:
		writer.writeVarStart(writer.opts.topVarName(), typeSchema, false)
		generateSchema(0, &it, writer)
		writer.writeLine(0, "}", true)
	case *chioas.Schema:
		writer.writeVarStart(writer.opts.topVarName(), typeSchema, true)
		generateSchema(0, it, writer)
		writer.writeLine(0, "}", true)
	case chioas.Components:
		if !opts.HoistComponents {
			writer.writeVarStart(writer.opts.topVarName(), typeComponents, false)
			generateComponentsInner(0, &it, writer)
			writer.writeLine(0, "}", true)
		} else {
			generateComponentsVars(&it, writer, true, false)
		}
	case *chioas.Components:
		if !opts.HoistComponents {
			writer.writeVarStart(writer.opts.topVarName(), typeComponents, true)
			generateComponentsInner(0, it, writer)
			writer.writeLine(0, "}", true)
		} else {
			generateComponentsVars(it, writer, true, true)
		}
	}
	return writer.format()
}

func generateDefinition(def chioas.Definition, ptr bool, w *codeWriter) {
	w.writeVarStart(w.opts.topVarName(), typeDefinition, ptr)
	generateInfo(1, def.Info, w)
	if len(def.Servers) > 0 {
		w.writeLine(1, "Servers: "+w.opts.alias()+typeServers+"{", false)
		ks := maps.Keys(def.Servers)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(2, k)
			s := def.Servers[k]
			writeZeroField(w, 3, "Description", s.Description)
			w.writeExtensions(3, s.Extensions)
			writeZeroField(w, 3, "Comment", s.Comment)
			w.writeEnd(2, "},")
		}
		w.writeEnd(1, "},")
	}
	if len(def.Tags) > 0 {
		w.writeLine(1, "Tags: "+w.opts.alias()+typeTags+"{", false)
		for _, tg := range def.Tags {
			w.writeLine(2, "{", false)
			writeZeroField(w, 3, "Name", tg.Name)
			writeZeroField(w, 3, "Description", tg.Description)
			if tg.ExternalDocs != nil {
				w.writeLine(3, "ExternalDocs: &"+w.opts.alias()+typeExternalDocs+"{", false)
				writeZeroField(w, 4, "Description", tg.ExternalDocs.Description)
				writeZeroField(w, 4, "Url", tg.ExternalDocs.Url)
				w.writeExtensions(4, tg.ExternalDocs.Extensions)
				writeZeroField(w, 4, "Comment", tg.ExternalDocs.Comment)
				w.writeEnd(3, "},")
			}
			w.writeExtensions(3, tg.Extensions)
			writeZeroField(w, 3, "Comment", tg.Comment)
			w.writeEnd(2, "},")
		}
		w.writeEnd(1, "},")
	}
	generateMethods(1, def.Methods, w)
	var paths []string
	if !w.opts.HoistPaths {
		generatePaths(1, def.Paths, w)
	} else {
		w.writeCollectionFieldStart(1, typePaths, typePaths)
		paths = maps.Keys(def.Paths)
		sort.Strings(paths)
		deduper := newNameDeDuper()
		for _, k := range paths {
			useName := k
			if useName == "" || useName == "/" {
				useName = "Root"
			}
			w.writeLine(2, strconv.Quote(k)+": "+deduper.take(w.opts.varName("Path", useName))+",", false)
		}
		w.writeLine(1, "},", false)
	}
	if len(def.Security) > 0 {
		w.writeLine(1, "Security: "+w.opts.alias()+typeSecuritySchemes+"{", false)
		for _, ss := range def.Security {
			w.writeLine(2, "{", false)
			generateSecurityScheme(2, ss, w)
			w.writeEnd(2, "},")
		}
		w.writeEnd(1, "},")
	}
	if def.Components != nil {
		generateComponents(1, def.Components, w)
	}
	w.writeExtensions(1, def.Extensions)
	writeZeroField(w, 1, "Comment", def.Comment)
	writeZeroField(w, 1, "AutoHeadMethods", def.AutoHeadMethods)
	writeZeroField(w, 1, "AutoOptionsMethods", def.AutoOptionsMethods)
	writeZeroField(w, 1, "RootAutoOptionsMethod", def.RootAutoOptionsMethod)
	writeZeroField(w, 1, "AutoMethodNotAllowed", def.AutoMethodNotAllowed)
	w.writeLine(0, "}", true)
	if w.opts.HoistPaths {
		w.writeLine(0, "var (", false)
		// write out the actual path vars...
		deduper := newNameDeDuper()
		for _, k := range paths {
			useName := k
			if useName == "" || useName == "/" {
				useName = "Root"
			}
			w.writeLine(1, deduper.take(w.opts.varName("Path", useName))+" = "+w.opts.alias()+typePath+"{", false)
			generatePath(1, def.Paths[k], w)
			w.writeLine(1, "}", false)
		}
		w.writeLine(0, ")", true)
	}
	if w.opts.HoistComponents && def.Components != nil {
		generateComponentsVars(def.Components, w, false, false)
	}
}

func generateComponentsVars(def *chioas.Components, w *codeWriter, topVar bool, topPtr bool) {
	deduper := newNameDeDuper()
	deDupe := func(kind, name string) string {
		return deduper.take(w.opts.varName(kind, name))
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
	w.writeLine(0, "var (", false)
	// write components var...
	if topVar {
		amp := ""
		if topPtr {
			amp = "&"
		}
		w.writeLine(1, w.opts.topVarName()+" = "+amp+w.opts.alias()+typeComponents+"{", false)
	} else {
		w.writeLine(1, w.opts.varName("", "components")+" = &"+w.opts.alias()+typeComponents+"{", false)
	}
	var schemas chioas.Schemas
	if len(def.Schemas) > 0 {
		w.writeCollectionFieldStart(2, typeSchemas, typeSchemas)
		schemas = append(schemas, def.Schemas...)
		slices.SortStableFunc(schemas, func(a, b chioas.Schema) bool {
			return a.Name < b.Name
		})
		for _, s := range schemas {
			w.writeLine(3, deDupe(varSchema, s.Name)+",", false)
		}
		w.writeEnd(2, "},")
	}
	var requests []string
	if len(def.Requests) > 0 {
		w.writeCollectionFieldStart(2, "Requests", typeCommonRequests)
		requests = maps.Keys(def.Requests)
		sort.Strings(requests)
		for _, k := range requests {
			w.writeLine(3, strconv.Quote(k)+": "+deDupe(varRequest, k)+",", false)
		}
		w.writeEnd(2, "},")
	}
	var responses []string
	if len(def.Responses) > 0 {
		w.writeCollectionFieldStart(2, "Responses", typeCommonResponses)
		responses = maps.Keys(def.Responses)
		sort.Strings(responses)
		for _, k := range responses {
			w.writeLine(3, strconv.Quote(k)+": "+deDupe(varResponse, k)+",", false)
		}
		w.writeEnd(2, "},")
	}
	var examples chioas.Examples
	if len(def.Examples) > 0 {
		w.writeCollectionFieldStart(2, typeExamples, typeExamples)
		examples = append(examples, def.Examples...)
		slices.SortStableFunc(examples, func(a, b chioas.Example) bool {
			return a.Name < b.Name
		})
		for _, eg := range examples {
			w.writeLine(3, deDupe(varExample, eg.Name)+",", false)
		}
		w.writeEnd(2, "},")
	}
	var params []string
	if len(def.Parameters) > 0 {
		w.writeCollectionFieldStart(2, "Parameters", typeCommonParameters)
		params = maps.Keys(def.Parameters)
		sort.Strings(params)
		for _, k := range params {
			w.writeLine(3, strconv.Quote(k)+": "+deDupe(varParameter, k)+",", false)
		}
		w.writeEnd(2, "},")
	}
	var secSchemes chioas.SecuritySchemes
	if len(def.SecuritySchemes) > 0 {
		w.writeCollectionFieldStart(2, typeSecuritySchemes, typeSecuritySchemes)
		secSchemes = append(secSchemes, def.SecuritySchemes...)
		slices.SortStableFunc(secSchemes, func(a, b chioas.SecurityScheme) bool {
			return a.Name < b.Name
		})
		for _, s := range secSchemes {
			w.writeLine(3, deDupe(varSecurityScheme, s.Name)+",", false)
		}
		w.writeEnd(2, "},")
	}
	w.writeExtensions(2, def.Extensions)
	writeZeroField(w, 2, "Comment", def.Comment)
	w.writeEnd(1, "}")
	// write components parts vars...
	deduper.clear()
	// schemas...
	for _, s := range schemas {
		w.writeLine(1, deDupe(varSchema, s.Name)+" = "+w.opts.alias()+typeSchema+"{", false)
		generateSchema(1, &s, w)
		w.writeEnd(1, "}")
	}
	// requests...
	for _, k := range requests {
		w.writeLine(1, deDupe(varRequest, k)+" = "+w.opts.alias()+typeRequest+"{", false)
		r := def.Requests[k]
		generateRequest(1, &r, w)
		w.writeEnd(1, "}")
	}
	// responses...
	for _, k := range responses {
		w.writeLine(1, deDupe(varResponse, k)+" = "+w.opts.alias()+typeResponse+"{", false)
		generateResponse(1, def.Responses[k], w)
		w.writeEnd(1, "}")
	}
	// examples...
	for _, eg := range examples {
		w.writeLine(1, deDupe(varExample, eg.Name)+" = "+w.opts.alias()+typeExample+"{", false)
		generateExample(1, eg, w)
		w.writeEnd(1, "}")
	}
	// parameters...
	for _, k := range params {
		w.writeLine(1, deDupe(varParameter, k)+" = "+w.opts.alias()+typeCommonParameter+"{", false)
		generateCommonParam(1, def.Parameters[k], w)
		w.writeEnd(1, "}")
	}
	// security schemes...
	for _, s := range secSchemes {
		w.writeLine(1, deDupe(varSecurityScheme, s.Name)+" = "+w.opts.alias()+typeSecurityScheme+"{", false)
		generateSecurityScheme(1, s, w)
		w.writeEnd(1, "}")
	}
	// end vars...
	w.writeLine(0, ")", true)
}

func generateComponents(indent int, def *chioas.Components, w *codeWriter) {
	if w.opts.HoistComponents {
		w.writeLine(indent, "Components: "+w.opts.varName("", "Components")+",", false)
		return
	}
	w.writeLine(indent, "Components: &"+w.opts.alias()+typeComponents+"{", false)
	generateComponentsInner(indent, def, w)
	w.writeEnd(indent, "},")
}

func generateComponentsInner(indent int, def *chioas.Components, w *codeWriter) {
	if len(def.Schemas) > 0 {
		w.writeCollectionFieldStart(indent+1, typeSchemas, typeSchemas)
		ss := append(chioas.Schemas{}, def.Schemas...)
		slices.SortStableFunc(ss, func(a, b chioas.Schema) bool {
			return a.Name < b.Name
		})
		for _, s := range ss {
			w.writeLine(indent+2, "{", false)
			generateSchema(indent+2, &s, w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	if len(def.Requests) > 0 {
		w.writeCollectionFieldStart(indent+1, "Requests", typeCommonRequests)
		ks := maps.Keys(def.Requests)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(indent+2, k)
			r := def.Requests[k]
			generateRequest(indent+2, &r, w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	if len(def.Responses) > 0 {
		w.writeCollectionFieldStart(indent+1, "Responses", typeCommonResponses)
		ks := maps.Keys(def.Responses)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(indent+2, k)
			generateResponse(indent+2, def.Responses[k], w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	if len(def.Examples) > 0 {
		w.writeCollectionFieldStart(indent+1, typeExamples, typeExamples)
		egs := append(chioas.Examples{}, def.Examples...)
		slices.SortStableFunc(egs, func(a, b chioas.Example) bool {
			return a.Name < b.Name
		})
		for _, eg := range egs {
			w.writeLine(indent+2, "{", false)
			generateExample(indent+2, eg, w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	if len(def.Parameters) > 0 {
		w.writeCollectionFieldStart(indent+1, "Parameters", typeCommonParameters)
		ks := maps.Keys(def.Parameters)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(indent+2, k)
			generateCommonParam(indent+2, def.Parameters[k], w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	if len(def.SecuritySchemes) > 0 {
		w.writeCollectionFieldStart(indent+1, typeSecuritySchemes, typeSecuritySchemes)
		ss := append(chioas.SecuritySchemes{}, def.SecuritySchemes...)
		slices.SortStableFunc(ss, func(a, b chioas.SecurityScheme) bool {
			return a.Name < b.Name
		})
		for _, s := range ss {
			w.writeLine(indent+2, "{", false)
			generateSecurityScheme(indent+3, s, w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	w.writeExtensions(indent+2, def.Extensions)
	writeZeroField(w, indent+2, "Comment", def.Comment)
}

func generateInfo(indent int, def chioas.Info, w *codeWriter) {
	if hasNonZeroValues(def.Title, def.Description, def.Version, def.TermsOfService, def.Comment,
		def.Contact, def.License, def.Extensions, def.ExternalDocs) {
		w.writeLine(indent+1, "Info: "+w.opts.alias()+typeInfo+"{", false)
		writeZeroField(w, indent+2, "Title", def.Title)
		writeZeroField(w, indent+2, "Description", def.Description)
		writeZeroField(w, indent+2, "Version", def.Version)
		writeZeroField(w, indent+2, "TermsOfService", def.TermsOfService)
		if def.Contact != nil {
			w.writeLine(indent+2, "Contact: &"+w.opts.alias()+typeContact+"{", false)
			writeZeroField(w, indent+3, "Name", def.Contact.Name)
			writeZeroField(w, indent+3, "Url", def.Contact.Url)
			writeZeroField(w, indent+3, "Email", def.Contact.Email)
			w.writeExtensions(indent+3, def.Contact.Extensions)
			writeZeroField(w, indent+3, "Comment", def.Contact.Comment)
			w.writeEnd(indent+2, "},")
		}
		if def.License != nil {
			w.writeLine(indent+2, "License: &"+w.opts.alias()+typeLicense+"{", false)
			writeZeroField(w, indent+3, "Name", def.License.Name)
			writeZeroField(w, indent+3, "Url", def.License.Url)
			w.writeExtensions(indent+3, def.License.Extensions)
			writeZeroField(w, indent+3, "Comment", def.License.Comment)
			w.writeEnd(indent+2, "},")
		}
		w.writeExtensions(indent+2, def.Extensions)
		if def.ExternalDocs != nil {
			w.writeLine(indent+2, "ExternalDocs: &"+w.opts.alias()+typeExternalDocs+"{", false)
			writeZeroField(w, indent+3, "Description", def.ExternalDocs.Description)
			writeZeroField(w, indent+3, "Url", def.ExternalDocs.Url)
			w.writeExtensions(indent+3, def.ExternalDocs.Extensions)
			writeZeroField(w, indent+3, "Comment", def.ExternalDocs.Comment)
			w.writeEnd(indent+2, "},")
		}
		writeZeroField(w, indent+2, "Comment", def.Comment)
		w.writeEnd(indent+1, "},")
	}
}

func generatePath(indent int, def chioas.Path, w *codeWriter) {
	generateMethods(indent+1, def.Methods, w)
	generatePaths(indent+1, def.Paths, w)
	writeZeroField(w, indent+1, "Tag", def.Tag)
	if len(def.PathParams) > 0 {
		w.writeLine(indent+1, "PathParams: "+w.opts.alias()+typePathParams+"{", false)
		ks := maps.Keys(def.PathParams)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(indent+2, k)
			generatePathParam(indent+2, def.PathParams[k], w)
			w.writeLine(indent+2, "},", false)
		}
		w.writeEnd(indent+1, "},")
	}
	writeZeroField(w, indent+1, "HideDocs", def.HideDocs)
	w.writeExtensions(indent+1, def.Extensions)
	writeZeroField(w, indent+1, "Comment", def.Comment)
	writeZeroField(w, indent+1, "AutoOptionsMethod", def.AutoOptionsMethod)
}

func generatePaths(indent int, paths chioas.Paths, w *codeWriter) {
	if len(paths) > 0 {
		w.writeCollectionFieldStart(indent, typePaths, typePaths)
		ks := maps.Keys(paths)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(indent+1, k)
			generatePath(indent+1, paths[k], w)
			w.writeLine(indent+1, "},", false)
		}
		w.writeLine(indent, "},", false)
	}
}

func generateMethods(indent int, methods chioas.Methods, w *codeWriter) {
	if len(methods) > 0 {
		w.writeCollectionFieldStart(indent, typeMethods, typeMethods)
		sms := maps.Keys(methods)
		sort.Slice(sms, func(i, j int) bool {
			return compareMethods(sms[i], sms[j])
		})
		for _, m := range sms {
			generateMethod(indent+1, m, methods[m], w)
		}
		w.writeLine(indent, "},", false)
	}
}

func generateMethod(indent int, method string, def chioas.Method, w *codeWriter) {
	w.writeLine(indent, w.opts.translateMethod(method)+": {", false)
	generateMethodInner(indent, def, w)
	w.writeLine(indent, "},", false)
}

func generateMethodInner(indent int, def chioas.Method, w *codeWriter) {
	writeZeroField(w, indent+1, "Description", def.Description)
	writeZeroField(w, indent+1, "Summary", def.Summary)
	if hs, ok := def.Handler.(string); ok {
		// only output handler if a string
		writeZeroField(w, indent+1, "Handler", hs)
	}
	writeZeroField(w, indent+1, "OperationId", def.OperationId)
	writeZeroField(w, indent+1, "Tag", def.Tag)
	if len(def.QueryParams) > 0 {
		w.writeLine(indent+1, "QueryParams: "+w.opts.alias()+typeQueryParams+"{", false)
		for _, qp := range def.QueryParams {
			w.writeLine(indent+2, "{", false)
			generateQueryParam(indent+2, qp, w)
			w.writeLine(indent+2, "},", false)
		}
		w.writeEnd(indent+1, "},")
	}
	if def.Request != nil {
		w.writeLine(indent+1, "Request: &"+w.opts.alias()+typeRequest+"{", false)
		generateRequest(indent+1, def.Request, w)
		w.writeEnd(indent+1, "},")
	}
	if len(def.Responses) > 0 {
		w.writeLine(indent+1, "Responses: "+w.opts.alias()+typeResponses+"{", false)
		ks := maps.Keys(def.Responses)
		slices.Sort(ks)
		for _, k := range ks {
			w.writeLine(indent+2, w.opts.translateStatus(k)+": {", false)
			generateResponse(indent+2, def.Responses[k], w)
			w.writeEnd(indent+2, "},")
		}
		w.writeEnd(indent+1, "},")
	}
	writeZeroField(w, indent+1, "Deprecated", def.Deprecated)
	if len(def.Security) > 0 {
		w.writeLine(indent+1, "Security: "+w.opts.alias()+typeSecuritySchemes+"{", false)
		for _, s := range def.Security {
			w.writeLine(indent+2, "{", false)
			w.writeLine(indent+3, "Name: "+strconv.Quote(s.Name)+",", false)
			w.writeLine(indent+2, "},", false)
		}
		w.writeEnd(indent+1, "},")
	}
	writeZeroField(w, indent+1, "OptionalSecurity", def.OptionalSecurity)
	w.writeExtensions(indent+1, def.Extensions)
	writeZeroField(w, indent+1, "Comment", def.Comment)
	writeZeroField(w, indent+1, "HideDocs", def.HideDocs)
}

func generateRequest(indent int, def *chioas.Request, w *codeWriter) {
	if def.Ref != "" {
		w.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.RequestBodies, def.Ref))+",", false)
	} else {
		writeZeroField(w, indent+1, "Description", def.Description)
		writeZeroField(w, indent+1, "Required", def.Required)
		writeZeroField(w, indent+1, "ContentType", def.ContentType)
		generateAlternativeContentTypes(indent, def.AlternativeContentTypes, w)
		if s := def.Schema; s != nil {
			generateVaryingSchema(indent, def.Schema, w)
		} else if def.SchemaRef != "" {
			w.writeSchemaRef(indent+1, def.SchemaRef)
		}
		writeZeroField(w, indent+1, "IsArray", def.IsArray)
		generateExamples(indent, def.Examples, w)
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generateResponse(indent int, def chioas.Response, w *codeWriter) {
	if def.Ref != "" {
		w.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.Responses, def.Ref))+",", false)
	} else {
		writeZeroField(w, indent+1, "Description", def.Description)
		writeZeroField(w, indent+1, "NoContent", def.NoContent)
		writeZeroField(w, indent+1, "ContentType", def.ContentType)
		generateAlternativeContentTypes(indent, def.AlternativeContentTypes, w)
		if s := def.Schema; s != nil {
			generateVaryingSchema(indent, def.Schema, w)
		} else if def.SchemaRef != "" {
			w.writeSchemaRef(indent+1, def.SchemaRef)
		}
		writeZeroField(w, indent+1, "IsArray", def.IsArray)
		generateExamples(indent, def.Examples, w)
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generateAlternativeContentTypes(indent int, cts chioas.ContentTypes, w *codeWriter) {
	if len(cts) > 0 {
		w.writeLine(indent+1, "AlternativeContentTypes: "+w.opts.alias()+typeContentTypes+"{", false)
		ks := maps.Keys(cts)
		sort.Strings(ks)
		for _, k := range ks {
			w.writeKey(indent+2, k)
			generateContentType(indent+2, cts[k], w)
			w.writeLine(indent+2, "},", false)
		}
		w.writeEnd(indent+1, "},")
	}
}

func generateContentType(indent int, def chioas.ContentType, w *codeWriter) {
	if s := def.Schema; s != nil {
		generateVaryingSchema(indent, def.Schema, w)
	} else if def.SchemaRef != "" {
		w.writeSchemaRef(indent+1, def.SchemaRef)
	}
	writeZeroField(w, indent+1, "IsArray", def.IsArray)
	generateExamples(indent, def.Examples, w)
	w.writeExtensions(indent+1, def.Extensions)
	writeZeroField(w, indent+1, "Comment", def.Comment)
}

func generateVaryingSchema(indent int, s any, w *codeWriter) {
	switch schema := s.(type) {
	case chioas.Schema:
		w.writeLine(indent+1, "Schema: "+w.opts.alias()+typeSchema+"{", false)
		generateSchema(indent+1, &schema, w)
		w.writeLine(indent+1, "},", false)
	case *chioas.Schema:
		w.writeLine(indent+1, "Schema: &"+w.opts.alias()+typeSchema+"{", false)
		generateSchema(indent+1, schema, w)
		w.writeLine(indent+1, "},", false)
	case chioas.SchemaConverter:
		w.writeLine(indent+1, "Schema: &"+w.opts.alias()+typeSchema+"{", false)
		ts := schema.ToSchema()
		if ts != nil {
			generateSchema(indent+1, ts, w)
		}
		w.writeLine(indent+1, "},", false)
	default:
		if fs, err := chioas.SchemaFrom(s); err == nil {
			w.writeLine(indent+1, "Schema: &"+w.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, fs, w)
			w.writeLine(indent+1, "},", false)
		} else {
			// note: this will produce bad code - but indicates needs manual attention
			w.writeLine(indent+1, fmt.Sprintf("Schema: %T,", s), false)
		}
	}
}

func generateExamples(indent int, egs chioas.Examples, w *codeWriter) {
	if len(egs) > 0 {
		w.writeLine(indent+1, "Examples: "+w.opts.alias()+typeExamples+"{", false)
		for _, eg := range egs {
			w.writeLine(indent+2, "{", false)
			generateExample(indent+2, eg, w)
			w.writeLine(indent+2, "},", false)
		}
		w.writeEnd(indent+1, "},")
	}
}

func generateExample(indent int, def chioas.Example, w *codeWriter) {
	if def.ExampleRef != "" {
		w.writeLine(indent+1, "ExampleRef: "+strconv.Quote(refs.Normalize(tags.Examples, def.ExampleRef))+",", false)
	} else {
		writeZeroField(w, indent+1, "Name", def.Name)
		writeZeroField(w, indent+1, "Summary", def.Summary)
		writeZeroField(w, indent+1, "Description", def.Description)
		if def.Value != nil {
			w.writeStart(indent+1, "Value: ")
			w.writeValue(indent+1, def.Value)
		}
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generatePathParam(indent int, def chioas.PathParam, w *codeWriter) {
	if def.Ref != "" {
		w.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.Parameters, def.Ref))+",", false)
	} else {
		writeZeroField(w, indent+1, "Description", def.Description)
		if def.Example != nil {
			w.writeStart(indent+1, "Example: ")
			w.writeValue(indent+1, def.Example)
		}
		if def.Schema != nil {
			w.writeLine(indent+1, "Schema: &"+w.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, def.Schema, w)
			w.writeEnd(indent+1, "},")
		} else if def.SchemaRef != "" {
			w.writeSchemaRef(indent+1, def.SchemaRef)
		}
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generateQueryParam(indent int, def chioas.QueryParam, w *codeWriter) {
	if def.Ref != "" {
		w.writeLine(indent+1, "Ref: "+strconv.Quote(refs.Normalize(tags.Parameters, def.Ref))+",", false)
	} else {
		writeZeroField(w, indent+1, "Name", def.Name)
		writeZeroField(w, indent+1, "Description", def.Description)
		writeZeroField(w, indent+1, "Required", def.Required)
		writeZeroField(w, indent+1, "In", def.In)
		if def.Example != nil {
			w.writeStart(indent+1, "Example: ")
			w.writeValue(indent+1, def.Example)
		}
		if def.Schema != nil {
			w.writeLine(indent+1, "Schema: &"+w.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+1, def.Schema, w)
			w.writeEnd(indent+1, "},")
		} else if def.SchemaRef != "" {
			w.writeSchemaRef(indent+1, def.SchemaRef)
		}
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generateCommonParam(indent int, def chioas.CommonParameter, w *codeWriter) {
	writeZeroField(w, indent+1, "Name", def.Name)
	writeZeroField(w, indent+1, "Description", def.Description)
	writeZeroField(w, indent+1, "Required", def.Required)
	writeZeroField(w, indent+1, "In", def.In)
	if def.Example != nil {
		w.writeStart(indent+1, "Example: ")
		w.writeValue(indent+1, def.Example)
	}
	if def.Schema != nil {
		w.writeLine(indent+1, "Schema: &"+w.opts.alias()+typeSchema+"{", false)
		generateSchema(indent+1, def.Schema, w)
		w.writeEnd(indent+1, "},")
	} else if def.SchemaRef != "" {
		w.writeSchemaRef(indent+1, def.SchemaRef)
	}
	w.writeExtensions(indent+1, def.Extensions)
	writeZeroField(w, indent+1, "Comment", def.Comment)
}

func generateSchema(indent int, def *chioas.Schema, w *codeWriter) {
	if def.SchemaRef != "" {
		w.writeSchemaRef(indent+1, def.SchemaRef)
	} else {
		writeZeroField(w, indent+1, "Name", def.Name)
		writeZeroField(w, indent+1, "Description", def.Description)
		writeZeroField(w, indent+1, "Type", def.Type)
		writeZeroField(w, indent+1, "Format", def.Format)
		if len(def.RequiredProperties) > 0 {
			w.writeStart(indent+1, "RequiredProperties: ")
			w.writeValue(indent+1, def.RequiredProperties)
		}
		if len(def.Properties) > 0 {
			w.writeLine(indent+1, "Properties: "+w.opts.alias()+typeProperties+"{", false)
			for _, p := range def.Properties {
				w.writeLine(indent+2, "{", false)
				generateProperty(indent+2, p, w)
				w.writeLine(indent+2, "},", false)
			}
			w.writeEnd(indent+1, "},")
		}
		if def.Default != nil {
			w.writeStart(indent+1, "Default: ")
			w.writeValue(indent+1, def.Default)
		}
		if def.Example != nil {
			w.writeStart(indent+1, "Example: ")
			w.writeValue(indent+1, def.Example)
		}
		if len(def.Enum) > 0 {
			w.writeStart(indent+1, "Enum: ")
			w.writeValue(indent+1, def.Enum)
		}
		if def.Discriminator != nil {
			w.writeLine(indent+1, "Discriminator: &"+w.opts.alias()+typeDiscriminator+"{", false)
			generateDiscriminator(indent+1, def.Discriminator, w)
			w.writeEnd(indent+1, "},")
		}
		if def.Ofs != nil {
			w.writeLine(indent+1, "Ofs: &"+w.opts.alias()+typeOfs+"{", false)
			generateOfs(indent+1, def.Ofs, w)
			w.writeEnd(indent+1, "},")
		}
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generateDiscriminator(indent int, def *chioas.Discriminator, w *codeWriter) {
	writeZeroField(w, indent+1, "PropertyName", def.PropertyName)
	w.writeLine(indent+1, "Mapping: map[string]string{", false)
	ks := maps.Keys(def.Mapping)
	sort.Strings(ks)
	for _, k := range ks {
		w.writeLine(indent+2, strconv.Quote(k)+": "+strconv.Quote(refs.Normalize(tags.Schemas, def.Mapping[k]))+",", false)
	}
	w.writeEnd(indent+1, "},")
	w.writeExtensions(indent+1, def.Extensions)
	writeZeroField(w, indent+1, "Comment", def.Comment)
}

func generateOfs(indent int, def *chioas.Ofs, w *codeWriter) {
	switch def.OfType {
	case chioas.AllOf:
		w.writeLine(indent+1, "OfType: "+w.opts.alias()+"AllOf,", false)
	case chioas.AnyOf:
		w.writeLine(indent+1, "OfType: "+w.opts.alias()+"AnyOf,", false)
	case chioas.OneOf:
		w.writeLine(indent+1, "OfType: "+w.opts.alias()+"OneOf,", false)
	default:
		w.writeLine(indent+1, "OfType: "+strconv.Itoa(int(def.OfType))+",", false)
	}
	w.writeLine(indent+1, "Of: []"+w.opts.alias()+typeOfSchema+"{", false)
	for _, ofs := range def.Of {
		w.writeLine(indent+2, "&"+w.opts.alias()+typeOf+"{", false)
		if ofs.IsRef() {
			w.writeSchemaRef(indent+3, ofs.Ref())
		} else {
			s := ofs.Schema()
			if s == nil {
				s = &chioas.Schema{}
			}
			w.writeLine(indent+3, "SchemaDef: &"+w.opts.alias()+typeSchema+"{", false)
			generateSchema(indent+3, s, w)
			w.writeEnd(indent+3, "},")
		}
		w.writeEnd(indent+2, "},")
	}
	w.writeEnd(indent+1, "},")
}

func generateProperty(indent int, def chioas.Property, w *codeWriter) {
	if def.SchemaRef != "" {
		w.writeSchemaRef(indent+1, def.SchemaRef)
	} else {
		writeZeroField(w, indent+1, "Name", def.Name)
		writeZeroField(w, indent+1, "Description", def.Description)
		writeZeroField(w, indent+1, "Type", def.Type)
		writeZeroField(w, indent+1, "ItemType", def.ItemType)
		if len(def.Properties) > 0 {
			w.writeLine(indent+1, "Properties: "+w.opts.alias()+typeProperties+"{", false)
			for _, p := range def.Properties {
				w.writeLine(indent+2, "{", false)
				generateProperty(indent+2, p, w)
				w.writeLine(indent+2, "},", false)
			}
			w.writeEnd(indent+1, "},")
		}
		writeZeroField(w, indent+1, "Required", def.Required)
		writeZeroField(w, indent+1, "Format", def.Format)
		if def.Example != nil {
			w.writeStart(indent+1, "Example: ")
			w.writeValue(indent+1, def.Example)
		}
		if len(def.Enum) > 0 {
			w.writeStart(indent+1, "Enum: ")
			w.writeValue(indent+1, def.Enum)
		}
		writeZeroField(w, indent+1, "Deprecated", def.Deprecated)
		if hasNonZeroValues(def.Constraints.Pattern, def.Constraints.Maximum, def.Constraints.Minimum,
			def.Constraints.ExclusiveMaximum, def.Constraints.ExclusiveMinimum, def.Constraints.Nullable,
			def.Constraints.MultipleOf, def.Constraints.MaxLength, def.Constraints.MinLength,
			def.Constraints.MaxItems, def.Constraints.MinItems, def.Constraints.UniqueItems,
			def.Constraints.MaxProperties, def.Constraints.MinProperties) || len(def.Constraints.Additional) > 0 {
			w.writeLine(indent+1, "Constraints: "+w.opts.alias()+typeConstraints+"{", false)
			writeZeroField(w, indent+2, "Pattern", def.Constraints.Pattern)
			writeZeroField(w, indent+2, "Maximum", def.Constraints.Maximum)
			writeZeroField(w, indent+2, "Minimum", def.Constraints.Minimum)
			writeZeroField(w, indent+2, "ExclusiveMaximum", def.Constraints.ExclusiveMaximum)
			writeZeroField(w, indent+2, "ExclusiveMinimum", def.Constraints.ExclusiveMinimum)
			writeZeroField(w, indent+2, "Nullable", def.Constraints.Nullable)
			writeZeroField(w, indent+2, "MultipleOf", def.Constraints.MultipleOf)
			writeZeroField(w, indent+2, "MaxLength", def.Constraints.MaxLength)
			writeZeroField(w, indent+2, "MinLength", def.Constraints.MinLength)
			writeZeroField(w, indent+2, "MaxItems", def.Constraints.MaxItems)
			writeZeroField(w, indent+2, "MinItems", def.Constraints.MinItems)
			writeZeroField(w, indent+2, "UniqueItems", def.Constraints.UniqueItems)
			writeZeroField(w, indent+2, "MaxProperties", def.Constraints.MaxProperties)
			writeZeroField(w, indent+2, "MinProperties", def.Constraints.MinProperties)
			if len(def.Constraints.Additional) > 0 {
				w.writeStart(indent+2, "Additional: ")
				w.writeValue(indent+2, def.Constraints.Additional)
			}
			w.writeEnd(indent+1, "},")
		}
		w.writeExtensions(indent+1, def.Extensions)
		writeZeroField(w, indent+1, "Comment", def.Comment)
	}
}

func generateSecurityScheme(indent int, def chioas.SecurityScheme, w *codeWriter) {
	writeZeroField(w, indent+1, "Name", def.Name)
	writeZeroField(w, indent+1, "Description", def.Description)
	writeZeroField(w, indent+1, "Type", def.Type)
	writeZeroField(w, indent+1, "Scheme", def.Scheme)
	writeZeroField(w, indent+1, "ParamName", def.ParamName)
	writeZeroField(w, indent+1, "In", def.In)
	if len(def.Scopes) > 0 {
		w.writeStart(indent+1, "Scopes: ")
		w.writeValue(indent+1, def.Scopes)
	}
	w.writeExtensions(indent+1, def.Extensions)
	writeZeroField(w, indent+1, "Comment", def.Comment)
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
