package chioas

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestWriteSchemaRef(t *testing.T) {
	testCases := []struct {
		ref     string
		isArray bool
		expect  string
	}{
		{
			ref: "foo",
			expect: `$ref: "#/components/schemas/foo"
`,
		},
		{
			ref:     "foo",
			isArray: true,
			expect: `type: array
items:
  $ref: "#/components/schemas/foo"
`,
		},
		{
			ref: "some/uri",
			expect: `$ref: "some/uri"
`,
		},
		{
			ref:     "some/uri",
			isArray: true,
			expect: `type: array
items:
  $ref: "some/uri"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			writeSchemaRef(tc.ref, tc.isArray, w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}

func TestWriteRef_Errors(t *testing.T) {
	w := yaml.NewWriter(nil)
	w.RefChecker(&testRefErrorringChecker{})
	writeRef("foo", "bar", w)
	_, err := w.Bytes()
	assert.Error(t, err)

	w = yaml.NewWriter(nil)
	w.RefChecker(&testRefErrorringChecker{})
	writeRef("", "/my-schema", w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, tagNameRef+": \"/my-schema\"\n", string(data))

	w = yaml.NewWriter(nil)
	w.RefChecker(&testRefErrorringChecker{})
	writeRef("", refComponentsPrefix+tagNameSchemas+"/my-schema", w)
	_, err = w.Bytes()
	assert.Error(t, err)

	w = yaml.NewWriter(nil)
	writeRef("foo", "bar", w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, tagNameRef+": \""+refComponentsPrefix+"foo/bar\"\n", string(data))
}

func TestWriteItemRef_Errors(t *testing.T) {
	w := yaml.NewWriter(nil)
	w.RefChecker(&testRefErrorringChecker{})
	writeItemRef("foo", "bar", w)
	_, err := w.Bytes()
	assert.Error(t, err)

	w = yaml.NewWriter(nil)
	w.RefChecker(&testRefErrorringChecker{})
	writeItemRef("", "/my-schema", w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "- "+tagNameRef+": \"/my-schema\"\n", string(data))

	w = yaml.NewWriter(nil)
	w.RefChecker(&testRefErrorringChecker{})
	writeItemRef("", refComponentsPrefix+tagNameSchemas+"/my-schema", w)
	_, err = w.Bytes()
	assert.Error(t, err)

	w = yaml.NewWriter(nil)
	writeItemRef("foo", "bar", w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "- "+tagNameRef+": \""+refComponentsPrefix+"foo/bar\"\n", string(data))
}

func TestDefinition_RefCheck(t *testing.T) {
	defWithComponents := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "good",
				},
			},
			Requests: CommonRequests{
				"good": {},
			},
			Responses: CommonResponses{
				"good": {},
			},
			Parameters: CommonParameters{
				"good": {},
			},
			Examples: Examples{
				{
					Name: "good",
				},
			},
		},
	}
	defWithoutComponents := &Definition{}

	err := defWithComponents.RefCheck(tagNameSchemas, "good")
	assert.NoError(t, err)
	err = defWithComponents.RefCheck(tagNameSchemas, "bad")
	assert.Error(t, err)
	err = defWithoutComponents.RefCheck(tagNameSchemas, "good")
	assert.Error(t, err)

	err = defWithComponents.RefCheck(tagNameRequestBodies, "good")
	assert.NoError(t, err)
	err = defWithComponents.RefCheck(tagNameRequestBodies, "bad")
	assert.Error(t, err)
	err = defWithoutComponents.RefCheck(tagNameRequestBodies, "good")
	assert.Error(t, err)

	err = defWithComponents.RefCheck(tagNameResponses, "good")
	assert.NoError(t, err)
	err = defWithComponents.RefCheck(tagNameResponses, "bad")
	assert.Error(t, err)
	err = defWithoutComponents.RefCheck(tagNameResponses, "good")
	assert.Error(t, err)

	err = defWithComponents.RefCheck(tagNameParameters, "good")
	assert.NoError(t, err)
	err = defWithComponents.RefCheck(tagNameParameters, "bad")
	assert.Error(t, err)
	err = defWithoutComponents.RefCheck(tagNameParameters, "good")
	assert.Error(t, err)

	err = defWithComponents.RefCheck(tagNameExamples, "good")
	assert.NoError(t, err)
	err = defWithComponents.RefCheck(tagNameExamples, "bad")
	assert.Error(t, err)
	err = defWithoutComponents.RefCheck(tagNameExamples, "good")
	assert.Error(t, err)

	err = defWithComponents.RefCheck("unknown", "good")
	assert.Error(t, err)
	err = defWithComponents.RefCheck("unknown", "bad")
	assert.Error(t, err)
	err = defWithoutComponents.RefCheck("unknown", "good")
	assert.Error(t, err)
}

func TestDefinition_WithRefCheck(t *testing.T) {
	testCases := []struct {
		def          *Definition
		expectErr    bool
		expectErrMsg string
	}{
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
			},
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodPost: {
						Request: &Request{
							SchemaRef: "foo",
						},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/schemas/foo' invalid (definition has no components)`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodPost: {
						Request: &Request{
							SchemaRef: "foo",
						},
					},
				},
				Components: &Components{
					Schemas: Schemas{
						{
							Name: "bar",
						},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/schemas/foo' invalid`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodPost: {
						Request: &Request{
							SchemaRef: "foo",
						},
					},
				},
				Components: &Components{
					Schemas: Schemas{
						{
							Name: "foo",
						},
					},
				},
			},
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodGet: {
						Responses: Responses{
							200: {
								Ref: "foo",
							},
						},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/responses/foo' invalid (definition has no components)`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodGet: {
						Responses: Responses{
							200: {
								Ref: "foo",
							},
						},
					},
				},
				Components: &Components{
					Responses: CommonResponses{
						"bar": {},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/responses/foo' invalid`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodGet: {
						Responses: Responses{
							200: {
								Ref: "foo",
							},
						},
					},
				},
				Components: &Components{
					Responses: CommonResponses{
						"foo": {},
					},
				},
			},
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodPost: {
						Request: &Request{
							Ref: "foo",
						},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/requestBodies/foo' invalid (definition has no components)`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodPost: {
						Request: &Request{
							Ref: "foo",
						},
					},
				},
				Components: &Components{
					Requests: CommonRequests{
						"bar": {},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/requestBodies/foo' invalid`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodPost: {
						Request: &Request{
							Ref: "foo",
						},
					},
				},
				Components: &Components{
					Requests: CommonRequests{
						"foo": {},
					},
				},
			},
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodGet: {
						QueryParams: QueryParams{
							{
								Ref: "foo",
							},
						},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/parameters/foo' invalid (definition has no components)`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodGet: {
						QueryParams: QueryParams{
							{
								Ref: "foo",
							},
						},
					},
				},
				Components: &Components{
					Parameters: CommonParameters{
						"bar": {},
					},
				},
			},
			expectErr:    true,
			expectErrMsg: `$ref '#/components/parameters/foo' invalid`,
		},
		{
			def: &Definition{
				DocOptions: DocOptions{CheckRefs: true},
				Methods: Methods{
					http.MethodGet: {
						QueryParams: QueryParams{
							{
								Ref: "foo",
							},
						},
					},
				},
				Components: &Components{
					Parameters: CommonParameters{
						"foo": {},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			_, err := tc.def.AsYaml()
			if tc.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErrMsg, err.Error())
				tc.def.DocOptions.CheckRefs = false
				_, err = tc.def.AsYaml()
				assert.NoError(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRefCheck(t *testing.T) {
	testCases := []struct {
		area      string
		ref       string
		expect    string
		expectErr bool
	}{
		{
			expectErr: true,
		},
		{
			ref:    "/my-schema",
			expect: "/my-schema",
		},
		{
			ref:       refComponentsPrefix + tagNameSchemas + "/my-schema",
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			w.RefChecker(&testRefErrorringChecker{})
			actual := refCheck(tc.area, tc.ref, w)
			if tc.expectErr {
				assert.Error(t, w.Errored())
			} else {
				assert.NoError(t, w.Errored())
				assert.Equal(t, tc.expect, actual)
			}
		})
	}
}

func TestNeedsRefCheck(t *testing.T) {
	testCases := []struct {
		area       string
		ref        string
		expect     bool
		expectArea string
		expectRef  string
	}{
		{
			expect: true,
		},
		{
			ref:       "/my-schema.yaml",
			expectRef: "/my-schema.yaml",
		},
		{
			area:       tagNameSchemas,
			ref:        "my-schema",
			expect:     true,
			expectArea: tagNameSchemas,
			expectRef:  "my-schema",
		},
		{
			ref:        refComponentsPrefix + tagNameSchemas + "/my-schema",
			expect:     true,
			expectArea: tagNameSchemas,
			expectRef:  "my-schema",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			area, ref, check := needsRefCheck(tc.area, tc.ref)
			assert.Equal(t, tc.expect, check)
			assert.Equal(t, tc.expectArea, area)
			assert.Equal(t, tc.expectRef, ref)
		})
	}
}

type testRefErrorringChecker struct {
}

func (t *testRefErrorringChecker) RefCheck(area, ref string) error {
	return errors.New("fooey")
}
