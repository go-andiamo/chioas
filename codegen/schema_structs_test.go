package codegen

import (
	"bytes"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func TestGenerateSchemaStructs_Schema(t *testing.T) {
	def := chioas.Schema{
		Type:               "object",
		RequiredProperties: []string{"prop1"},
		Properties: chioas.Properties{
			{
				Name:      "prop1",
				SchemaRef: "#/components/schemas/foo",
			},
		},
	}
	options := SchemaStructOptions{
		KeepComponentProperties: true,
		OASTags:                 true,
		GoDoc:                   true,
		PublicStructs:           true,
		Components: &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "foo",
					Type: "object",
					Properties: chioas.Properties{
						{
							Name: "sub1",
							Type: "string",
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	err := GenerateSchemaStructs(def, &buf, options)
	require.NoError(t, err)
	fmt.Println(buf.String())
	const expect = `package api

// Schema schema struct
type Schema struct {
	Prop1 SchemaFoo ~json:"prop1" oas:"required,type:object"~
}

// SchemaFoo #/components/schemas/foo
type SchemaFoo struct {
	Sub1 *string ~json:"sub1" oas:"type:string"~
}

`
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())

	buf.Reset()
	err = GenerateSchemaStructs(&def, &buf, options)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
}

func TestGenerateSchemaStructs_Components(t *testing.T) {
	def := chioas.Components{
		Schemas: chioas.Schemas{
			{
				Name: "foo",
				Type: "object",
				Properties: chioas.Properties{
					{
						Name: "sub1",
						Type: "string",
					},
					{
						Name:      "sub2",
						SchemaRef: "#/components/schemas/bar",
					},
				},
			},
			{
				Name: "bar",
				Type: "object",
				Properties: chioas.Properties{
					{
						Name: "pty1",
						Type: "string",
					},
				},
			},
			{
				Name: "buzz",
				Type: "string",
			},
		},
		Requests: chioas.CommonRequests{
			"req1": {
				Schema: &chioas.Schema{
					Type:               "object",
					RequiredProperties: []string{"prop1"},
					Properties: chioas.Properties{
						{
							Name: "prop1",
							Type: "string",
						},
					},
				},
			},
			"req2": {
				SchemaRef: "#/components/schemas/foo",
			},
		},
		Responses: chioas.CommonResponses{
			"res1": {
				Schema: &chioas.Schema{
					Type:               "object",
					RequiredProperties: []string{"prop1"},
					Properties: chioas.Properties{
						{
							Name: "prop1",
							Type: "string",
						},
					},
				},
			},
			"res2": {
				SchemaRef: "#/components/schemas/foo",
			},
		},
	}
	t.Run("public", func(t *testing.T) {
		options := SchemaStructOptions{
			OASTags:       true,
			GoDoc:         true,
			PublicStructs: true,
		}
		var buf bytes.Buffer
		err := GenerateSchemaStructs(def, &buf, options)
		require.NoError(t, err)
		fmt.Println(buf.String())
		const expect = `package api

// RequestReq1 request #/components/requestBodies/req1
type RequestReq1 struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// RequestReq2 request #/components/requestBodies/req2
type RequestReq2 SchemaFoo

// ResponseRes1 response #/components/responses/res1
type ResponseRes1 struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// ResponseRes2 response #/components/responses/res2
type ResponseRes2 SchemaFoo

// SchemaBar #/components/schemas/bar
type SchemaBar struct {
	Pty1 *string ~json:"pty1" oas:"type:string"~
}

// SchemaFoo #/components/schemas/foo
type SchemaFoo struct {
	Sub1 *string ~json:"sub1" oas:"type:string"~
	Sub2 *SchemaBar ~json:"sub2" oas:"type:object"~
}

`
		require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())

		buf.Reset()
		err = GenerateSchemaStructs(&def, &buf, options)
		require.NoError(t, err)
		require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
	})
	t.Run("private", func(t *testing.T) {
		options := SchemaStructOptions{
			OASTags:       true,
			GoDoc:         true,
			PublicStructs: false,
		}
		var buf bytes.Buffer
		err := GenerateSchemaStructs(def, &buf, options)
		require.NoError(t, err)
		fmt.Println(buf.String())
		const expect = `package api

// requestReq1 request #/components/requestBodies/req1
type requestReq1 struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// requestReq2 request #/components/requestBodies/req2
type requestReq2 schemaFoo

// responseRes1 response #/components/responses/res1
type responseRes1 struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// responseRes2 response #/components/responses/res2
type responseRes2 schemaFoo

// schemaBar #/components/schemas/bar
type schemaBar struct {
	Pty1 *string ~json:"pty1" oas:"type:string"~
}

// schemaFoo #/components/schemas/foo
type schemaFoo struct {
	Sub1 *string ~json:"sub1" oas:"type:string"~
	Sub2 *schemaBar ~json:"sub2" oas:"type:object"~
}

`
		require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
	})
}

func TestGenerateSchemaStructs_Definition(t *testing.T) {
	def := chioas.Definition{
		Methods: chioas.Methods{
			http.MethodGet: {
				Responses: chioas.Responses{
					http.StatusOK: {
						Ref: "root",
					},
				},
			},
			http.MethodOptions: {},
		},
		Paths: chioas.Paths{
			"/foos": {
				Methods: chioas.Methods{
					http.MethodGet: {
						Responses: chioas.Responses{
							http.StatusOK: {
								Ref: "foo",
							},
						},
					},
					http.MethodPost: {
						Request: &chioas.Request{
							Ref: "addFoo",
						},
						Responses: chioas.Responses{
							http.StatusOK: {
								Ref: "foo",
							},
						},
					},
				},
			},
		},
		Components: &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name:               "root",
					Type:               "object",
					RequiredProperties: []string{"prop1"},
					Properties: chioas.Properties{
						{
							Name: "prop1",
							Type: "string",
						},
					},
				},
				{
					Name:               "foo",
					Type:               "object",
					RequiredProperties: []string{"prop2"},
					Properties: chioas.Properties{
						{
							Name: "prop2",
							Type: "string",
						},
					},
				},
			},
			Requests: chioas.CommonRequests{
				"addFoo": {
					SchemaRef: "#/components/schemas/foo",
				},
			},
			Responses: chioas.CommonResponses{
				"foo": {
					SchemaRef: "#/components/schemas/foo",
				},
				"root": {
					SchemaRef: "root",
				},
			},
		},
	}
	options := SchemaStructOptions{
		KeepComponentProperties: true,
		OASTags:                 true,
		GoDoc:                   true,
		PublicStructs:           true,
	}
	var buf bytes.Buffer
	err := GenerateSchemaStructs(def, &buf, options)
	require.NoError(t, err)
	fmt.Println(buf.String())
	const expect = `package api

// GetRootOkResponse response GET root
type GetRootOkResponse SchemaRoot

// GetFoosOkResponse response GET /foos
type GetFoosOkResponse SchemaFoo

// PostFoosRequest request POST /foos
type PostFoosRequest SchemaFoo

// PostFoosOkResponse response POST /foos
type PostFoosOkResponse SchemaFoo

// SchemaFoo #/components/schemas/foo
type SchemaFoo struct {
	Prop2 string ~json:"prop2" oas:"required,type:string"~
}

// SchemaRoot #/components/schemas/root
type SchemaRoot struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

`
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())

	buf.Reset()
	err = GenerateSchemaStructs(&def, &buf, options)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
}

func TestGenerateSchemaStructs_Method(t *testing.T) {
	def := chioas.Method{
		Request: &chioas.Request{
			Schema: &chioas.Schema{
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name: "prop1",
						Type: "string",
					},
				},
			},
		},
		Responses: chioas.Responses{
			http.StatusOK: {
				Schema: &chioas.Schema{
					Type:               "object",
					RequiredProperties: []string{"name"},
					Properties: chioas.Properties{
						{
							Name: "name",
							Type: "string",
						},
					},
				},
			},
			http.StatusBadRequest: {
				Schema: &chioas.Schema{
					Type:               "object",
					RequiredProperties: []string{"error"},
					Properties: chioas.Properties{
						{
							Name: "error",
							Type: "string",
						},
					},
				},
			},
		},
	}
	options := SchemaStructOptions{
		OASTags:       true,
		GoDoc:         true,
		PublicStructs: true,
	}
	var buf bytes.Buffer
	err := GenerateSchemaStructs(def, &buf, options)
	require.NoError(t, err)
	fmt.Println(buf.String())
	const expect = `package api

// MethodRequest request
type MethodRequest struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// MethodOkResponse response
type MethodOkResponse struct {
	Name string ~json:"name" oas:"required,type:string"~
}

// MethodBadRequestResponse response
type MethodBadRequestResponse struct {
	Error string ~json:"error" oas:"required,type:string"~
}

`
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())

	buf.Reset()
	err = GenerateSchemaStructs(&def, &buf, options)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
}

func TestGenerateSchemaStructs_Path(t *testing.T) {
	def := chioas.Path{
		Methods: chioas.Methods{
			http.MethodGet: {
				Responses: chioas.Responses{
					http.StatusOK: {
						Schema: &chioas.Schema{
							Type:               "object",
							RequiredProperties: []string{"prop1"},
							Properties: chioas.Properties{
								{
									Name: "prop1",
									Type: "string",
								},
							},
						},
					},
				},
			},
			http.MethodPost: {
				Request: &chioas.Request{
					Schema: &chioas.Schema{
						Type:               "object",
						RequiredProperties: []string{"prop1"},
						Properties: chioas.Properties{
							{
								Name: "prop2",
								Type: "string",
							},
						},
					},
				},
			},
		},
		Paths: chioas.Paths{
			"/subs": {
				Methods: chioas.Methods{
					http.MethodGet: {
						Responses: chioas.Responses{
							http.StatusOK: {
								Schema: &chioas.Schema{
									Type:               "object",
									RequiredProperties: []string{"prop1"},
									Properties: chioas.Properties{
										{
											Name: "prop3",
											Type: "string",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	options := SchemaStructOptions{
		OASTags:       true,
		GoDoc:         true,
		PublicStructs: true,
	}
	var buf bytes.Buffer
	err := GenerateSchemaStructs(def, &buf, options)
	require.NoError(t, err)
	fmt.Println(buf.String())
	const expect = `package api

// GetOkResponse response GET
type GetOkResponse struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// PostRequest request POST
type PostRequest struct {
	Prop2 *string ~json:"prop2" oas:"type:string"~
}

// GetSubsOkResponse response GET /subs
type GetSubsOkResponse struct {
	Prop3 *string ~json:"prop3" oas:"type:string"~
}

`
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())

	buf.Reset()
	err = GenerateSchemaStructs(&def, &buf, options)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
}

func TestGenerateSchemaStructs_Paths(t *testing.T) {
	def := chioas.Paths{
		"/foos": {
			Methods: chioas.Methods{
				http.MethodGet: {
					Responses: chioas.Responses{
						http.StatusOK: {
							Schema: &chioas.Schema{
								Type:               "object",
								RequiredProperties: []string{"prop1"},
								Properties: chioas.Properties{
									{
										Name: "prop1",
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
		},
		"/bars": {
			Methods: chioas.Methods{
				http.MethodGet: {
					Responses: chioas.Responses{
						http.StatusOK: {
							Schema: &chioas.Schema{
								Type:               "object",
								RequiredProperties: []string{"prop1"},
								Properties: chioas.Properties{
									{
										Name: "prop1",
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	options := SchemaStructOptions{
		OASTags:       true,
		GoDoc:         true,
		PublicStructs: true,
	}
	var buf bytes.Buffer
	err := GenerateSchemaStructs(def, &buf, options)
	require.NoError(t, err)
	fmt.Println(buf.String())
	const expect = `package api

// GetBarsOkResponse response GET /bars
type GetBarsOkResponse struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

// GetFoosOkResponse response GET /foos
type GetFoosOkResponse struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

`
	require.Equal(t, strings.ReplaceAll(expect, "~", "`"), buf.String())
}

func Test_generateSchemaStruct(t *testing.T) {
	testCases := []struct {
		schema  chioas.Schema
		info    *pathInfo
		options SchemaStructOptions
		expect  string
	}{
		{
			// not an object
			schema: chioas.Schema{},
			expect: ``,
		},
		{
			// empty
			schema: chioas.Schema{
				Type: "object",
			},
			expect: `type schema struct {
}

`,
		},
		{
			// empty public
			schema: chioas.Schema{
				Type: "object",
			},
			options: SchemaStructOptions{PublicStructs: true},
			expect: `type Schema struct {
}

`,
		},
		{
			// basic types
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name: "prop1",
						Type: "string",
					},
					{
						Name: "prop2",
						Type: "boolean",
					},
					{
						Name: "prop3",
						Type: "integer",
					},
					{
						Name: "prop4",
						Type: "number",
					},
					{
						Name: "prop5",
						Type: "null",
					},
				},
			},
			expect: `type test struct {
	Prop1 string ~json:"prop1"~
	Prop2 *bool ~json:"prop2"~
	Prop3 *int ~json:"prop3"~
	Prop4 *float64 ~json:"prop4"~
	Prop5 any ~json:"prop5"~
}

`,
		},
		{
			// oas tags
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:        "prop1",
						Type:        "string",
						Description: `this is "quoted"`,
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 string ~json:"prop1" oas:"required,type:string,description:this is \\\"quoted\\\""~
}

`,
		},
		{
			// object property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name: "prop1",
						Type: "object",
						Properties: chioas.Properties{
							{
								Name:     "sub1",
								Type:     "string",
								Required: true,
							},
						},
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 struct {
		Sub1 string ~json:"sub1" oas:"required,type:string"~
	} ~json:"prop1" oas:"required,type:object"~
}

`,
		},
		{
			// object property with schema $ref (to object)
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:               "foo",
							Type:               "object",
							RequiredProperties: []string{"sub1"},
							Properties: chioas.Properties{
								{
									Name:     "sub1",
									Type:     "string",
									Required: true,
								},
							},
						},
					},
				},
			},
			expect: `type test struct {
	Prop1 struct {
		Sub1 string ~json:"sub1" oas:"required,type:string"~
	} ~json:"prop1" oas:"required,type:object"~
}

`,
		},
		{
			// object property with schema $ref (to object - keep)
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags:                 true,
				KeepComponentProperties: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:               "foo",
							Type:               "object",
							RequiredProperties: []string{"sub1"},
							Properties: chioas.Properties{
								{
									Name:     "sub1",
									Type:     "string",
									Required: true,
								},
							},
						},
					},
				},
			},
			expect: `type test struct {
	Prop1 schemaFoo ~json:"prop1" oas:"required,type:object"~
}

`,
		},
		{
			// object property with schema $ref (to object - keep, public)
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags:                 true,
				PublicStructs:           true,
				KeepComponentProperties: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:               "foo",
							Type:               "object",
							RequiredProperties: []string{"sub1"},
							Properties: chioas.Properties{
								{
									Name:     "sub1",
									Type:     "string",
									Required: true,
								},
							},
						},
					},
				},
			},
			expect: `type Test struct {
	Prop1 SchemaFoo ~json:"prop1" oas:"required,type:object"~
}

`,
		},
		{
			// object property with schema $ref (to object - keep, not required, public)
			schema: chioas.Schema{
				Name: "Test",
				Type: "object",
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags:                 true,
				PublicStructs:           true,
				KeepComponentProperties: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:               "foo",
							Type:               "object",
							RequiredProperties: []string{"sub1"},
							Properties: chioas.Properties{
								{
									Name:     "sub1",
									Type:     "string",
									Required: true,
								},
							},
						},
					},
				},
			},
			expect: `type Test struct {
	Prop1 *SchemaFoo ~json:"prop1" oas:"type:object"~
}

`,
		},
		{
			// object property with schema $ref (to object - keep, array, public)
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags:                 true,
				PublicStructs:           true,
				KeepComponentProperties: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:               "foo",
							Type:               "array",
							RequiredProperties: []string{"sub1"},
							Properties: chioas.Properties{
								{
									Name:     "sub1",
									Type:     "string",
									Required: true,
								},
							},
						},
					},
				},
			},
			expect: `type Test struct {
	Prop1 []SchemaFoo ~json:"prop1" oas:"required,type:array,itemType:object"~
}

`,
		},
		{
			// object property with schema $ref (to string)
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name: "foo",
							Type: "string",
						},
					},
				},
			},
			expect: `type test struct {
	Prop1 string ~json:"prop1" oas:"required,type:string"~
}

`,
		},
		{
			// object property with unresolved schema $ref
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:      "prop1",
						SchemaRef: "#/components/schemas/foo",
					},
				},
			},
			options: SchemaStructOptions{
				OASTags:    true,
				Components: &chioas.Components{},
			},
			expect: `type test struct {
	// Prop1 - error: cannot resolve internal schema $ref: foo
}

`,
		},
		{
			// empty object property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name: "prop1",
						Type: "object",
					},
				},
			},
			expect: `type test struct {
	Prop1 struct {
		// no properties
	} ~json:"prop1"~
}

`,
		},
		{
			// array[string] property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:     "prop1",
						Type:     "array",
						ItemType: "string",
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 []string ~json:"prop1" oas:"required,type:array,itemType:string"~
}

`,
		},
		{
			// array[boolean] property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:     "prop1",
						Type:     "array",
						ItemType: "boolean",
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 []bool ~json:"prop1" oas:"required,type:array,itemType:boolean"~
}

`,
		},
		{
			// array[integer] property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:     "prop1",
						Type:     "array",
						ItemType: "integer",
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 []int ~json:"prop1" oas:"required,type:array,itemType:integer"~
}

`,
		},
		{
			// array[number] property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:     "prop1",
						Type:     "array",
						ItemType: "number",
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 []float64 ~json:"prop1" oas:"required,type:array,itemType:number"~
}

`,
		},
		{
			// array[object] property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:     "prop1",
						Type:     "array",
						ItemType: "object",
						Properties: chioas.Properties{
							{
								Name:     "sub1",
								Type:     "string",
								Required: true,
							},
						},
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 []struct {
		Sub1 string ~json:"sub1" oas:"required,type:string"~
	} ~json:"prop1" oas:"required,type:array,itemType:object"~
}

`,
		},
		{
			// array[unknown?] property
			schema: chioas.Schema{
				Name:               "Test",
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name:     "prop1",
						Type:     "array",
						ItemType: "null",
					},
				},
			},
			options: SchemaStructOptions{OASTags: true},
			expect: `type test struct {
	Prop1 []any ~json:"prop1" oas:"required,type:array,itemType:null"~
}

`,
		},
		{
			// $ref
			schema: chioas.Schema{
				SchemaRef: "#/components/schemas/foo",
			},
			options: SchemaStructOptions{
				OASTags: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:               "foo",
							Type:               "object",
							RequiredProperties: []string{"prop1"},
							Properties: chioas.Properties{
								{
									Name:   "prop1",
									Type:   "string",
									Format: "uuid",
								},
							},
						},
					},
				},
			},
			expect: `type schema struct {
	Prop1 string ~json:"prop1" oas:"required,type:string,format:uuid"~
}

`,
		},
		{
			// $ref $ref
			schema: chioas.Schema{
				SchemaRef: "#/components/schemas/foo",
			},
			options: SchemaStructOptions{
				OASTags: true,
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:      "foo",
							SchemaRef: "#/components/schemas/bar",
						},
						{
							Name:               "bar",
							Type:               "object",
							RequiredProperties: []string{"prop1"},
							Properties: chioas.Properties{
								{
									Name:   "prop1",
									Type:   "string",
									Format: "uuid",
								},
							},
						},
					},
				},
			},
			expect: `type schema struct {
	Prop1 string ~json:"prop1" oas:"required,type:string,format:uuid"~
}

`,
		},
		{
			// $ref no components
			schema: chioas.Schema{
				SchemaRef: "#/components/schemas/foo",
			},
			options: SchemaStructOptions{},
			expect: `type schema struct {
	// error - cannot resolve internal schema $ref (no components!): #/components/schemas/foo
}

`,
		},
		{
			// $ref external
			schema: chioas.Schema{
				SchemaRef: "/this/is/external",
			},
			options: SchemaStructOptions{},
			expect: `type schema struct {
	// error - cannot resolve external/invalid schema $ref: /this/is/external
}

`,
		},
		{
			// cyclic $ref (full)
			schema: chioas.Schema{
				SchemaRef: "#/components/schemas/foo",
			},
			options: SchemaStructOptions{
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:      "foo",
							SchemaRef: "#/components/schemas/foo",
						},
					},
				},
			},
			expect: `type schema struct {
	// error - cyclic schema $ref: foo
}

`,
		},
		{
			// cyclic $ref (short)
			schema: chioas.Schema{
				SchemaRef: "foo",
			},
			options: SchemaStructOptions{
				Components: &chioas.Components{
					Schemas: chioas.Schemas{
						{
							Name:      "foo",
							SchemaRef: "foo",
						},
					},
				},
			},
			expect: `type schema struct {
	// error - cyclic schema $ref: foo
}

`,
		},
		{
			// with go docs
			schema: chioas.Schema{
				Type: "object",
			},
			options: SchemaStructOptions{GoDoc: true},
			expect: `// schema schema struct
type schema struct {
}

`,
		},
		{
			// with go docs & info request
			schema: chioas.Schema{
				Type: "object",
			},
			info: &pathInfo{
				method:   http.MethodPost,
				path:     "/api/foos",
				infoType: infoTypeRequest,
			},
			options: SchemaStructOptions{GoDoc: true},
			expect: `// schema request POST /api/foos
type schema struct {
}

`,
		},
		{
			// with go docs & info response
			schema: chioas.Schema{
				Type: "object",
			},
			info: &pathInfo{
				method:   http.MethodGet,
				path:     "/api/foos",
				infoType: infoTypeResponse,
			},
			options: SchemaStructOptions{GoDoc: true},
			expect: `// schema response GET /api/foos
type schema struct {
}

`,
		},
		{
			// object property with sub-properties
			schema: chioas.Schema{
				Type:               "object",
				RequiredProperties: []string{"prop1"},
				Properties: chioas.Properties{
					{
						Name: "prop1",
						Type: "object",
						Properties: chioas.Properties{
							{
								Name:     "sub2",
								Type:     "string",
								Required: true,
							},
							{
								Name:     "sub1",
								Type:     "string",
								Required: true,
							},
						},
					},
				},
			},
			expect: `type schema struct {
	Prop1 struct {
		Sub1 string ~json:"sub1"~
		Sub2 string ~json:"sub2"~
	} ~json:"prop1"~
}

`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newStructsWriter(&buf, tc.options)
			generateSchemaStruct(tc.info, tc.schema, w)
			require.NoError(t, w.err)
			fmt.Println(buf.String())
			require.Equal(t, strings.ReplaceAll(tc.expect, "~", "`"), buf.String())
		})
	}
}

func TestStructsWriter_resolveRequest(t *testing.T) {
	t.Run("same", func(t *testing.T) {
		r := &chioas.Request{}
		w := newStructsWriter(nil, SchemaStructOptions{})
		r2, err := w.resolveRequest(r, nil)
		require.NoError(t, err)
		require.Equal(t, r, r2)
	})
	t.Run("found", func(t *testing.T) {
		r := &chioas.Request{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{
			Components: &chioas.Components{
				Requests: chioas.CommonRequests{
					"foo": chioas.Request{},
				},
			},
		})
		r2, err := w.resolveRequest(r, nil)
		require.NoError(t, err)
		require.NotEqual(t, r, r2)
	})
	t.Run("cyclic", func(t *testing.T) {
		r := &chioas.Request{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{
			Components: &chioas.Components{
				Requests: chioas.CommonRequests{
					"foo": chioas.Request{
						Ref: "foo",
					},
				},
			},
		})
		_, err := w.resolveRequest(r, nil)
		require.Error(t, err)
		require.Equal(t, "cyclic request $ref: foo", err.Error())
	})
	t.Run("not found", func(t *testing.T) {
		r := &chioas.Request{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{
			Components: &chioas.Components{},
		})
		_, err := w.resolveRequest(r, nil)
		require.Error(t, err)
		require.Equal(t, "cannot resolve internal request $ref: foo", err.Error())
	})
	t.Run("no components", func(t *testing.T) {
		r := &chioas.Request{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{})
		_, err := w.resolveRequest(r, nil)
		require.Error(t, err)
		require.Equal(t, "cannot resolve internal request $ref (no components!): foo", err.Error())
	})
	t.Run("bad ref", func(t *testing.T) {
		r := &chioas.Request{
			Ref: "bad/ref",
		}
		w := newStructsWriter(nil, SchemaStructOptions{})
		_, err := w.resolveRequest(r, nil)
		require.Error(t, err)
		require.Equal(t, "cannot resolve external/invalid request $ref: bad/ref", err.Error())
	})
}

func TestStructsWriter_resolveResponse(t *testing.T) {
	t.Run("same", func(t *testing.T) {
		r := &chioas.Response{}
		w := newStructsWriter(nil, SchemaStructOptions{})
		r2, err := w.resolveResponse(r, nil)
		require.NoError(t, err)
		require.Equal(t, r, r2)
	})
	t.Run("found", func(t *testing.T) {
		r := &chioas.Response{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{
			Components: &chioas.Components{
				Responses: chioas.CommonResponses{
					"foo": chioas.Response{},
				},
			},
		})
		r2, err := w.resolveResponse(r, nil)
		require.NoError(t, err)
		require.NotEqual(t, r, r2)
	})
	t.Run("cyclic", func(t *testing.T) {
		r := &chioas.Response{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{
			Components: &chioas.Components{
				Responses: chioas.CommonResponses{
					"foo": chioas.Response{
						Ref: "foo",
					},
				},
			},
		})
		_, err := w.resolveResponse(r, nil)
		require.Error(t, err)
		require.Equal(t, "cyclic response $ref: foo", err.Error())
	})
	t.Run("not found", func(t *testing.T) {
		r := &chioas.Response{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{
			Components: &chioas.Components{},
		})
		_, err := w.resolveResponse(r, nil)
		require.Error(t, err)
		require.Equal(t, "cannot resolve internal response $ref: foo", err.Error())
	})
	t.Run("no components", func(t *testing.T) {
		r := &chioas.Response{
			Ref: "foo",
		}
		w := newStructsWriter(nil, SchemaStructOptions{})
		_, err := w.resolveResponse(r, nil)
		require.Error(t, err)
		require.Equal(t, "cannot resolve internal response $ref (no components!): foo", err.Error())
	})
	t.Run("bad ref", func(t *testing.T) {
		r := &chioas.Response{
			Ref: "bad/ref",
		}
		w := newStructsWriter(nil, SchemaStructOptions{})
		_, err := w.resolveResponse(r, nil)
		require.Error(t, err)
		require.Equal(t, "cannot resolve external/invalid response $ref: bad/ref", err.Error())
	})
}

func TestStructsWriter_writePropertyTags(t *testing.T) {
	testCases := []struct {
		property chioas.Property
		required bool
		options  SchemaStructOptions
		expect   string
	}{
		{
			property: chioas.Property{
				Name: "pty",
			},
			expect: " `json:\"pty\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
			},
			expect: " `json:\"pty\" oas:\"\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name:        "pty",
				Type:        "string",
				Description: `this is "quoted"`,
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,type:string,description:this is \\\\\\\"quoted\\\\\\\"\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name:   "pty",
				Format: "uuid",
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,format:uuid\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name:     "pty",
				Type:     "array",
				ItemType: "string",
			},
			expect: " `json:\"pty\" oas:\"type:array,itemType:string\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Constraints: chioas.Constraints{
					Pattern: "^[a-z0-9]+$",
				},
			},
			expect: " `json:\"pty\" oas:\"pattern:\\\"^[a-z0-9]+$\\\"\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name:       "pty",
				Deprecated: true,
			},
			expect: " `json:\"pty\" oas:\"deprecated\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "integer",
				Constraints: chioas.Constraints{
					Minimum: "1",
					Maximum: "10",
				},
			},
			expect: " `json:\"pty\" oas:\"type:integer,maximum:10,minimum:1\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "integer",
				Constraints: chioas.Constraints{
					MultipleOf: 2,
				},
			},
			expect: " `json:\"pty\" oas:\"type:integer,multipleOf:2\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "string",
				Constraints: chioas.Constraints{
					MinLength: 2,
					MaxLength: 10,
				},
			},
			expect: " `json:\"pty\" oas:\"type:string,maxLength:10,minLength:2\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "array",
				Constraints: chioas.Constraints{
					MinItems:    2,
					MaxItems:    10,
					UniqueItems: true,
				},
			},
			expect: " `json:\"pty\" oas:\"type:array,maxItems:10,minItems:2,uniqueItems\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "object",
				Constraints: chioas.Constraints{
					MinProperties: 2,
					MaxProperties: 10,
				},
			},
			expect: " `json:\"pty\" oas:\"type:object,maxProperties:10,minProperties:2\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Constraints: chioas.Constraints{
					ExclusiveMinimum: true,
					ExclusiveMaximum: true,
				},
			},
			expect: " `json:\"pty\" oas:\"exclusiveMaximum,exclusiveMinimum\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Constraints: chioas.Constraints{
					Nullable: true,
				},
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,nullable\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "string",
				Enum: []any{nil, "a", "b"},
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,type:string,enum:[\\\"a\\\",\\\"b\\\"]\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "integer",
				Enum: []any{nil, 1, 2, 3},
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,type:integer,enum:[1,2,3]\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "number",
				Enum: []any{nil, 1.1, 2.2, float32(3.3)},
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,type:number,enum:[1.1,2.2,3.3]\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Type: "boolean",
				Enum: []any{nil, false, true},
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,type:boolean,enum:[false,true]\"`",
		},
		{
			options: SchemaStructOptions{OASTags: true},
			property: chioas.Property{
				Name: "pty",
				Extensions: chioas.Extensions{
					"foo":  "bar",
					"bar":  1,
					"baz":  1.1,
					"baz2": float32(2.2),
					"buzz": true,
				},
			},
			required: true,
			expect:   " `json:\"pty\" oas:\"required,x-bar:1,x-baz:1.1,x-baz2:2.2,x-buzz:true,x-foo:\\\"bar\\\"\"`",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newStructsWriter(&buf, tc.options)
			w.writePropertyTags(tc.property, tc.property.Type, tc.property.ItemType, tc.required)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func TestStructsWriter_writePrologue(t *testing.T) {
	t.Run("default package", func(t *testing.T) {
		var buf bytes.Buffer
		w := newStructsWriter(&buf, SchemaStructOptions{})
		w.writePrologue()
		require.NoError(t, w.err)
		const expect = `package api

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("override package", func(t *testing.T) {
		var buf bytes.Buffer
		w := newStructsWriter(&buf, SchemaStructOptions{Package: "foo"})
		w.writePrologue()
		require.NoError(t, w.err)
		const expect = `package foo

`
		require.Equal(t, expect, buf.String())
	})
}

func Test_getSchema(t *testing.T) {
	s := getSchema(nil)
	require.Nil(t, s)
	s = getSchema(struct{}{})
	require.Nil(t, s)
	s = getSchema(&chioas.Schema{})
	require.NotNil(t, s)
	s = getSchema(chioas.Schema{})
	require.NotNil(t, s)
}

func Test_statusCodeName(t *testing.T) {
	s := statusCodeName(http.StatusOK)
	require.Equal(t, "Ok", s)
	s = statusCodeName(299)
	require.Equal(t, "299", s)
}
