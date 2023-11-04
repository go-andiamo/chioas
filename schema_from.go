package chioas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-andiamo/splitter"
	"golang.org/x/exp/slices"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// From generates a schema from the provided sample
//
// where the sample must be a struct or pointer to struct
//
// example:
//
//	type MyRequest struct {
//		GivenName    string   `json:"givenName" oas:"description:'Your first/given name',required,pattern:'^[A-Z][a-zA-Z]+$',example"`
//		FamilyName   string   `json:"familyName" oas:"description:'Your family name/surname',required,pattern:'^[A-Z][a-zA-Z]+$',example"`
//		Age          int      `oas:"name:age,required:true,example,#this is a comment,#'this is another, with commas in it, comment'"`
//		SiblingNames []string `oas:"name:siblings,$ref:'Siblings'"`
//	}
//	def := chioas.Definition{
//		Components: &chioas.Components{
//			Schemas: chioas.Schemas{
//				(&Schema{
//					Name:        "MyRequest",
//					Description: "My request",
//				}).Must(MyRequest{
//					GivenName:  "Bilbo",
//					FamilyName: "Baggins",
//					Age:        50,
//				}),
//			},
//		},
//	}
//	data, _ := def.AsYaml()
//	println(string(data))
//
// would produce the following OAS spec yaml:
//
//	openapi: "3.0.3"
//	info:
//	  title: "API Documentation"
//	  version: "1.0.0"
//	paths:
//	components:
//	  schemas:
//	    "MyRequest":
//	      description: "My request"
//	      type: "object"
//	      required:
//	        - "givenName"
//	        - "familyName"
//	        - "age"
//	      properties:
//	        "givenName":
//	          description: "Your first/given name"
//	          type: "string"
//	          example: "Bilbo"
//	          pattern: "^[A-Z][a-zA-Z]+$"
//	        "familyName":
//	          description: "Your family name/surname"
//	          type: "string"
//	          example: "Baggins"
//	          pattern: "^[A-Z][a-zA-Z]+$"
//	        "age":
//	          #this is a comment
//	          #this is another, with commas in it, comment
//	          type: "integer"
//	          example: 50
//	        "siblings":
//	          type: "array"
//	          items:
//	            $ref: "#/components/schemas/Siblings"
//
// In the field tags, the following OAS tokens can be used:
//
//	$ref:string (sets the SchemaRef)
//	deprecated:true|false
//	deprecated (same as deprecated:true)
//	description:string
//	enum:any (adds an enum value to the property)
//	enum:[any,any,...] (adds multiple enum values to the property)
//	example (sets the Property.Example from the provided sample - if available)
//	exclusiveMaximum:true|false
//	exclusiveMaximum (same as exclusiveMaximum:true)
//	exclusiveMinimum:true|false
//	exclusiveMinimum (same as exclusiveMinimum:true)
//	format:string
//	itemType:""|"string"|"object"|"array"|"boolean"|"integer"|"number"|"null"
//	maxItems:uint
//	maxLength:uint
//	maxProperties:uint
//	maximum:number
//	minItems:uint
//	minLength:uint
//	minProperties:uint
//	minimum:number
//	multipleOf:uint
//	name:string
//	nullable:true|false
//	nullable (same as nullable:true)
//	pattern:string
//	required:true|false
//	required (same as required:true)
//	type:""|"string"|"object"|"array"|"boolean"|"integer"|"number"|"null"
//	uniqueItems:true|false
//	uniqueItems (same as uniqueItems:true)
//	x-...:string (sets an OAS extension property)
//	#comment (adds a comment to the property)
func (s *Schema) From(sample any) (*Schema, error) {
	var t reflect.Type
	var vo *reflect.Value
	if sample != nil {
		switch vt := sample.(type) {
		case reflect.Type:
			t = vt
		default:
			t = reflect.TypeOf(sample)
			voe := reflect.ValueOf(sample)
			vo = &voe
		}
	}
	if t != nil {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
			if vo != nil {
				voe := vo.Elem()
				vo = &voe
			}
		}
		if t.Kind() == reflect.Struct {
			return s.schemaFrom(t, vo)
		}
	}
	return nil, errors.New("sample must be a struct")
}

// MustFrom is the same as From except it panics on error
func (s *Schema) MustFrom(sample any) *Schema {
	if _, err := s.From(sample); err == nil {
		return s
	} else {
		panic(err)
	}
}

// Must is the same as MustFrom except it panics on error or returns non-pointer Schema
func (s *Schema) Must(sample any) Schema {
	return *(s.MustFrom(sample))
}

func (s *Schema) schemaFrom(t reflect.Type, vo *reflect.Value) (*Schema, error) {
	s.Type = tagValueTypeObject
	if ptys, err := propertiesFrom(t, vo); err == nil {
		s.Properties = ptys
		for _, pty := range ptys {
			if pty.Required {
				s.RequiredProperties = append(s.RequiredProperties, pty.Name)
			}
		}
	} else {
		return nil, err
	}
	return s, nil
}

func propertiesFrom(t reflect.Type, vo *reflect.Value) ([]Property, error) {
	useT := t
	if t.Kind() == reflect.Pointer {
		useT = useT.Elem()
	}
	l := useT.NumField()
	ptys := make([]Property, 0, l)
	for i := 0; i < l; i++ {
		fld := useT.Field(i)
		if fld.IsExported() {
			if pty, err := propertyFrom(fld, vo); err == nil && pty != nil {
				ptys = append(ptys, *pty)
			} else if err != nil {
				return nil, err
			}
		}
	}
	return ptys, nil
}

const (
	tagNameJson = "json"
	tagNameOas  = "oas"
)

func propertyFrom(fld reflect.StructField, vo *reflect.Value) (*Property, error) {
	pty := &Property{}
	if !setNameFromJsonTag(pty, fld) {
		return nil, nil
	}
	if err := setFromOasTag(pty, fld, vo); err != nil {
		return nil, err
	}
	setPropertyType(pty, fld)
	if pty.SchemaRef == "" {
		if pty.Type == tagValueTypeObject || pty.Type == tagValueTypeArray {
			k := fld.Type.Kind()
			if k == reflect.Pointer {
				k = fld.Type.Elem().Kind()
			}
			switch k {
			case reflect.Struct:
				if pty.Type == tagValueTypeObject {
					return pty, setSubPropertiesStruct(pty, fld, vo)
				}
			case reflect.Slice:
				if pty.Type == tagValueTypeArray {
					return pty, setSubPropertiesSlice(pty, fld, vo)
				}
			}
		}
	}
	return pty, nil
}

func setPropertyType(pty *Property, fld reflect.StructField) {
	if pty.Type == "" {
		k := fld.Type.Kind()
		if k == reflect.Pointer {
			k = fld.Type.Elem().Kind()
		}
		oasType, oasFormat, _ := toOasTypeAndFormat(k, fld.Type)
		pty.Type = oasType
		if pty.Format == "" {
			pty.Format = oasFormat
		}
	}
}

func setSubPropertiesStruct(pty *Property, fld reflect.StructField, vo *reflect.Value) error {
	subT := fld.Type
	if subT.Kind() == reflect.Pointer {
		subT = fld.Type.Elem()
	}
	var subVo *reflect.Value
	if vo != nil && vo.Kind() != reflect.Pointer {
		if voe := vo.FieldByName(fld.Name); voe.IsValid() {
			subVo = &voe
		}
	}
	if subPtys, err := propertiesFrom(subT, subVo); err == nil {
		pty.Properties = subPtys
	} else {
		return err
	}
	return nil
}

func setSubPropertiesSlice(pty *Property, fld reflect.StructField, vo *reflect.Value) error {
	subT := fld.Type.Elem()
	k := subT.Kind()
	if k == reflect.Pointer {
		k = subT.Elem().Kind()
	}
	oasType, _, isStruct := toOasTypeAndFormat(k, subT)
	if oasType == tagValueTypeArray {
		return errors.New("arrays of array not supported")
	}
	pty.ItemType = oasType
	if isStruct {
		var subVo *reflect.Value
		if vo != nil && vo.IsValid() {
			if vo.Kind() == reflect.Struct {
				voe := vo.FieldByName(fld.Name)
				if voe.IsValid() && voe.Len() > 0 {
					voe = voe.Index(0)
					subVo = &voe
				}
			} else if vo.Kind() == reflect.Slice && vo.Len() > 0 {
				if voe := vo.Index(0); voe.IsValid() {
					subVo = &voe
				}
			}
		}
		if subPtys, err := propertiesFrom(subT, subVo); err == nil {
			pty.Properties = subPtys
		} else {
			return err
		}
	}
	return nil
}

var dtType = reflect.TypeOf(time.Time{})
var dtTypePtr = reflect.TypeOf(&time.Time{})
var jn = json.Number("")
var jnType = reflect.TypeOf(jn)
var jnTypePtr = reflect.TypeOf(&jn)

func toOasTypeAndFormat(k reflect.Kind, t reflect.Type) (oasType string, oasFormat string, isStruct bool) {
	switch k {
	case reflect.Struct:
		if t == dtType || t == dtTypePtr {
			oasType = tagValueTypeString
			oasFormat = "date-time"
		} else {
			oasType = tagValueTypeObject
			isStruct = true
		}
	case reflect.Map, reflect.Interface:
		oasType = tagValueTypeObject
	case reflect.Slice:
		oasType = tagValueTypeArray
	case reflect.String:
		if t == jnType || t == jnTypePtr {
			oasType = tagValueTypeNumber
		} else {
			oasType = tagValueTypeString
		}
	case reflect.Bool:
		oasType = tagValueTypeBoolean
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Uint, reflect.Uint8, reflect.Uint16:
		oasType = tagValueTypeInteger
	case reflect.Int32, reflect.Uint32:
		oasType = tagValueTypeInteger
		oasFormat = "int32"
	case reflect.Int64, reflect.Uint64:
		oasType = tagValueTypeInteger
		oasFormat = "int64"
	case reflect.Float32:
		oasType = tagValueTypeNumber
		oasFormat = "float"
	case reflect.Float64:
		oasType = tagValueTypeNumber
		oasFormat = "double"
	}
	return
}

func setNameFromJsonTag(pty *Property, fld reflect.StructField) bool {
	if jt, ok := fld.Tag.Lookup(tagNameJson); ok {
		if jt == "-" {
			return false
		} else if jt == "-," {
			pty.Name = "-"
		} else if cAt := strings.IndexByte(jt, ','); cAt > 0 {
			pty.Name = jt[:cAt]
		} else if jt != "" && cAt == -1 {
			pty.Name = jt
		}
	}
	if pty.Name == "" {
		pty.Name = fld.Name
	}
	return true
}

func setFromOasTag(pty *Property, fld reflect.StructField, vo *reflect.Value) error {
	if oas, ok := fld.Tag.Lookup(tagNameOas); ok {
		if tokens, err := oasTagSplitter.Split(oas, splitter.TrimSpaces, splitter.IgnoreEmpties); err == nil {
			for _, token := range tokens {
				if strings.HasPrefix(token, "#") {
					cmt := unquoteString(token[1:])
					if pty.Comment == "" {
						pty.Comment = cmt
					} else {
						pty.Comment = pty.Comment + "\n" + cmt
					}
				}
			}
			slices.Sort(tokens) // need to sort so that $ref is first
			seenRef := false
			for _, token := range tokens {
				if !strings.HasPrefix(token, "#") {
					if isRef, err := setFromOasToken(seenRef, token, pty, fld, vo); err != nil {
						return err
					} else {
						seenRef = seenRef || isRef
					}
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func setFromOasToken(seenRef bool, token string, pty *Property, fld reflect.StructField, vo *reflect.Value) (bool, error) {
	parts, err := oasColonSplitter.Split(token,
		splitter.TrimSpaces,
		splitter.NotEmptyFirstMsg(fmt.Sprintf("oas must have a token name - '%s'", token)),
		splitter.NotEmptyLastMsg(fmt.Sprintf("oas must have a token value - '%s'", token)))
	if err != nil {
		return false, err
	} else if len(parts) > 2 {
		return false, fmt.Errorf("invalid oas tag token '%s'", token)
	}
	value := ""
	hasValue := len(parts) > 1
	if hasValue {
		token = parts[0]
		value = parts[1]
	}
	isRef := false
	if token == tagNameExample {
		if hasValue {
			return false, fmt.Errorf("oas tag token '%s' - must not have a value (flag only)", token)
		}
		if !seenRef && vo != nil {
			pty.Example = extractExample(fld, vo)
		}
	} else if strings.HasPrefix(token, "x-") {
		if !hasValue {
			return false, fmt.Errorf("invalid oas tag token '%s' (missing value)", token)
		} else if !seenRef {
			if pty.Extensions == nil {
				pty.Extensions = map[string]any{}
			}
			pty.Extensions[token] = literalValue(value)
		}
	} else if sf, ok := tokenSetters[token]; ok {
		if !hasValue && !allowedTokenOnly[token] {
			return false, fmt.Errorf("invalid oas tag token '%s' (missing value)", token)
		}
		if !seenRef || tokensAfterRef[token] {
			if err := sf(pty, value, hasValue); err != nil {
				return false, err
			}
		}
		isRef = token == tagNameRef
	} else {
		return false, fmt.Errorf("unknown oas tag token '%s'", token)
	}
	return isRef, nil
}

func extractExample(fld reflect.StructField, vo *reflect.Value) (result any) {
	if vo != nil && vo.IsValid() {
		if subFld := vo.FieldByName(fld.Name); subFld.IsValid() {
			switch subFld.Kind() {
			case reflect.Slice:
				if subFld.Len() > 0 {
					result = subFld.Index(0).Interface()
				}
			default:
				result = subFld.Interface()
			}
		}
	}
	return
}

var tokensAfterRef = map[string]bool{
	tagNameRef:      true,
	tagNameName:     true,
	tagNameType:     true,
	tagNameItemType: true,
	tagNameRequired: true,
}

var allowedTokenOnly = map[string]bool{
	tagNameRequired:         true,
	tagNameDeprecated:       true,
	tagNameExample:          true,
	tagNameExclusiveMaximum: true,
	tagNameExclusiveMinimum: true,
	tagNameNullable:         true,
	tagNameUniqueItems:      true,
}

func checkOasType(token string, value string) error {
	if value == "" || value == "string" || value == "object" || value == "array" ||
		value == "boolean" || value == "integer" || value == "number" || value == "null" {
		return nil
	}
	return fmt.Errorf(`invalid oas token '%s' value '%s' (must be: ""|"string"|"object"|"array"|"boolean"|"integer"|"number"|"null")`, token, value)
}

var tokenSetters = map[string]func(pty *Property, value string, hasValue bool) error{
	tagNameRef: func(pty *Property, value string, hasValue bool) error {
		pty.SchemaRef = unquoteString(value)
		return nil
	},
	tagNameDeprecated: func(pty *Property, value string, hasValue bool) error {
		if hasValue {
			if b, err := strconv.ParseBool(value); err == nil {
				pty.Deprecated = b
			} else {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameDeprecated, value)
			}
		} else {
			pty.Deprecated = true
		}
		return nil
	},
	tagNameDescription: func(pty *Property, value string, hasValue bool) error {
		pty.Description = unquoteString(value)
		return nil
	},
	tagNameEnum: func(pty *Property, value string, hasValue bool) error {
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			pts, err := oasTagSplitter.Split(value[1:len(value)-1], splitter.TrimSpaces, splitter.IgnoreEmpties)
			if err != nil {
				return err
			}
			for _, v := range pts {
				pty.Enum = append(pty.Enum, literalValue(v))
			}
		} else {
			pty.Enum = append(pty.Enum, literalValue(value))
		}
		return nil
	},
	tagNameExclusiveMaximum: func(pty *Property, value string, hasValue bool) error {
		if hasValue {
			if b, err := strconv.ParseBool(value); err == nil {
				pty.Constraints.ExclusiveMaximum = b
			} else {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameExclusiveMaximum, value)
			}
		} else {
			pty.Constraints.ExclusiveMaximum = true
		}
		return nil
	},
	tagNameExclusiveMinimum: func(pty *Property, value string, hasValue bool) error {
		if hasValue {
			if b, err := strconv.ParseBool(value); err == nil {
				pty.Constraints.ExclusiveMinimum = b
			} else {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameExclusiveMinimum, value)
			}
		} else {
			pty.Constraints.ExclusiveMinimum = true
		}
		return nil
	},
	tagNameFormat: func(pty *Property, value string, hasValue bool) error {
		pty.Format = unquoteString(value)
		return nil
	},
	tagNameItemType: func(pty *Property, value string, hasValue bool) error {
		pty.ItemType = unquoteString(value)
		return checkOasType(tagNameItemType, pty.ItemType)
	},
	tagNameMaximum: func(pty *Property, value string, hasValue bool) error {
		jn := json.Number(value)
		if _, err := jn.Int64(); err != nil {
			if _, err = jn.Float64(); err != nil {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMaximum, value)
			}
		}
		pty.Constraints.Maximum = jn
		return nil
	},
	tagNameMinimum: func(pty *Property, value string, hasValue bool) error {
		jn := json.Number(value)
		if _, err := jn.Int64(); err != nil {
			if _, err = jn.Float64(); err != nil {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMinimum, value)
			}
		}
		pty.Constraints.Minimum = jn
		return nil
	},
	tagNameMaxItems: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MaxItems = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMaxItems, value)
	},
	tagNameMinItems: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MinItems = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMinItems, value)
	},
	tagNameMaxLength: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MaxLength = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMaxLength, value)
	},
	tagNameMinLength: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MinLength = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMinLength, value)
	},
	tagNameMaxProperties: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MaxProperties = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMaxProperties, value)
	},
	tagNameMinProperties: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MinProperties = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMinProperties, value)
	},
	tagNameMultipleOf: func(pty *Property, value string, hasValue bool) error {
		if i, err := strconv.ParseUint(value, 10, 32); err == nil {
			pty.Constraints.MultipleOf = uint(i)
			return nil
		}
		return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameMultipleOf, value)
	},
	tagNameName: func(pty *Property, value string, hasValue bool) error {
		pty.Name = unquoteString(value)
		return nil
	},
	tagNameNullable: func(pty *Property, value string, hasValue bool) error {
		if hasValue {
			if b, err := strconv.ParseBool(value); err == nil {
				pty.Constraints.Nullable = b
			} else {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameNullable, value)
			}
		} else {
			pty.Constraints.Nullable = true
		}
		return nil
	},
	tagNamePattern: func(pty *Property, value string, hasValue bool) error {
		pty.Constraints.Pattern = unquoteString(value)
		return nil
	},
	tagNameRequired: func(pty *Property, value string, hasValue bool) error {
		if hasValue {
			if b, err := strconv.ParseBool(value); err == nil {
				pty.Required = b
			} else {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameRequired, value)
			}
		} else {
			pty.Required = true
		}
		return nil
	},
	tagNameType: func(pty *Property, value string, hasValue bool) error {
		pty.Type = unquoteString(value)
		return checkOasType(tagNameType, pty.Type)
	},
	tagNameUniqueItems: func(pty *Property, value string, hasValue bool) error {
		if hasValue {
			if b, err := strconv.ParseBool(value); err == nil {
				pty.Constraints.UniqueItems = b
			} else {
				return fmt.Errorf("invalid oas token '%s' value '%s'", tagNameUniqueItems, value)
			}
		} else {
			pty.Constraints.UniqueItems = true
		}
		return nil
	},
}

func literalValue(s string) yaml.LiteralValue {
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) || (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return yaml.LiteralValue{Value: `"` + strings.ReplaceAll(s[1:len(s)-1], `"`, `\"`) + `"`}
	}
	return yaml.LiteralValue{Value: s}
}

func unquoteString(s string) string {
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) || (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1]
	}
	return s
}

var splitterEncs = []*splitter.Enclosure{splitter.DoubleQuotesDoubleEscaped, splitter.SingleQuotesDoubleEscaped, splitter.Parenthesis, splitter.SquareBrackets, splitter.CurlyBrackets}
var oasTagSplitter = splitter.MustCreateSplitter(',', splitterEncs...)
var oasColonSplitter = splitter.MustCreateSplitter(':', splitterEncs...)
