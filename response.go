package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Responses is a map of Response where the key is the http status code
type Responses map[int]Response

func (r Responses) writeYaml(w yaml.Writer) {
	if l := len(r); l > 0 {
		sortCodes := make([]int, 0, l)
		for rc := range r {
			sortCodes = append(sortCodes, rc)
		}
		sort.Ints(sortCodes)
		w.WriteTagStart(tagNameResponses)
		for _, sc := range sortCodes {
			r[sc].writeYaml(sc, w)
		}
		w.WriteTagEnd()
	}
}

// Response is the OAS definition of a response
type Response struct {
	// Ref is the OAS $ref name for the response
	//
	// If this is a non-empty string and the response is used by Method.Responses, then a $ref to "#/components/responses/" is used
	//
	// If the Response is used by Components.ResponseBodies this value is ignored
	Ref string
	// Description is the OAS description
	Description string
	// NoContent indicates that this response has not content
	//
	// This does not eed to set this when status code is http.StatusNoContent
	NoContent bool
	// ContentType is the OAS content type
	//
	// defaults to "application/json"
	ContentType string
	// Schema is the optional OAS Schema
	//
	// Only used if the value is non-nil - otherwise uses SchemaRef is used
	//
	// The value can be any of the following:
	//
	// * chioas.Schema (or *chioas.Schema)
	//
	// * a chioas.SchemaConverter
	//
	// * a chioas.SchemaWriter
	//
	// * a struct or ptr to struct (schema written is determined by examining struct fields)
	//
	// * a slice of structs (items schema written is determined by examining struct fields)
	Schema any
	// SchemaRef is the OAS schema reference
	//
	// Only used if value is a non-empty string - if both Schema is nil and SchemaRef is empty string, then an
	// empty object schema is written to the spec yaml, e.g.
	//   schema:
	//     type: "object"
	//
	// If the value does not contain a path (i.e. does not contain any "/") then the ref
	// path will be the value prefixed with components schemas path.  For example, specifying "foo"
	// will result in a schema ref:
	//   schema:
	//     $ref: "#/components/schemas/foo"
	SchemaRef string
	// IsArray indicates that request is an array of items
	IsArray bool
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
}

func (r Response) writeYaml(statusCode int, w yaml.Writer) {
	w.WriteTagStart(strconv.Itoa(statusCode))
	if r.Ref == "" {
		desc := r.Description
		if desc == "" {
			desc = http.StatusText(statusCode)
		}
		w.WriteTagValue(tagNameDescription, desc)
		if !r.NoContent && statusCode != http.StatusNoContent {
			writeContent(r.ContentType, r.Schema, r.SchemaRef, r.IsArray, w)
		}
		writeExtensions(r.Extensions, w)
		writeAdditional(r.Additional, r, w)
	} else {
		w.WriteTagValue(tagNameRef, refPathResponses+r.Ref)
	}
	w.WriteTagEnd()
}

func (r Response) componentsWriteYaml(name string, w yaml.Writer) {
	w.WriteTagStart(name)
	w.WriteTagValue(tagNameDescription, r.Description)
	if !r.NoContent {
		writeContent(r.ContentType, r.Schema, r.SchemaRef, r.IsArray, w)
	}
	writeExtensions(r.Extensions, w)
	writeAdditional(r.Additional, r, w)
	w.WriteTagEnd()
}

func writeContent(contentType string, schema any, schemaRef string, isArray bool, w yaml.Writer) {
	w.WriteTagStart(tagNameContent)
	if contentType != "" {
		w.WriteTagStart(contentType)
	} else {
		w.WriteTagStart(tagNameApplicationJson)
	}
	w.WriteTagStart(tagNameSchema)
	if schema != nil {
		writeSchema(schema, isArray, w)
	} else if schemaRef != "" {
		writeSchemaRef(schemaRef, isArray, w)
	} else {
		w.WriteTagValue(tagNameType, tagValueTypeObject)
	}
	w.WriteTagEnd().
		WriteTagEnd().
		WriteTagEnd()
}

func writeSchemaRef(ref string, isArray bool, w yaml.Writer) {
	if isArray {
		w.WriteTagValue(tagNameType, tagValueTypeArray).
			WriteTagStart(tagNameItems)
	}
	if strings.Contains(ref, "/") {
		w.WriteTagValue(tagNameRef, ref)
	} else {
		w.WriteTagValue(tagNameRef, refPathSchemas+ref)
	}
	if isArray {
		w.WriteTagEnd()
	}
}

func writeSchema(schema any, isArray bool, w yaml.Writer) {
	actual, writer := isActualSchema(schema)
	if actual == nil && writer == nil {
		actual, isArray = extractSchema(schema, isArray)
	}
	if actual != nil {
		if isArray {
			w.WriteTagValue(tagNameType, tagValueTypeArray).
				WriteTagStart(tagNameItems)
		}
		actual.writeYaml(false, w)
		if isArray {
			w.WriteTagEnd()
		}
	} else if writer != nil {
		if isArray {
			w.WriteTagValue(tagNameType, tagValueTypeArray).
				WriteTagStart(tagNameItems)
		}
		writer.WriteSchema(w)
		if isArray {
			w.WriteTagEnd()
		}
	} else {
		w.WriteTagValue(tagNameType, tagValueTypeNull)
	}
}

func isActualSchema(schema any) (*Schema, SchemaWriter) {
	if conv, ok := schema.(SchemaConverter); ok {
		return conv.ToSchema(), nil
	} else if writer, ok := schema.(SchemaWriter); ok {
		return nil, writer
	}
	switch ts := schema.(type) {
	case Schema:
		return &ts, nil
	case *Schema:
		return ts, nil
	}
	return nil, nil
}

func extractSchema(example any, isArray bool) (*Schema, bool) {
	arrayDetect := isArray
	st := reflect.TypeOf(example)
	sv := reflect.ValueOf(example)
	hasValue := true
	if st.Kind() == reflect.Pointer {
		st = st.Elem()
		sv = sv.Elem()
	}
	if st.Kind() == reflect.Slice {
		arrayDetect = true
		st = st.Elem()
		if sv.Len() > 0 {
			sv = sv.Index(0)
		} else {
			hasValue = false
		}
	}
	if st.Kind() == reflect.Struct {
		ptys := make([]Property, 0)
		reqds := make([]string, 0)
		for i := 0; i < st.NumField(); i++ {
			fld := st.Field(i)
			if fld.IsExported() {
				var fv *reflect.Value
				if hasValue {
					fvv := sv.FieldByName(fld.Name)
					fv = &fvv
				}
				pty, reqd := extractSchemaField(fld, fv)
				ptys = append(ptys, pty)
				if reqd {
					reqds = append(reqds, pty.Name)
				}
			}
		}
		return &Schema{
			RequiredProperties: reqds,
			Properties:         ptys,
		}, arrayDetect
	}
	return nil, false
}

func extractSchemaField(fld reflect.StructField, value *reflect.Value) (Property, bool) {
	reqd := true
	name := fld.Name
	if tgv, ok := fld.Tag.Lookup("json"); ok {
		if tgs := strings.Split(tgv, ","); len(tgs) > 1 {
			name = tgs[0]
			reqd = tgs[1] != "omitempty"
		} else {
			name = tgv
		}
	}
	k := fld.Type.Kind()
	if k == reflect.Pointer {
		reqd = false
		k = fld.Type.Elem().Kind()
	}
	ft := "null"
	var eg any
	switch k {
	case reflect.String:
		ft = tagValueTypeString
		if value != nil {
			eg = value.Interface()
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ft = tagValueTypeInteger
		if value != nil {
			eg = value.Interface()
		}
	case reflect.Float32, reflect.Float64:
		ft = tagValueTypeNumber
		if value != nil {
			eg = value.Interface()
		}
	case reflect.Bool:
		ft = tagValueTypeBoolean
		if value != nil {
			eg = value.Interface()
		}
	case reflect.Slice:
		ft = tagValueTypeArray
	case reflect.Map, reflect.Struct:
		ft = tagValueTypeObject
	}
	return Property{
		Name:    name,
		Type:    ft,
		Example: eg,
	}, reqd
}
