package codegen

import (
	"bytes"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/go-andiamo/chioas/internal/values"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type SchemaStructOptions struct {
	Package       string // e.g. "api" (default "api")
	SkipPrologue  bool   // don't write package & imports
	PublicStructs bool   // public schema structs
	NoRequests    bool   // suppresses schema structs for method requests
	NoResponses   bool   // suppresses schema structs for method responses
	OASTags       bool   // if true, writes `oas:"..."` tags for fields
	GoDoc         bool   // if true, writes a godoc comment for each schema struct (indicating its usage)
	// Components is the components to use for $ref resolution
	//
	// Usage:
	//   - if generating schema structs (GenerateSchemaStructs) with chioas.Components, this field is not used
	//   - if generating schema structs (GenerateSchemaStructs) with chioas.Definition, this field (if non-nil) overrides the Definition.Components
	//   - for other types passed to GenerateSchemaStructs, this chioas.Components is used to resolve schema $ref
	//
	Components *chioas.Components
	// KeepComponentProperties if true (and Components supplied), is used to make
	// struct fields for properties that have an object schema - reference another generated struct
	KeepComponentProperties bool
	// Format if set, formats output in canonical gofmt style (and checks syntax)
	//
	// Note: using this option means the output will be buffered before writing to the final writer
	Format  bool
	UseCRLF bool // true to use \r\n as the line terminator
	/*	NoAlternates  bool   // suppresses schema structs for AlternativeContentTypes in method requests/responses */
}

type infoType int

const (
	infoTypeUnknown infoType = iota
	infoTypeRequest
	infoTypeResponse
)

type pathInfo struct {
	name         string
	path         string
	method       string
	infoType     infoType
	explicitName bool
}

func copyInfo(i *pathInfo, name, path, method string, t infoType) *pathInfo {
	r := &pathInfo{}
	if i != nil {
		r.name, r.path, r.method, r.infoType = i.name, i.path, i.method, i.infoType
	}
	if name != "" {
		r.name = name
	}
	if path != "" {
		r.path = path
	}
	if method != "" {
		r.method = method
	}
	if t != infoTypeUnknown {
		r.infoType = t
	}
	return r
}

func (pi *pathInfo) goDoc() []byte {
	var buf bytes.Buffer
	buf.Grow(20 + len(pi.path))
	switch pi.infoType {
	case infoTypeRequest:
		buf.WriteString("request")
	case infoTypeResponse:
		buf.WriteString("response")
	}
	if pi.method != "" {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(pi.method)
	}
	if pi.path != "" {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(pi.path)
	}
	return buf.Bytes()
}

type StructItemType interface {
	chioas.Definition | *chioas.Definition |
		chioas.Path | *chioas.Path |
		chioas.Method | *chioas.Method |
		chioas.Paths |
		chioas.Schema | *chioas.Schema |
		chioas.Components | *chioas.Components
}

func GenerateSchemaStructs[T StructItemType](item T, w io.Writer, opts SchemaStructOptions) error {
	sw := newStructsWriter(w, opts)
	sw.writePrologue()
	switch it := any(item).(type) {
	case chioas.Definition:
		sw.components = it.Components
		generateDefinitionStructs(it, sw)
	case *chioas.Definition:
		sw.components = it.Components
		generateDefinitionStructs(*it, sw)
	case chioas.Path:
		generatePathStructs(nil, it, sw)
	case *chioas.Path:
		generatePathStructs(nil, *it, sw)
	case chioas.Method:
		generateMethodStructs(nil, it, sw)
	case *chioas.Method:
		generateMethodStructs(nil, *it, sw)
	case chioas.Paths:
		generatePathsStructs(nil, it, sw)
	case chioas.Schema:
		generateSchemaStruct(nil, it, sw)
	case *chioas.Schema:
		generateSchemaStruct(nil, *it, sw)
	case chioas.Components:
		sw.components = &it
		sw.keep = true
		generateComponentsStructs(it, sw)
	case *chioas.Components:
		sw.components = it
		sw.keep = true
		generateComponentsStructs(*it, sw)
	}
	sw.generateIssuedSchemas()
	return sw.format()
}

func (w *structsWriter) generateIssuedSchemas() {
	if len(w.issuedSchemas) > 0 {
		names := make([]string, 0, len(w.issuedSchemas))
		nameToSchema := make(map[string]*chioas.Schema, len(w.issuedSchemas))
		nameToRef := make(map[string]string, len(w.issuedSchemas))
		// sort...
		for ref, name := range w.issuedSchemas {
			names = append(names, name)
			nameToSchema[name] = getComponentsSchema(w.components, ref)
			nameToRef[name] = ref
		}
		sort.Strings(names)
		// write...
		for _, name := range names {
			schema := nameToSchema[name]
			generateSchemaStruct(&pathInfo{
				name:         name,
				path:         refs.ComponentsPrefix + tags.Schemas + "/" + nameToRef[name],
				explicitName: true,
			}, *schema, w)
		}
	}
}

func getComponentsSchema(components *chioas.Components, ref string) (result *chioas.Schema) {
	for _, s := range components.Schemas {
		if s.Name == ref {
			result = &s
			break
		}
	}
	return result
}

func generateDefinitionStructs(def chioas.Definition, sw *structsWriter) {
	if sw.keep {
		// ensure that all components schemas are output
		if def.Components != nil {
			for _, s := range def.Components.Schemas {
				if s.SchemaRef == "" && s.Type == values.TypeObject {
					sw.issueSchema(s.Name)
				}
			}
		}
	}
	sms := maps.Keys(def.Methods)
	sort.Slice(sms, func(i, j int) bool {
		return compareMethods(sms[i], sms[j])
	})
	for _, m := range sms {
		method := def.Methods[m]
		mi := copyInfo(nil, "", "root", m, 0)
		generateMethodStructs(mi, method, sw)
	}
	generatePathsStructs(nil, def.Paths, sw)
}

func generatePathStructs(info *pathInfo, def chioas.Path, sw *structsWriter) {
	sms := maps.Keys(def.Methods)
	sort.Slice(sms, func(i, j int) bool {
		return compareMethods(sms[i], sms[j])
	})
	for _, m := range sms {
		method := def.Methods[m]
		mi := copyInfo(info, "", "", m, 0)
		generateMethodStructs(mi, method, sw)
	}
	generatePathsStructs(info, def.Paths, sw)
}

func generateMethodStructs(info *pathInfo, def chioas.Method, sw *structsWriter) {
	info = copyInfo(info, "", "", "", 0)
	if info.method == "" && info.path == "" {
		info.name = "Method"
	} else {
		info.name = defaultStubNaming.Name(info.path, info.method, def)
	}
	if !sw.opts.NoRequests && def.Request != nil {
		ir := copyInfo(info, toPascal(info.name+" Request"), "", "", infoTypeRequest)
		generateRequestStructs(ir, def.Request, sw)
	}
	if !sw.opts.NoResponses {
		ks := maps.Keys(def.Responses)
		sort.Ints(ks)
		for _, status := range ks {
			response := def.Responses[status]
			ir := copyInfo(info, toPascal(info.name+" "+statusCodeName(status)+" Response"), "", "", infoTypeResponse)
			generateResponseStructs(ir, &response, sw)
		}
	}
}

func generateRequestStructs(info *pathInfo, def *chioas.Request, sw *structsWriter) {
	if !sw.opts.NoRequests {
		if def, err := sw.resolveRequest(def, nil); err == nil {
			var schema *chioas.Schema
			if def.SchemaRef != "" {
				if s, aref, err := sw.resolveSchema(def.SchemaRef, nil); err == nil {
					schema = s
					if sw.keep {
						schema = nil
						sw.writeTypeAlias(info, sw.issueSchema(aref))
					}
				}
			} else {
				schema = getSchema(def.Schema)
			}
			if schema != nil {
				generateSchemaStruct(info, *schema, sw)
			}
		}
	}
}

func generateResponseStructs(info *pathInfo, def *chioas.Response, sw *structsWriter) {
	if !sw.opts.NoRequests {
		if def, err := sw.resolveResponse(def, nil); err == nil {
			var schema *chioas.Schema
			if def.SchemaRef != "" {
				if s, aref, err := sw.resolveSchema(def.SchemaRef, nil); err == nil {
					schema = s
					if sw.keep {
						schema = nil
						sw.writeTypeAlias(info, sw.issueSchema(aref))
					}
				}
			} else {
				schema = getSchema(def.Schema)
			}
			if schema != nil {
				generateSchemaStruct(info, *schema, sw)
			}
		}
	}
}

func generatePathsStructs(info *pathInfo, def chioas.Paths, sw *structsWriter) {
	info = copyInfo(info, "", "", "", 0)
	ks := maps.Keys(def)
	sort.Strings(ks)
	for _, k := range ks {
		path := def[k]
		pi := copyInfo(info, "", info.path+k, "", 0)
		generatePathStructs(pi, path, sw)
	}
}

func generateComponentsStructs(def chioas.Components, sw *structsWriter) {
	for _, s := range def.Schemas {
		if s.SchemaRef == "" && s.Type == values.TypeObject {
			sw.issueSchema(s.Name)
		}
	}
	if !sw.opts.NoRequests {
		ks := maps.Keys(def.Requests)
		sort.Strings(ks)
		for _, k := range ks {
			r := def.Requests[k]
			name := sw.deduper.take(toPascal("Request " + k))
			info := &pathInfo{
				name:         name,
				path:         refs.ComponentsPrefix + tags.RequestBodies + "/" + k,
				infoType:     infoTypeRequest,
				explicitName: true,
			}
			generateRequestStructs(info, &r, sw)
		}
	}
	if !sw.opts.NoResponses {
		ks := maps.Keys(def.Responses)
		sort.Strings(ks)
		for _, k := range ks {
			r := def.Responses[k]
			name := sw.deduper.take(toPascal("Response " + k))
			info := &pathInfo{
				name:         name,
				path:         refs.ComponentsPrefix + tags.Responses + "/" + k,
				infoType:     infoTypeResponse,
				explicitName: true,
			}
			generateResponseStructs(info, &r, sw)
		}
	}
}

func generateSchemaStruct(info *pathInfo, def chioas.Schema, sw *structsWriter) {
	if def.Type == values.TypeObject || def.SchemaRef != "" {
		var name string
		explicitName := false
		if info != nil {
			name = info.name
			explicitName = info.explicitName
		}
		if name == "" {
			name = def.Name
		}
		if name == "" {
			name = "Schema"
		}
		sw.writeStructStart(name, explicitName, info)
		if ptys, reqdPtys, err := sw.schemaProperties(def); err != nil {
			sw.writeLine(1, "// error - "+err.Error(), false)
		} else {
			for _, pty := range ptys {
				sw.writeProperty(1, pty, reqdPtys)
			}
		}
		sw.writeLine(0, "}", true)
	}
}

func newStructsWriter(w io.Writer, opts SchemaStructOptions) *structsWriter {
	return &structsWriter{
		writer:        newWriter(w, opts.Format, opts.UseCRLF),
		opts:          opts,
		deduper:       newNameDeDuper(),
		components:    opts.Components,
		issuedSchemas: make(map[string]string),
		keep:          opts.KeepComponentProperties,
	}
}

type structsWriter struct {
	*writer
	opts          SchemaStructOptions
	deduper       *nameDeDuper
	components    *chioas.Components
	issuedSchemas map[string]string
	keep          bool
}

func (w *structsWriter) writePrologue() {
	if w.err == nil && !w.opts.SkipPrologue {
		pkg := w.opts.Package
		if pkg == "" {
			pkg = defaultPackage
		}
		w.writeLine(0, "package "+pkg, true)
	}
}

var (
	bType    = []byte("type ")
	bStruct  = []byte(" struct {")
	bEnd     = []byte("}")
	bComment = []byte("// ")
)

func (w *structsWriter) writeTypeAlias(info *pathInfo, aType string) {
	if w.err == nil {
		name := info.name
		if !info.explicitName {
			name = w.deduper.take(toPascal(name))
		}
		if w.opts.PublicStructs {
			name = strings.ToUpper(name[:1]) + name[1:]
		} else {
			name = strings.ToLower(name[:1]) + name[1:]
		}
		if w.writeStructGoDoc(name, info) {
			if _, w.err = w.w.Write(bType); w.err == nil {
				if _, w.err = w.w.Write([]byte(name + " " + aType)); w.err == nil {
					w.writeLf(true)
				}
			}
		}
	}
}

func (w *structsWriter) writeStructStart(name string, explicitName bool, info *pathInfo) {
	if w.err == nil {
		if !explicitName {
			name = w.deduper.take(w.scopedName(toPascal(name)))
		} else {
			name = w.scopedName(name)
		}
		if w.writeStructGoDoc(name, info) {
			if _, w.err = w.w.Write(bType); w.err == nil {
				if _, w.err = w.w.Write([]byte(name)); w.err == nil {
					if _, w.err = w.w.Write(bStruct); w.err == nil {
						w.writeLf(false)
					}
				}
			}
		}
	}
}

func (w *structsWriter) writeStructGoDoc(name string, info *pathInfo) bool {
	if w.opts.GoDoc {
		if _, w.err = w.w.Write(bComment); w.err == nil {
			if _, w.err = w.w.Write([]byte(name + " ")); w.err == nil {
				if info != nil {
					_, w.err = w.w.Write(info.goDoc())
				} else {
					_, w.err = w.w.Write([]byte("schema struct"))
				}
				w.writeLf(false)
			}
		}
	}
	return w.err == nil
}

func (w *structsWriter) writeProperty(indent int, pty chioas.Property, reqdPtys []string) {
	if w.err == nil && w.writeIndent(indent) {
		fName := toPascal(pty.Name)
		fType, aType, iType, reqd, innerPtys, subs, reqdSubs, err := w.propertyType(pty, reqdPtys)
		if err != nil {
			_, w.err = w.w.Write([]byte("// " + fName + " - error: " + err.Error()))
			w.writeLf(false)
			return
		}
		if _, w.err = w.w.Write([]byte(fName + " ")); w.err == nil {
			if _, w.err = w.w.Write([]byte(fType)); w.err == nil {
				if innerPtys {
					if len(subs) == 0 {
						w.writeLine(indent+1, "// no properties", false)
					} else {
						for _, sub := range subs {
							w.writeProperty(indent+1, sub, reqdSubs)
						}
					}
					if w.writeIndent(indent) {
						_, w.err = w.w.Write(bEnd)
					}
				}
				if w.writePropertyTags(pty, aType, iType, reqd) {
					w.writeLf(false)
				}
			}
		}
	}
}

func (w *structsWriter) propertyType(def chioas.Property, reqdPtys []string) (fType string, aType string, iType string, reqd bool, innerPtys bool, subs chioas.Properties, reqdSubs []string, err error) {
	if def.SchemaRef != "" {
		var s *chioas.Schema
		var aref string
		if s, aref, err = w.resolveSchema(def.SchemaRef, nil); err != nil {
			return
		}
		subs = s.Properties
		reqdSubs = s.RequiredProperties
		aType = s.Type
		reqd = slices.Contains(reqdPtys, def.Name)
		if w.keep && w.components != nil && (aType == values.TypeArray || aType == values.TypeObject) {
			// keep component properties...
			isArray := aType == values.TypeArray
			fType = w.issueSchema(aref)
			if isArray {
				iType = values.TypeObject
				fType = "[]" + fType
			} else if !reqd {
				fType = "*" + fType
			}
			return
		}
	} else {
		subs = def.Properties
		aType = def.Type
		iType = def.ItemType
		reqd = def.Required || slices.Contains(reqdPtys, def.Name)
	}
	noPtr := false
	switch aType {
	case "string":
		fType = "string"
	case "boolean":
		fType = "bool"
	case "integer":
		fType = "int"
	case "number":
		fType = "float64"
	case "object":
		fType = "struct {\n"
		innerPtys = true
	case "array":
		noPtr = true
		switch iType {
		case "string":
			fType = "[]string"
		case "boolean":
			fType = "[]bool"
		case "integer":
			fType = "[]int"
		case "number":
			fType = "[]float64"
		case "object":
			fType = "[]struct {\n"
			innerPtys = true
		default:
			fType = "[]any"
		}
	default:
		noPtr = true
		fType = "any"
	}
	if !reqd && !noPtr {
		fType = "*" + fType
	}
	slices.SortStableFunc(subs, func(a, b chioas.Property) int {
		return strings.Compare(a.Name, b.Name)
	})
	return
}

func (w *structsWriter) issueSchema(aref string) string {
	if name, ok := w.issuedSchemas[aref]; ok {
		return name
	}
	name := w.deduper.take(w.scopedName(toPascal("schema " + aref)))
	w.issuedSchemas[aref] = name
	return name
}

func (w *structsWriter) schemaProperties(def chioas.Schema) (ptys chioas.Properties, reqdPtys []string, err error) {
	if def.SchemaRef != "" {
		if s, _, sErr := w.resolveSchema(def.SchemaRef, nil); sErr != nil {
			return nil, nil, sErr
		} else {
			ptys, reqdPtys = s.Properties, s.RequiredProperties
		}
	} else {
		ptys, reqdPtys = def.Properties, def.RequiredProperties
	}
	slices.SortStableFunc(ptys, func(a, b chioas.Property) int {
		return strings.Compare(a.Name, b.Name)
	})
	return ptys, reqdPtys, err
}

func (w *structsWriter) resolveSchema(ref string, seen map[string]struct{}) (result *chioas.Schema, aref string, err error) {
	if strings.Contains(ref, "/") && !strings.HasPrefix(ref, refs.ComponentsPrefix+tags.Schemas) {
		return nil, "", fmt.Errorf("cannot resolve external/invalid schema $ref: %s", ref)
	} else if w.components == nil {
		return nil, "", fmt.Errorf("cannot resolve internal schema $ref (no components!): %s", ref)
	}
	if seen == nil {
		seen = map[string]struct{}{}
	}
	ref = refs.Normalize(tags.Schemas, ref)
	if _, ok := seen[ref]; ok {
		return nil, "", fmt.Errorf("cyclic schema $ref: %s", ref)
	}
	for _, s := range w.components.Schemas {
		if s.Name == ref {
			if s.SchemaRef == "" {
				aref = ref
				result = &s
			} else {
				seen[ref] = struct{}{}
				result, aref, err = w.resolveSchema(s.SchemaRef, seen)
			}
			break
		}
	}
	if err == nil && result == nil {
		err = fmt.Errorf("cannot resolve internal schema $ref: %s", ref)
	}
	return result, aref, err
}

func (w *structsWriter) resolveRequest(def *chioas.Request, seen map[string]struct{}) (result *chioas.Request, err error) {
	ref := def.Ref
	if ref == "" {
		return def, nil
	} else if strings.Contains(ref, "/") && !strings.HasPrefix(ref, refs.ComponentsPrefix+tags.RequestBodies) {
		return nil, fmt.Errorf("cannot resolve external/invalid request $ref: %s", ref)
	} else if w.components == nil {
		return nil, fmt.Errorf("cannot resolve internal request $ref (no components!): %s", def.Ref)
	}
	ref = refs.Normalize(tags.RequestBodies, ref)
	if seen == nil {
		seen = map[string]struct{}{}
	}
	if _, ok := seen[ref]; ok {
		return nil, fmt.Errorf("cyclic request $ref: %s", ref)
	}
	if r, ok := w.components.Requests[ref]; ok {
		seen[ref] = struct{}{}
		return w.resolveRequest(&r, seen)
	}
	return nil, fmt.Errorf("cannot resolve internal request $ref: %s", ref)
}

func (w *structsWriter) resolveResponse(def *chioas.Response, seen map[string]struct{}) (result *chioas.Response, err error) {
	ref := def.Ref
	if ref == "" {
		return def, nil
	} else if strings.Contains(ref, "/") && !strings.HasPrefix(ref, refs.ComponentsPrefix+tags.Responses) {
		return nil, fmt.Errorf("cannot resolve external/invalid response $ref: %s", ref)
	} else if w.components == nil {
		return nil, fmt.Errorf("cannot resolve internal response $ref (no components!): %s", def.Ref)
	}
	ref = refs.Normalize(tags.Responses, ref)
	if seen == nil {
		seen = map[string]struct{}{}
	}
	if _, ok := seen[ref]; ok {
		return nil, fmt.Errorf("cyclic response $ref: %s", ref)
	}
	if r, ok := w.components.Responses[ref]; ok {
		seen[ref] = struct{}{}
		return w.resolveResponse(&r, seen)
	}
	return nil, fmt.Errorf("cannot resolve internal response $ref: %s", ref)
}

func (w *structsWriter) writePropertyTags(pty chioas.Property, aType string, iType string, reqd bool) bool {
	var tb bytes.Buffer
	tb.WriteString(" `json:")
	tb.WriteString(strconv.Quote(pty.Name))
	w.addOasTag(&tb, pty, aType, iType, reqd)
	tb.WriteString("`")
	_, w.err = w.w.Write(tb.Bytes())
	return w.err == nil
}

func (w *structsWriter) addOasTag(tb *bytes.Buffer, pty chioas.Property, aType string, iType string, reqd bool) {
	if w.opts.OASTags {
		tb.WriteString(" oas:")
		tokens := make([]string, 0)
		if reqd {
			tokens = append(tokens, "required")
		}
		if aType != "" {
			tokens = append(tokens, "type:"+aType)
		}
		if aType == "array" && iType != "" {
			tokens = append(tokens, "itemType:"+iType)
		}
		if pty.Format != "" {
			tokens = append(tokens, "format:"+pty.Format)
		}
		if pty.Constraints.Pattern != "" {
			tokens = append(tokens, "pattern:"+strconv.Quote(pty.Constraints.Pattern))
		}
		if pty.Deprecated {
			tokens = append(tokens, "deprecated")
		}
		if pty.Type == "integer" || pty.Type == "number" {
			if pty.Constraints.Maximum != "" {
				tokens = append(tokens, "maximum:"+string(pty.Constraints.Maximum))
			}
			if pty.Constraints.Minimum != "" {
				tokens = append(tokens, "minimum:"+string(pty.Constraints.Minimum))
			}
		}
		if pty.Type == "integer" && pty.Constraints.MultipleOf != 0 {
			tokens = append(tokens, fmt.Sprintf("multipleOf:%d", pty.Constraints.MultipleOf))
		}
		if pty.Type == "string" {
			if pty.Constraints.MaxLength != 0 {
				tokens = append(tokens, fmt.Sprintf("maxLength:%d", pty.Constraints.MaxLength))
			}
			if pty.Constraints.MinLength != 0 {
				tokens = append(tokens, fmt.Sprintf("minLength:%d", pty.Constraints.MinLength))
			}
		}
		if pty.Type == "array" {
			if pty.Constraints.MaxItems != 0 {
				tokens = append(tokens, fmt.Sprintf("maxItems:%d", pty.Constraints.MaxItems))
			}
			if pty.Constraints.MinItems != 0 {
				tokens = append(tokens, fmt.Sprintf("minItems:%d", pty.Constraints.MinItems))
			}
			if pty.Constraints.UniqueItems {
				tokens = append(tokens, "uniqueItems")
			}
		}
		if pty.Type == "object" {
			if pty.Constraints.MaxProperties != 0 {
				tokens = append(tokens, fmt.Sprintf("maxProperties:%d", pty.Constraints.MaxProperties))
			}
			if pty.Constraints.MinProperties != 0 {
				tokens = append(tokens, fmt.Sprintf("minProperties:%d", pty.Constraints.MinProperties))
			}
		}
		if pty.Constraints.ExclusiveMaximum {
			tokens = append(tokens, "exclusiveMaximum")
		}
		if pty.Constraints.ExclusiveMinimum {
			tokens = append(tokens, "exclusiveMinimum")
		}
		if pty.Constraints.Nullable {
			tokens = append(tokens, "nullable")
		}
		if len(pty.Enum) > 0 && (pty.Type == "string" || pty.Type == "boolean" || pty.Type == "integer" || pty.Type == "number") {
			enums := make([]string, 0, len(pty.Enum))
			for _, e := range pty.Enum {
				switch et := e.(type) {
				case string:
					enums = append(enums, strconv.Quote(et))
				case bool:
					enums = append(enums, strconv.FormatBool(et))
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					enums = append(enums, fmt.Sprintf("%d", et))
				case float32:
					enums = append(enums, strconv.FormatFloat(float64(et), 'f', -1, 32))
				case float64:
					enums = append(enums, strconv.FormatFloat(et, 'f', -1, 64))
				}
			}
			tokens = append(tokens, "enum:["+strings.Join(enums, ",")+"]")
		}
		if pty.Description != "" {
			desc := strconv.Quote(pty.Description)
			desc = desc[1 : len(desc)-1]
			tokens = append(tokens, "description:"+desc)
		}
		xks := maps.Keys(pty.Extensions)
		sort.Strings(xks)
		for _, k := range xks {
			v := pty.Extensions[k]
			switch vt := v.(type) {
			case string:
				tokens = append(tokens, "x-"+k+":"+strconv.Quote(vt))
			case bool:
				tokens = append(tokens, "x-"+k+":"+strconv.FormatBool(vt))
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				tokens = append(tokens, "x-"+k+":"+fmt.Sprintf("%d", vt))
			case float32:
				tokens = append(tokens, "x-"+k+":"+strconv.FormatFloat(float64(vt), 'f', -1, 32))
			case float64:
				tokens = append(tokens, "x-"+k+":"+strconv.FormatFloat(vt, 'f', -1, 64))
			}
		}
		tb.WriteString(strconv.Quote(strings.Join(tokens, ",")))
	}
}

func (w *structsWriter) scopedName(name string) string {
	if w.opts.PublicStructs {
		return strings.ToUpper(name[:1]) + name[1:]
	} else {
		return strings.ToLower(name[:1]) + name[1:]
	}
}

func getSchema(schema any) *chioas.Schema {
	if schema == nil {
		return nil
	}
	switch s := schema.(type) {
	case *chioas.Schema:
		return s
	case chioas.Schema:
		return &s
	}
	return nil
}

func statusCodeName(status int) string {
	if n, ok := httpStatusNames[status]; ok {
		return n
	}
	return strconv.Itoa(status)
}

var httpStatusNames = map[int]string{
	http.StatusContinue:                      "Continue",
	http.StatusSwitchingProtocols:            "SwitchingProtocols",
	http.StatusProcessing:                    "Processing",
	http.StatusEarlyHints:                    "EarlyHints",
	http.StatusOK:                            "Ok",
	http.StatusCreated:                       "Created",
	http.StatusAccepted:                      "Accepted",
	http.StatusNonAuthoritativeInfo:          "NonAuthoritativeInfo",
	http.StatusNoContent:                     "NoContent",
	http.StatusResetContent:                  "ResetContent",
	http.StatusPartialContent:                "PartialContent",
	http.StatusMultiStatus:                   "MultiStatus",
	http.StatusAlreadyReported:               "AlreadyReported",
	http.StatusIMUsed:                        "IMUsed",
	http.StatusMultipleChoices:               "MultipleChoices",
	http.StatusMovedPermanently:              "MovedPermanently",
	http.StatusFound:                         "Found",
	http.StatusSeeOther:                      "SeeOther",
	http.StatusNotModified:                   "NotModified",
	http.StatusUseProxy:                      "UseProxy",
	http.StatusTemporaryRedirect:             "TemporaryRedirect",
	http.StatusPermanentRedirect:             "PermanentRedirect",
	http.StatusBadRequest:                    "BadRequest",
	http.StatusUnauthorized:                  "Unauthorized",
	http.StatusPaymentRequired:               "PaymentRequired",
	http.StatusForbidden:                     "Forbidden",
	http.StatusNotFound:                      "NotFound",
	http.StatusMethodNotAllowed:              "MethodNotAllowed",
	http.StatusNotAcceptable:                 "NotAcceptable",
	http.StatusProxyAuthRequired:             "ProxyAuthRequired",
	http.StatusRequestTimeout:                "RequestTimeout",
	http.StatusConflict:                      "Conflict",
	http.StatusGone:                          "Gone",
	http.StatusLengthRequired:                "LengthRequired",
	http.StatusPreconditionFailed:            "PreconditionFailed",
	http.StatusRequestEntityTooLarge:         "RequestEntityTooLarge",
	http.StatusRequestURITooLong:             "RequestURITooLong",
	http.StatusUnsupportedMediaType:          "UnsupportedMediaType",
	http.StatusRequestedRangeNotSatisfiable:  "RequestedRangeNotSatisfiable",
	http.StatusExpectationFailed:             "ExpectationFailed",
	http.StatusTeapot:                        "Teapot",
	http.StatusMisdirectedRequest:            "MisdirectedRequest",
	http.StatusUnprocessableEntity:           "UnprocessableEntity",
	http.StatusLocked:                        "Locked",
	http.StatusFailedDependency:              "FailedDependency",
	http.StatusTooEarly:                      "TooEarly",
	http.StatusUpgradeRequired:               "UpgradeRequired",
	http.StatusPreconditionRequired:          "PreconditionRequired",
	http.StatusTooManyRequests:               "TooManyRequests",
	http.StatusRequestHeaderFieldsTooLarge:   "RequestHeaderFieldsTooLarge",
	http.StatusUnavailableForLegalReasons:    "UnavailableForLegalReasons",
	http.StatusInternalServerError:           "InternalServerError",
	http.StatusNotImplemented:                "NotImplemented",
	http.StatusBadGateway:                    "BadGateway",
	http.StatusServiceUnavailable:            "ServiceUnavailable",
	http.StatusGatewayTimeout:                "GatewayTimeout",
	http.StatusHTTPVersionNotSupported:       "HTTPVersionNotSupported",
	http.StatusVariantAlsoNegotiates:         "VariantAlsoNegotiates",
	http.StatusInsufficientStorage:           "InsufficientStorage",
	http.StatusLoopDetected:                  "LoopDetected",
	http.StatusNotExtended:                   "NotExtended",
	http.StatusNetworkAuthenticationRequired: "NetworkAuthenticationRequired",
}
