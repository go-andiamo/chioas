package chioas

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"reflect"
	"strings"
)

type Schemas []Schema

func (ss Schemas) writeYaml(w yaml.Writer) {
	if len(ss) > 0 {
		// check for duplicates...
		m := map[string]struct{}{}
		for _, s := range ss {
			if s.Name != "" {
				if _, ok := m[s.Name]; ok {
					w.SetError(fmt.Errorf("duplicate schema name '%s' in components", s.Name))
					return
				}
				m[s.Name] = struct{}{}
			}
		}
		w.WriteTagStart(tagNameSchemas)
		for _, s := range ss {
			s.writeYaml(true, w)
		}
		w.WriteTagEnd()
	}
}

// SchemaConverter is an interface that can be implemented by anything to convert it to a schema
type SchemaConverter interface {
	ToSchema() *Schema
}

// SchemaWriter is an interface that can be implemented by anything to write a schema for that item
type SchemaWriter interface {
	WriteSchema(w yaml.Writer)
}

// Schema represents the OAS definition of a schema
type Schema struct {
	// Name is the OAS name of the schema
	Name string
	// Description is the OAS description
	Description string
	// Type is the OAS type
	//
	// Should be one of "string", "object", "array", "boolean", "integer", "number" or "null"
	Type string
	// Format is the OAS format
	Format string
	// RequiredProperties is the ordered collection of required properties
	//
	// If any of the items in Properties is also denoted as Property.Required, these are
	// automatically added to RequiredProperties
	RequiredProperties []string
	// Properties is the ordered collection of properties
	Properties Properties
	// Default is the OAS default
	Default any
	// Example is the OAS example for the schema
	Example any
	// Enum is the OAS enum
	Enum []any
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
	// SchemaRef is the OAS schema reference
	//
	// Only used if value is a non-empty string
	//
	// If the value does not contain a path (i.e. does not contain any "/") then the ref
	// path will be the value prefixed with components schemas path.  For example, specifying "foo"
	// will result in a schema ref:
	//   schema:
	//     $ref: "#/components/schemas/foo"
	SchemaRef string
	// Discriminator is the optional OAS discriminator for the schema
	Discriminator *Discriminator
	// Ofs is the optional OAS ofs (oneOf, anyOf or allOf) for the schema
	Ofs *Ofs
}

func (s *Schema) Ref() string {
	panic("*Schema does not contain Ref() string")
}

func (s *Schema) Schema() *Schema {
	return s
}

func (s *Schema) IsRef() bool {
	return false
}

func (s *Schema) writeYaml(withName bool, w yaml.Writer) {
	if withName {
		if s.Name == "" {
			w.SetError(errors.New("schema without name"))
			return
		}
		w.WriteTagStart("\"" + s.Name + "\"")
	}
	w.WriteComments(s.Comment)
	if s.SchemaRef != "" {
		writeSchemaRef(s.SchemaRef, s.Type == tagValueTypeArray, w)
	} else {
		w.WriteTagValue(tagNameDescription, s.Description).
			WriteTagValue(tagNameType, defValue(s.Type, tagValueTypeObject)).
			WriteTagValue(tagNameFormat, s.Format)
		if s.Ofs != nil {
			s.Ofs.writeYaml(w)
		}
		if s.Discriminator != nil {
			s.Discriminator.writeYaml(w)
		}
		if reqs, has := s.getRequiredProperties(); has {
			w.WriteTagStart(tagNameRequired)
			for _, rp := range reqs {
				w.WriteItem(rp)
			}
			w.WriteTagEnd()
		}
		if len(s.Properties) > 0 {
			w.WriteTagStart(tagNameProperties)
			for _, p := range s.Properties {
				p.writeYaml(w, true)
			}
			w.WriteTagEnd()
		}
		w.WriteTagValue(tagNameDefault, s.Default).
			WriteTagValue(tagNameExample, s.Example)
		if len(s.Enum) > 0 {
			w.WriteTagStart(tagNameEnum)
			for _, e := range s.Enum {
				w.WriteItem(e)
			}
			w.WriteTagEnd()
		}
		writeExtensions(s.Extensions, w)
		writeAdditional(s.Additional, s, w)
	}
	if withName {
		w.WriteTagEnd()
	}
}

func (s *Schema) writeOfYaml(w yaml.Writer) {
	w.WriteTagValue(tagNameDescription, s.Description).
		WriteItemStart(tagNameType, defValue(s.Type, tagValueTypeObject)).
		WriteTagValue(tagNameFormat, s.Format)
	if reqs, has := s.getRequiredProperties(); has {
		w.WriteTagStart(tagNameRequired)
		for _, rp := range reqs {
			w.WriteItem(rp)
		}
		w.WriteTagEnd()
	}
	if len(s.Properties) > 0 {
		w.WriteTagStart(tagNameProperties)
		for _, p := range s.Properties {
			p.writeYaml(w, true)
		}
		w.WriteTagEnd()
	}
	w.WriteTagValue(tagNameDefault, s.Default).
		WriteTagValue(tagNameExample, s.Example)
	if len(s.Enum) > 0 {
		w.WriteTagStart(tagNameEnum)
		for _, e := range s.Enum {
			w.WriteItem(e)
		}
		w.WriteTagEnd()
	}
	writeExtensions(s.Extensions, w)
	writeAdditional(s.Additional, s, w)
	w.WriteTagEnd()
}

func (s *Schema) getRequiredProperties() ([]string, bool) {
	result := make([]string, 0, len(s.RequiredProperties))
	m := map[string]bool{}
	for _, v := range s.RequiredProperties {
		if !m[v] {
			result = append(result, v)
			m[v] = true
		}
	}
	for _, pty := range s.Properties {
		if pty.Required && !m[pty.Name] {
			result = append(result, pty.Name)
			m[pty.Name] = true
		}
	}
	return result, len(result) > 0
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
