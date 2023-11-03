package chioas

import (
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestSchema_From_DocsExample(t *testing.T) {
	type MyRequest struct {
		GivenName    string   `json:"givenName" oas:"description:'Your first/given name',required,pattern:'^[A-Z][a-zA-Z]+$',example"`
		FamilyName   string   `json:"familyName" oas:"description:'Your family name/surname',required,pattern:'^[A-Z][a-zA-Z]+$',example"`
		Age          int      `oas:"name:age,required:true,example,#this is a comment,#\"this is another, with commas in it, comment\""`
		SiblingNames []string `oas:"name:siblings,$ref:'Siblings'"`
	}
	def := Definition{
		Components: &Components{
			Schemas: Schemas{
				(&Schema{
					Name:        "MyRequest",
					Description: "My request",
				}).Must(MyRequest{
					GivenName:  "Bilbo",
					FamilyName: "Baggins",
					Age:        50,
				}),
			},
		},
	}
	data, err := def.AsYaml()
	assert.NoError(t, err)
	const expect = `openapi: "3.0.3"
info:
  title: "API Documentation"
  version: "1.0.0"
paths:
components:
  schemas:
    "MyRequest":
      description: "My request"
      type: "object"
      required:
        - "givenName"
        - "familyName"
        - "age"
      properties:
        "givenName":
          description: "Your first/given name"
          type: "string"
          example: "Bilbo"
          pattern: "^[A-Z][a-zA-Z]+$"
        "familyName":
          description: "Your family name/surname"
          type: "string"
          example: "Baggins"
          pattern: "^[A-Z][a-zA-Z]+$"
        "age":
          #this is a comment
          #this is another, with commas in it, comment
          type: "integer"
          example: 50
        "siblings":
          type: "array"
          items:
            $ref: "#/components/schemas/Siblings"
`
	assert.Equal(t, expect, string(data))
}

func TestSchema_From_Errors(t *testing.T) {
	type badStruct struct {
		Foo string `oas:"bad-token"`
	}
	s := &Schema{}
	_, err := s.From(badStruct{})
	assert.Error(t, err)
	assert.Equal(t, "unknown oas tag token 'bad-token'", err.Error())

	assert.Panics(t, func() {
		_ = s.MustFrom(badStruct{})
	})
}

func TestSchemaMustFrom_Simple(t *testing.T) {
	s := &Schema{}
	s2 := s.MustFrom(struct {
		Foo string `oas:"required"`
	}{})
	assert.Equal(t, s, s2)
	assert.Equal(t, 1, len(s.Properties))
	assert.Equal(t, tagValueTypeString, s.Properties[0].Type)
	assert.True(t, s.Properties[0].Required)
	assert.Equal(t, 1, len(s.RequiredProperties))
	assert.Equal(t, "Foo", s.RequiredProperties[0])
}

type basicSchemaSample struct {
	Id       int    `oas:"name:ID,description:'Desc of id',type:number,format: int64,required,minimum:1,maximum:10,example"`
	Name     string `json:"name,omitempty" oas:"description:this is the desc,required:true,pattern:xyz,minLength:1,maxLength:10,x-something:foo,example"`
	Flag     *bool  `json:",omitempty"`
	DontShow any    `json:"-"`
	Hyphen   string `json:"-,"`
}

func TestSchemaFrom_Basic(t *testing.T) {
	s, err := (&Schema{Name: "test", Type: "?"}).From(basicSchemaSample{
		Id:   123,
		Name: "example name",
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	w := yaml.NewWriter(nil)
	s.writeYaml(true, w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"test":
  type: "object"
  required:
    - "ID"
    - "name"
  properties:
    "ID":
      description: "Desc of id"
      type: "number"
      example: 123
      format: "int64"
      maximum: 10
      minimum: 1
    "name":
      description: "this is the desc"
      type: "string"
      example: "example name"
      pattern: "xyz"
      maxLength: 10
      minLength: 1
      x-something: "foo"
    "Flag":
      type: "boolean"
    "-":
      type: "string"
`
	assert.Equal(t, expect, string(data))
}

func TestSchemaFrom_BasicPtr(t *testing.T) {
	s, err := (&Schema{Name: "test", Type: "?"}).From(&basicSchemaSample{
		Id:   123,
		Name: "example name",
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	w := yaml.NewWriter(nil)
	s.writeYaml(true, w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"test":
  type: "object"
  required:
    - "ID"
    - "name"
  properties:
    "ID":
      description: "Desc of id"
      type: "number"
      example: 123
      format: "int64"
      maximum: 10
      minimum: 1
    "name":
      description: "this is the desc"
      type: "string"
      example: "example name"
      pattern: "xyz"
      maxLength: 10
      minLength: 1
      x-something: "foo"
    "Flag":
      type: "boolean"
    "-":
      type: "string"
`
	assert.Equal(t, expect, string(data))
}

type complexSchemaSample struct {
	Id        int64       `oas:"name:ID,description:'Desc of id',type:number,required,minimum:1,maximum:10"`
	Name      string      `json:"name,omitempty" oas:"description:this is the desc,required:true,pattern:xyz,minLength:1,maxLength:10"`
	Flag      *bool       `json:",omitempty"`
	DontShow  any         `json:"-"`
	Hyphen    string      `json:"-,"`
	Number    json.Number `json:"number"`
	Datetime  *time.Time  `json:"datetime"`
	Arr       []int       `json:"arr" oas:"$ref:Arr"`
	Subs      []subSample `json:"subs"`
	SubStruct subSample   `json:"subStruct"`
}

type subSample struct {
	Name  string  `json:"subName" oas:"required,example"`
	Value float32 `json:"subValue" oas:"description:\"this is the desc, with commas!\",minimum:0.1,maximum:99.9,required,example"`
}

func TestSchemaFrom_Complex(t *testing.T) {
	s, err := (&Schema{Name: "test", Type: "?"}).From(complexSchemaSample{
		Subs: []subSample{
			{
				Name:  "name eg (in arr)",
				Value: 1.23,
			},
		},
		SubStruct: subSample{
			Name:  "name eg",
			Value: 4.56,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	w := yaml.NewWriter(nil)
	s.writeYaml(true, w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"test":
  type: "object"
  required:
    - "ID"
    - "name"
  properties:
    "ID":
      description: "Desc of id"
      type: "number"
      maximum: 10
      minimum: 1
    "name":
      description: "this is the desc"
      type: "string"
      pattern: "xyz"
      maxLength: 10
      minLength: 1
    "Flag":
      type: "boolean"
    "-":
      type: "string"
    "number":
      type: "number"
    "datetime":
      type: "string"
      format: "date-time"
    "arr":
      type: "array"
      items:
        $ref: "#/components/schemas/Arr"
    "subs":
      type: "array"
      items:
        type: "object"
        properties:
          "subName":
            type: "string"
            example: "name eg (in arr)"
            required: true
          "subValue":
            description: "this is the desc, with commas!"
            type: "number"
            example: 1.23
            format: "float"
            required: true
            maximum: 99.9
            minimum: 0.1
    "subStruct":
      type: "object"
      properties:
        "subName":
          type: "string"
          example: "name eg"
          required: true
        "subValue":
          description: "this is the desc, with commas!"
          type: "number"
          example: 4.56
          format: "float"
          required: true
          maximum: 99.9
          minimum: 0.1
`
	assert.Equal(t, expect, string(data))
}

func TestSchemaFrom_BadTypes(t *testing.T) {
	testCases := []any{
		nil,
		"",
		true,
		1,
		map[string]any{},
		[]any{},
		reflect.TypeOf(""),
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			s, err := (&Schema{}).From(tc)
			assert.Error(t, err)
			assert.Equal(t, "sample must be a struct", err.Error())
			assert.Nil(t, s)
			assert.Panics(t, func() {
				_ = (&Schema{}).MustFrom(tc)
			})
		})
	}
}

func TestPropertyFrom(t *testing.T) {
	type SomeInterface interface {
		Foo()
	}
	testCases := []struct {
		value               any
		expectType          string
		expectFormat        string
		expectItemType      string
		expectSubProperties int
		expectError         string
	}{
		{
			value: struct {
				Test string
			}{},
			expectType: "string",
		},
		{
			value: struct {
				Test *string
			}{},
			expectType: "string",
		},
		{
			value: struct {
				Test json.Number
			}{},
			expectType: "number",
		},
		{
			value: struct {
				Test *json.Number
			}{},
			expectType: "number",
		},
		{
			value: struct {
				Test SomeInterface
			}{},
			expectType: "object",
		},
		{
			value: struct {
				Test map[string]any
			}{},
			expectType: "object",
		},
		{
			value: struct {
				Test bool
			}{},
			expectType: "boolean",
		},
		{
			value: struct {
				Test *bool
			}{},
			expectType: "boolean",
		},
		{
			value: struct {
				Test int
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test *int
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test uint
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test *uint
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test int8
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test *int8
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test uint8
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test *uint8
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test int16
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test *int16
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test uint16
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test *uint16
			}{},
			expectType: "integer",
		},
		{
			value: struct {
				Test int32
			}{},
			expectType:   "integer",
			expectFormat: "int32",
		},
		{
			value: struct {
				Test *int32
			}{},
			expectType:   "integer",
			expectFormat: "int32",
		},
		{
			value: struct {
				Test uint32
			}{},
			expectType:   "integer",
			expectFormat: "int32",
		},
		{
			value: struct {
				Test *uint32
			}{},
			expectType:   "integer",
			expectFormat: "int32",
		},
		{
			value: struct {
				Test int64
			}{},
			expectType:   "integer",
			expectFormat: "int64",
		},
		{
			value: struct {
				Test *int64
			}{},
			expectType:   "integer",
			expectFormat: "int64",
		},
		{
			value: struct {
				Test uint64
			}{},
			expectType:   "integer",
			expectFormat: "int64",
		},
		{
			value: struct {
				Test *uint64
			}{},
			expectType:   "integer",
			expectFormat: "int64",
		},
		{
			value: struct {
				Test float32
			}{},
			expectType:   "number",
			expectFormat: "float",
		},
		{
			value: struct {
				Test *float32
			}{},
			expectType:   "number",
			expectFormat: "float",
		},
		{
			value: struct {
				Test float64
			}{},
			expectType:   "number",
			expectFormat: "double",
		},
		{
			value: struct {
				Test *float64
			}{},
			expectType:   "number",
			expectFormat: "double",
		},
		{
			value: struct {
				Test time.Time
			}{},
			expectType:   "string",
			expectFormat: "date-time",
		},
		{
			value: struct {
				Test *time.Time
			}{},
			expectType:   "string",
			expectFormat: "date-time",
		},
		{
			value: struct {
				Test struct {
					Foo string
				}
			}{},
			expectType:          "object",
			expectSubProperties: 1,
		},
		{
			value: struct {
				Test *struct {
					Foo string
				}
			}{},
			expectType:          "object",
			expectSubProperties: 1,
		},
		{
			value: struct {
				Test []string
			}{},
			expectType:     "array",
			expectItemType: "string",
		},
		{
			value: struct {
				Test []*string
			}{},
			expectType:     "array",
			expectItemType: "string",
		},
		{
			value: struct {
				Test []bool
			}{},
			expectType:     "array",
			expectItemType: "boolean",
		},
		{
			value: struct {
				Test []*bool
			}{},
			expectType:     "array",
			expectItemType: "boolean",
		},
		{
			value: struct {
				Test []int
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*int
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []uint
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*uint
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []int8
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*int8
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []uint8
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*uint8
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []int16
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*int16
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []uint16
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*uint16
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []int32
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*int32
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []uint32
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*uint32
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []int64
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*int64
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []uint64
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []*uint64
			}{},
			expectType:     "array",
			expectItemType: "integer",
		},
		{
			value: struct {
				Test []float32
			}{},
			expectType:     "array",
			expectItemType: "number",
		},
		{
			value: struct {
				Test []*float32
			}{},
			expectType:     "array",
			expectItemType: "number",
		},
		{
			value: struct {
				Test []float64
			}{},
			expectType:     "array",
			expectItemType: "number",
		},
		{
			value: struct {
				Test []*float64
			}{},
			expectType:     "array",
			expectItemType: "number",
		},
		{
			value: struct {
				Test []time.Time
			}{},
			expectType:     "array",
			expectItemType: "string",
		},
		{
			value: struct {
				Test []*time.Time
			}{},
			expectType:     "array",
			expectItemType: "string",
		},
		{
			value: struct {
				Test []json.Number
			}{},
			expectType:     "array",
			expectItemType: "number",
		},
		{
			value: struct {
				Test []*json.Number
			}{},
			expectType:     "array",
			expectItemType: "number",
		},
		{
			value: struct {
				Test []map[string]any
			}{},
			expectType:     "array",
			expectItemType: "object",
		},
		{
			value: struct {
				Test []SomeInterface
			}{},
			expectType:     "array",
			expectItemType: "object",
		},
		{
			value: struct {
				Test []struct {
					Foo string
				}
			}{},
			expectType:          "array",
			expectItemType:      "object",
			expectSubProperties: 1,
		},
		{
			value: struct {
				Test []struct {
					Foo string
				}
			}{
				Test: []struct{ Foo string }{
					{
						Foo: "eg",
					},
				},
			},
			expectType:          "array",
			expectItemType:      "object",
			expectSubProperties: 1,
		},
		{
			value: struct {
				Test []*struct {
					Foo string
				}
			}{},
			expectType:          "array",
			expectItemType:      "object",
			expectSubProperties: 1,
		},
		// errors...
		{
			value: struct {
				Test string `oas:"bad-token"`
			}{},
			expectError: "unknown oas tag token 'bad-token'",
		},
		{
			value: struct {
				Test string `oas:"bad:token:value"`
			}{},
			expectError: "invalid oas tag token 'bad:token:value'",
		},
		{
			value: struct {
				Test string `oas:"unbalanced ) parenthesis"`
			}{},
			expectError: "unopened ')' at position 11",
		},
		{
			value: struct {
				Test string `oas:":"` // no token name or value
			}{},
			expectError: "oas must have a token name - ':'",
		},
		{
			value: struct {
				Test string `oas:":foo"` // no token name
			}{},
			expectError: "oas must have a token name - ':foo'",
		},
		{
			value: struct {
				Test string `oas:"foo:"` // no token value
			}{},
			expectError: "oas must have a token value - 'foo:'",
		},
		{
			value: struct {
				Test string `oas:"example:foo"` // example token is flag only!
			}{},
			expectError: "oas tag token 'example' - must not have a value (flag only)",
		},
		{
			value: struct {
				Test string `oas:"name"` // name cannot be use without value
			}{},
			expectError: "invalid oas tag token 'name' (missing value)",
		},
		{
			value: struct {
				Test string `oas:"required:foo"` // foo is not a boolean
			}{},
			expectError: "invalid oas token 'required' value 'foo'",
		},
		{
			value: struct {
				Test [][]string
			}{},
			expectError: "arrays of array not supported",
		},
		{
			value: struct {
				Test []struct {
					Foo string `oas:"bad-token"`
				}
			}{},
			expectError: "unknown oas tag token 'bad-token'",
		},
		{
			value: struct {
				Test struct {
					Foo string `oas:"bad-token"`
				}
			}{},
			expectError: "unknown oas tag token 'bad-token'",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			vt := reflect.TypeOf(tc.value)
			fld := vt.Field(0)
			vo := reflect.ValueOf(tc.value).FieldByName(fld.Name)
			pty, err := propertyFrom(fld, &vo)
			if tc.expectError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectType, pty.Type)
				assert.Equal(t, tc.expectFormat, pty.Format)
				assert.Equal(t, tc.expectItemType, pty.ItemType)
				assert.Equal(t, tc.expectSubProperties, len(pty.Properties))
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectError, err.Error())
			}
		})
	}
}

func TestSchema_From_WithExamples(t *testing.T) {
	type testSub struct {
		Buzz string `oas:"example"`
	}
	type testStruct struct {
		Foo string   `oas:"example"`
		Bar []string `oas:"example"`
		Baz testSub
	}
	s, err := (&Schema{}).From(testStruct{
		Foo: "foo example",
		Bar: []string{"bar example"},
		Baz: testSub{
			Buzz: "buzz example",
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, 3, len(s.Properties))
	pty := s.Properties[0]
	assert.Equal(t, "Foo", pty.Name)
	assert.Equal(t, tagValueTypeString, pty.Type)
	assert.Equal(t, "", pty.ItemType)
	assert.Equal(t, "foo example", pty.Example)
	pty = s.Properties[1]
	assert.Equal(t, "Bar", pty.Name)
	assert.Equal(t, tagValueTypeArray, pty.Type)
	assert.Equal(t, tagValueTypeString, pty.ItemType)
	assert.Equal(t, "bar example", pty.Example)
	pty = s.Properties[2]
	assert.Equal(t, "Baz", pty.Name)
	assert.Equal(t, tagValueTypeObject, pty.Type)
	assert.Equal(t, "", pty.ItemType)
	assert.Nil(t, pty.Example)
	assert.Equal(t, 1, len(pty.Properties))
	pty = pty.Properties[0]
	assert.Equal(t, "Buzz", pty.Name)
	assert.Equal(t, tagValueTypeString, pty.Type)
	assert.Equal(t, "", pty.ItemType)
	assert.Equal(t, "buzz example", pty.Example)
}

func TestTokenSetters(t *testing.T) {
	done := map[string]bool{}
	testCases := []struct {
		token       string
		value       string
		noValue     bool
		pty         *Property
		assert      func(pty *Property) bool
		expectError string
	}{
		{
			token: tagNameName,
			value: `foo`,
			assert: func(pty *Property) bool {
				return pty.Name == `foo`
			},
		},
		{
			token: tagNameName,
			value: `"foo"`,
			assert: func(pty *Property) bool {
				return pty.Name == `foo`
			},
		},
		{
			token: tagNameName,
			value: `'foo'`,
			assert: func(pty *Property) bool {
				return pty.Name == `foo`
			},
		},
		{
			token: tagNameDescription,
			value: `foo`,
			assert: func(pty *Property) bool {
				return pty.Description == `foo`
			},
		},
		{
			token: tagNameDescription,
			value: `"foo"`,
			assert: func(pty *Property) bool {
				return pty.Description == `foo`
			},
		},
		{
			token: tagNameDescription,
			value: `'foo'`,
			assert: func(pty *Property) bool {
				return pty.Description == `foo`
			},
		},
		{
			token: tagNameFormat,
			value: `foo`,
			assert: func(pty *Property) bool {
				return pty.Format == `foo`
			},
		},
		{
			token: tagNameFormat,
			value: `"foo"`,
			assert: func(pty *Property) bool {
				return pty.Format == `foo`
			},
		},
		{
			token: tagNameFormat,
			value: `'foo'`,
			assert: func(pty *Property) bool {
				return pty.Format == `foo`
			},
		},
		{
			token:       tagNameType,
			value:       `foo`,
			expectError: `invalid oas token 'type' value 'foo' (must be: ""|"string"|"object"|"array"|"boolean"|"integer"|"number"|"null")`,
		},
		{
			token: tagNameType,
			value: `string`,
			assert: func(pty *Property) bool {
				return pty.Type == `string`
			},
		},
		{
			token: tagNameType,
			value: `"string"`,
			assert: func(pty *Property) bool {
				return pty.Type == `string`
			},
		},
		{
			token: tagNameType,
			value: `'string'`,
			assert: func(pty *Property) bool {
				return pty.Type == `string`
			},
		},
		{
			token:       tagNameType,
			value:       `foo`,
			expectError: `invalid oas token 'type' value 'foo' (must be: ""|"string"|"object"|"array"|"boolean"|"integer"|"number"|"null")`,
		},
		{
			token: tagNameType,
			value: `string`,
			assert: func(pty *Property) bool {
				return pty.Type == `string`
			},
		},
		{
			token: tagNameType,
			value: `"string"`,
			assert: func(pty *Property) bool {
				return pty.Type == `string`
			},
		},
		{
			token: tagNameType,
			value: `'string'`,
			assert: func(pty *Property) bool {
				return pty.Type == `string`
			},
		},
		{
			token:       tagNameItemType,
			value:       `foo`,
			expectError: `invalid oas token 'itemType' value 'foo' (must be: ""|"string"|"object"|"array"|"boolean"|"integer"|"number"|"null")`,
		},
		{
			token: tagNameItemType,
			value: `string`,
			assert: func(pty *Property) bool {
				return pty.ItemType == `string`
			},
		},
		{
			token: tagNameItemType,
			value: `"string"`,
			assert: func(pty *Property) bool {
				return pty.ItemType == `string`
			},
		},
		{
			token: tagNameItemType,
			value: `'string'`,
			assert: func(pty *Property) bool {
				return pty.ItemType == `string`
			},
		},
		{
			token:   tagNameRequired,
			noValue: true,
			assert: func(pty *Property) bool {
				return pty.Required
			},
		},
		{
			token: tagNameRequired,
			value: "true",
			assert: func(pty *Property) bool {
				return pty.Required
			},
		},
		{
			token: tagNameRequired,
			value: "false",
			pty:   &Property{Required: true},
			assert: func(pty *Property) bool {
				return !pty.Required
			},
		},
		{
			token:       tagNameRequired,
			value:       "blah",
			expectError: "invalid oas token 'required' value 'blah'",
		},
		{
			token:   tagNameDeprecated,
			noValue: true,
			assert: func(pty *Property) bool {
				return pty.Deprecated
			},
		},
		{
			token: tagNameDeprecated,
			value: "true",
			assert: func(pty *Property) bool {
				return pty.Deprecated
			},
		},
		{
			token: tagNameDeprecated,
			value: "false",
			pty:   &Property{Deprecated: true},
			assert: func(pty *Property) bool {
				return !pty.Deprecated
			},
		},
		{
			token:       tagNameDeprecated,
			value:       "blah",
			expectError: "invalid oas token 'deprecated' value 'blah'",
		},
		{
			token: tagNameRef,
			value: `foo`,
			assert: func(pty *Property) bool {
				return pty.SchemaRef == `foo`
			},
		},
		{
			token: tagNameRef,
			value: `"foo"`,
			assert: func(pty *Property) bool {
				return pty.SchemaRef == `foo`
			},
		},
		{
			token: tagNameRef,
			value: `'foo'`,
			assert: func(pty *Property) bool {
				return pty.SchemaRef == `foo`
			},
		},
		{
			token: tagNamePattern,
			value: `foo`,
			assert: func(pty *Property) bool {
				return pty.Constraints.Pattern == `foo`
			},
		},
		{
			token: tagNamePattern,
			value: `"foo"`,
			assert: func(pty *Property) bool {
				return pty.Constraints.Pattern == `foo`
			},
		},
		{
			token: tagNamePattern,
			value: `'foo'`,
			assert: func(pty *Property) bool {
				return pty.Constraints.Pattern == `foo`
			},
		},
		{
			token: tagNameMaximum,
			value: `1`,
			assert: func(pty *Property) bool {
				return pty.Constraints.Maximum.String() == `1`
			},
		},
		{
			token:       tagNameMaximum,
			value:       `foo`,
			expectError: "invalid oas token 'maximum' value 'foo'",
		},
		{
			token: tagNameMinimum,
			value: `1`,
			assert: func(pty *Property) bool {
				return pty.Constraints.Minimum.String() == `1`
			},
		},
		{
			token:       tagNameMinimum,
			value:       `foo`,
			expectError: "invalid oas token 'minimum' value 'foo'",
		},
		{
			token:   tagNameExclusiveMaximum,
			noValue: true,
			assert: func(pty *Property) bool {
				return pty.Constraints.ExclusiveMaximum
			},
		},
		{
			token: tagNameExclusiveMaximum,
			value: "true",
			assert: func(pty *Property) bool {
				return pty.Constraints.ExclusiveMaximum
			},
		},
		{
			token: tagNameExclusiveMaximum,
			value: "false",
			pty:   &Property{Constraints: Constraints{ExclusiveMaximum: true}},
			assert: func(pty *Property) bool {
				return !pty.Constraints.ExclusiveMaximum
			},
		},
		{
			token:       tagNameExclusiveMaximum,
			value:       "blah",
			expectError: "invalid oas token 'exclusiveMaximum' value 'blah'",
		},
		{
			token:   tagNameExclusiveMinimum,
			noValue: true,
			assert: func(pty *Property) bool {
				return pty.Constraints.ExclusiveMinimum
			},
		},
		{
			token: tagNameExclusiveMinimum,
			value: "true",
			assert: func(pty *Property) bool {
				return pty.Constraints.ExclusiveMinimum
			},
		},
		{
			token: tagNameExclusiveMinimum,
			value: "false",
			pty:   &Property{Constraints: Constraints{ExclusiveMinimum: true}},
			assert: func(pty *Property) bool {
				return !pty.Constraints.ExclusiveMinimum
			},
		},
		{
			token:       tagNameExclusiveMinimum,
			value:       "blah",
			expectError: "invalid oas token 'exclusiveMinimum' value 'blah'",
		},
		{
			token:   tagNameNullable,
			noValue: true,
			assert: func(pty *Property) bool {
				return pty.Constraints.Nullable
			},
		},
		{
			token: tagNameNullable,
			value: "true",
			assert: func(pty *Property) bool {
				return pty.Constraints.Nullable
			},
		},
		{
			token: tagNameNullable,
			value: "false",
			pty:   &Property{Constraints: Constraints{Nullable: true}},
			assert: func(pty *Property) bool {
				return !pty.Constraints.ExclusiveMinimum
			},
		},
		{
			token:       tagNameNullable,
			value:       "blah",
			expectError: "invalid oas token 'nullable' value 'blah'",
		},
		{
			token:   tagNameUniqueItems,
			noValue: true,
			assert: func(pty *Property) bool {
				return pty.Constraints.UniqueItems
			},
		},
		{
			token: tagNameUniqueItems,
			value: "true",
			assert: func(pty *Property) bool {
				return pty.Constraints.UniqueItems
			},
		},
		{
			token: tagNameUniqueItems,
			value: "false",
			pty:   &Property{Constraints: Constraints{UniqueItems: true}},
			assert: func(pty *Property) bool {
				return !pty.Constraints.UniqueItems
			},
		},
		{
			token:       tagNameUniqueItems,
			value:       "blah",
			expectError: "invalid oas token 'uniqueItems' value 'blah'",
		},
		{
			token: tagNameMaxLength,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MaxLength == uint(1)
			},
		},
		{
			token:       tagNameMaxLength,
			value:       "blah",
			expectError: "invalid oas token 'maxLength' value 'blah'",
		},
		{
			token: tagNameMinLength,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MinLength == uint(1)
			},
		},
		{
			token:       tagNameMinLength,
			value:       "blah",
			expectError: "invalid oas token 'minLength' value 'blah'",
		},
		{
			token: tagNameMaxItems,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MaxItems == uint(1)
			},
		},
		{
			token:       tagNameMaxItems,
			value:       "blah",
			expectError: "invalid oas token 'maxItems' value 'blah'",
		},
		{
			token: tagNameMinItems,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MinItems == uint(1)
			},
		},
		{
			token:       tagNameMinItems,
			value:       "blah",
			expectError: "invalid oas token 'minItems' value 'blah'",
		},
		{
			token: tagNameMaxProperties,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MaxProperties == uint(1)
			},
		},
		{
			token:       tagNameMaxProperties,
			value:       "blah",
			expectError: "invalid oas token 'maxProperties' value 'blah'",
		},
		{
			token: tagNameMinProperties,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MinProperties == uint(1)
			},
		},
		{
			token:       tagNameMinProperties,
			value:       "blah",
			expectError: "invalid oas token 'minProperties' value 'blah'",
		},
		{
			token: tagNameMultipleOf,
			value: "1",
			assert: func(pty *Property) bool {
				return pty.Constraints.MultipleOf == uint(1)
			},
		},
		{
			token:       tagNameMultipleOf,
			value:       "blah",
			expectError: "invalid oas token 'multipleOf' value 'blah'",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			pty := &Property{}
			if tc.pty != nil {
				pty = tc.pty
			}
			if sf, ok := tokenSetters[tc.token]; ok {
				err := sf(pty, tc.value, !tc.noValue)
				if tc.expectError == "" {
					assert.NoError(t, err)
					assert.True(t, tc.assert(pty))
					done[tc.token] = true
				} else {
					assert.Error(t, err)
					assert.Equal(t, tc.expectError, err.Error())
				}
			} else {
				t.Fatalf("unknown token '%s'", tc.token)
			}
		})
	}
	assert.Equal(t, len(tokenSetters), len(done))
}

func TestSetNameFromJsonTag(t *testing.T) {
	testCases := []struct {
		value      any
		expectName string
		expectOmit bool
	}{
		{
			value: struct {
				Foo string
			}{},
			expectName: "Foo",
		},
		{
			value: struct {
				Foo string `json:"foo"`
			}{},
			expectName: "foo",
		},
		{
			value: struct {
				Foo string `json:"-"`
			}{},
			expectOmit: true,
		},
		{
			value: struct {
				Foo string `json:"-,"`
			}{},
			expectName: "-",
		},
		{
			value: struct {
				Foo string `json:"foo,omitempty"`
			}{},
			expectName: "foo",
		},
		{
			value: struct {
				Foo string `json:"foo,"`
			}{},
			expectName: "foo",
		},
		{
			value: struct {
				Foo string `json:""`
			}{},
			expectName: "Foo",
		},
		{
			value: struct {
				Foo string `json:","`
			}{},
			expectName: "Foo",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			pty := &Property{}
			vt := reflect.TypeOf(tc.value)
			fld := vt.Field(0)
			keep := setNameFromJsonTag(pty, fld)
			if tc.expectOmit {
				assert.False(t, keep)
			} else {
				assert.True(t, keep)
				assert.Equal(t, tc.expectName, pty.Name)
			}
		})
	}
}
