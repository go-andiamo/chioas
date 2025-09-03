package chioas

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestDefinition_CheckRefs(t *testing.T) {
	testCases := []struct {
		definition   *Definition
		expectedErrs int
	}{
		{
			// nothing to check
			definition: &Definition{},
		},
		{
			// root method ok request ref
			definition: &Definition{
				Components: &Components{
					Requests: CommonRequests{
						"foo": Request{},
					},
				},
				Methods: Methods{
					http.MethodGet: {
						Request: &Request{
							Ref: "foo",
						},
					},
				},
			},
		},
		{
			// root method missing request ref
			definition: &Definition{
				Methods: Methods{
					http.MethodGet: {
						Request: &Request{
							Ref: "foo",
						},
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// paths ok
			definition: &Definition{
				Components: &Components{
					Requests: CommonRequests{
						"foo": Request{},
					},
				},
				Paths: Paths{
					"/api": {
						Paths: Paths{
							"/sub": {
								Methods: Methods{
									http.MethodGet: {
										Request: &Request{
											Ref: "foo",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// paths not ok
			definition: &Definition{
				Components: &Components{
					Requests: CommonRequests{
						"foo": Request{},
					},
				},
				Paths: Paths{
					"/api": {
						Paths: Paths{
							"/sub": {
								Methods: Methods{
									http.MethodGet: {
										Request: &Request{
											Ref: "bar",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.definition.CheckRefs()
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestMethod_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
			Requests: CommonRequests{
				"foo": {},
			},
			Responses: CommonResponses{
				"foo": {},
			},
			Parameters: CommonParameters{
				"foo": {},
			},
		},
	}
	testCases := []struct {
		method       Method
		expectedErrs int
	}{
		{
			//nothing to check
			method: Method{},
		},
		{
			//parameter ref not found
			method: Method{
				QueryParams: QueryParams{
					{
						Name: "param1",
						Ref:  "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// ok with params, request & responses
			method: Method{
				QueryParams: QueryParams{
					{
						Name: "param1",
						Ref:  "foo",
					},
				},
				Request: &Request{
					Ref: "foo",
				},
				Responses: Responses{
					200: {
						Ref: "foo",
					},
				},
			},
		},
		{
			// with bad params, request & responses
			method: Method{
				QueryParams: QueryParams{
					{
						Name: "param1",
						Ref:  "bar",
					},
				},
				Request: &Request{
					Ref: "bar",
				},
				Responses: Responses{
					200: {
						Ref: "bar",
					},
				},
			},
			expectedErrs: 3,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.method.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestComponents_checkRefs(t *testing.T) {
	testCases := []struct {
		components   *Components
		expectedErrs int
	}{
		{
			//nothing to check
			components: &Components{},
		},
		{
			// with schemas
			components: &Components{
				Schemas: Schemas{
					{
						Name: "foo",
					},
				},
			},
		},
		{
			// with cyclic schema ref
			components: &Components{
				Schemas: Schemas{
					{
						Name:      "foo",
						SchemaRef: "foo",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// with requests
			components: &Components{
				Requests: CommonRequests{
					"foo": {},
				},
			},
		},
		{
			// with request ok schema ref
			components: &Components{
				Schemas: Schemas{
					{
						Name: "foo",
					},
				},
				Requests: CommonRequests{
					"foo": {
						SchemaRef: "foo",
					},
				},
			},
		},
		{
			// with request not found schema ref
			components: &Components{
				Schemas: Schemas{
					{
						Name: "foo",
					},
				},
				Requests: CommonRequests{
					"foo": {
						SchemaRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// with responses
			components: &Components{
				Responses: CommonResponses{
					"foo": {},
				},
			},
		},
		{
			// with response ok schema ref
			components: &Components{
				Schemas: Schemas{
					{
						Name: "foo",
					},
				},
				Responses: CommonResponses{
					"foo": {
						SchemaRef: "foo",
					},
				},
			},
		},
		{
			// with response not found schema ref
			components: &Components{
				Schemas: Schemas{
					{
						Name: "foo",
					},
				},
				Responses: CommonResponses{
					"foo": {
						SchemaRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// with parameters
			components: &Components{
				Parameters: CommonParameters{
					"foo": {},
				},
			},
		},
		{
			// with examples
			components: &Components{
				Examples: Examples{
					{
						Name: "foo",
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.components.checkRefs(&Definition{Components: tc.components})
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestPath_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
			Parameters: CommonParameters{
				"foo": {},
			},
		},
	}
	testCases := []struct {
		path         Path
		expectedErrs int
	}{
		{
			//nothing to check
			path: Path{},
		},
		{
			//parameter ref ok
			path: Path{
				PathParams: PathParams{
					"foo": {
						Ref: "foo",
					},
				},
			},
		},
		{
			//parameter ref not found
			path: Path{
				PathParams: PathParams{
					"foo": {
						Ref: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			//full parameter ref ok
			path: Path{
				PathParams: PathParams{
					"foo": {
						Ref: refComponentsPrefix + tagNameParameters + "/foo",
					},
				},
			},
		},
		{
			//full parameter ref not found
			path: Path{
				PathParams: PathParams{
					"foo": {
						Ref: refComponentsPrefix + tagNameParameters + "/bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			//full parameter ref bad area
			path: Path{
				PathParams: PathParams{
					"foo": {
						Ref: refComponentsPrefix + tagNameSchemas + "/foo",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			//schema ref ok
			path: Path{
				PathParams: PathParams{
					"foo": {
						SchemaRef: "foo",
					},
				},
			},
		},
		{
			//full schema ref ok
			path: Path{
				PathParams: PathParams{
					"foo": {
						SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
					},
				},
			},
		},
		{
			//schema ref not found
			path: Path{
				PathParams: PathParams{
					"foo": {
						SchemaRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			//full schema ref not found
			path: Path{
				PathParams: PathParams{
					"foo": {
						SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			//full schema ref incorrect area
			path: Path{
				PathParams: PathParams{
					"foo": {
						SchemaRef: refComponentsPrefix + tagNameParameters + "/foo",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			//full schema ref not found + has both schema ref and schema
			path: Path{
				PathParams: PathParams{
					"foo": {
						Schema:    &Schema{},
						SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
					},
				},
			},
			expectedErrs: 2,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.path.checkRefs("", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestCommonParameter_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		param        CommonParameter
		expectedErrs int
	}{
		{
			//nothing to check
			param: CommonParameter{},
		},
		{
			//schema ref ok
			param: CommonParameter{
				SchemaRef: "foo",
			},
		},
		{
			//full schema ref ok
			param: CommonParameter{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
			},
		},
		{
			//schema ref not found
			param: CommonParameter{
				SchemaRef: "bar",
			},
			expectedErrs: 1,
		},
		{
			//full schema ref not found
			param: CommonParameter{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
			},
			expectedErrs: 1,
		},
		{
			//full schema ref incorrect area
			param: CommonParameter{
				SchemaRef: refComponentsPrefix + tagNameParameters + "/foo",
			},
			expectedErrs: 1,
		},
		{
			//full schema ref not found + has both schema ref and schema
			param: CommonParameter{
				Schema:    &Schema{},
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
			},
			expectedErrs: 2,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.param.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestQueryParams_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
			Parameters: CommonParameters{
				"foo": {},
			},
		},
	}
	testCases := []struct {
		params       QueryParams
		expectedErrs int
	}{
		{
			//nothing to check
			params: QueryParams{},
		},
		{
			//ref ok
			params: QueryParams{
				{
					Ref: "foo",
				},
			},
		},
		{
			//ref not found
			params: QueryParams{
				{
					Ref: "bar",
				},
			},
			expectedErrs: 1,
		},
		{
			//full ref ok
			params: QueryParams{
				{
					Ref: refComponentsPrefix + tagNameParameters + "/foo",
				},
			},
		},
		{
			//full ref not found
			params: QueryParams{
				{
					Ref: refComponentsPrefix + tagNameParameters + "/bar",
				},
			},
			expectedErrs: 1,
		},
		{
			//full ref bad area
			params: QueryParams{
				{
					Ref: refComponentsPrefix + tagNameSchemas + "/foo",
				},
			},
			expectedErrs: 1,
		},
		{
			//schema ref ok
			params: QueryParams{
				{
					SchemaRef: "foo",
				},
			},
		},
		{
			//full schema ref ok
			params: QueryParams{
				{
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
				},
			},
		},
		{
			//schema ref not found
			params: QueryParams{
				{
					SchemaRef: "bar",
				},
			},
			expectedErrs: 1,
		},
		{
			//full schema ref not found
			params: QueryParams{
				{
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
				},
			},
			expectedErrs: 1,
		},
		{
			//full schema ref incorrect area
			params: QueryParams{
				{
					SchemaRef: refComponentsPrefix + tagNameParameters + "/foo",
				},
			},
			expectedErrs: 1,
		},
		{
			//full schema ref not found + has both schema ref and schema
			params: QueryParams{
				{
					Schema:    &Schema{},
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
				},
			},
			expectedErrs: 2,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.params.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestRequest_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
			Requests: CommonRequests{
				"foo": {},
			},
			Examples: Examples{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		request      *Request
		expectedErrs int
	}{
		{
			// ref ok
			request: &Request{
				Ref: "foo",
			},
		},
		{
			// ref not found
			request: &Request{
				Ref: "bar",
			},
			expectedErrs: 1,
		},
		{
			// full ref ok
			request: &Request{
				Ref: refComponentsPrefix + tagNameRequestBodies + "/foo",
			},
		},
		{
			// full ref not found
			request: &Request{
				Ref: refComponentsPrefix + tagNameRequestBodies + "/bar",
			},
			expectedErrs: 1,
		},
		{
			// full ref invalid area
			request: &Request{
				Ref: refComponentsPrefix + tagNameSchemas + "/foo",
			},
			expectedErrs: 1,
		},
		{
			// schema ref ok
			request: &Request{
				SchemaRef: "/foo",
			},
		},
		{
			// schema ref not found
			request: &Request{
				SchemaRef: "bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref ok
			request: &Request{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
			},
		},
		{
			// full schema ref not found
			request: &Request{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref invalid area
			request: &Request{
				SchemaRef: refComponentsPrefix + tagNameExamples + "/foo",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref with schema
			request: &Request{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
				Schema:    &Schema{},
			},
			expectedErrs: 1,
		},
		{
			// full schema ref with schema
			request: &Request{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
				Schema:    Schema{},
			},
			expectedErrs: 1,
		},
		{
			// ok example
			request: &Request{
				Examples: Examples{
					{
						ExampleRef: "foo",
					},
				},
			},
		},
		{
			// example not found
			request: &Request{
				Examples: Examples{
					{
						ExampleRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// alt content type ok
			request: &Request{
				AlternativeContentTypes: ContentTypes{
					"text/csv": {
						SchemaRef: "foo",
					},
				},
			},
		},
		{
			// alt content type not found
			request: &Request{
				AlternativeContentTypes: ContentTypes{
					"text/csv": {
						SchemaRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.request.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestResponse_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
			Responses: CommonResponses{
				"foo": {},
			},
			Examples: Examples{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		response     Response
		expectedErrs int
	}{
		{
			// ref ok
			response: Response{
				Ref: "foo",
			},
		},
		{
			// ref not found
			response: Response{
				Ref: "bar",
			},
			expectedErrs: 1,
		},
		{
			// full ref ok
			response: Response{
				Ref: refComponentsPrefix + tagNameResponses + "/foo",
			},
		},
		{
			// full ref not found
			response: Response{
				Ref: refComponentsPrefix + tagNameResponses + "/bar",
			},
			expectedErrs: 1,
		},
		{
			// full ref invalid area
			response: Response{
				Ref: refComponentsPrefix + tagNameSchemas + "/foo",
			},
			expectedErrs: 1,
		},
		{
			// schema ref ok
			response: Response{
				SchemaRef: "/foo",
			},
		},
		{
			// schema ref not found
			response: Response{
				SchemaRef: "bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref ok
			response: Response{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
			},
		},
		{
			// full schema ref not found
			response: Response{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref invalid area
			response: Response{
				SchemaRef: refComponentsPrefix + tagNameExamples + "/foo",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref with schema
			response: Response{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
				Schema:    &Schema{},
			},
			expectedErrs: 1,
		},
		{
			// full schema ref with schema
			response: Response{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
				Schema:    Schema{},
			},
			expectedErrs: 1,
		},
		{
			// ok example
			response: Response{
				Examples: Examples{
					{
						ExampleRef: "foo",
					},
				},
			},
		},
		{
			// example not found
			response: Response{
				Examples: Examples{
					{
						ExampleRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// alt content type ok
			response: Response{
				AlternativeContentTypes: ContentTypes{
					"text/csv": {
						SchemaRef: "foo",
					},
				},
			},
		},
		{
			// alt content type not found
			response: Response{
				AlternativeContentTypes: ContentTypes{
					"text/csv": {
						SchemaRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.response.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestExamples_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Examples: Examples{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		examples     Examples
		expectedErrs int
	}{
		{
			//nothing to check
			examples: Examples{},
		},
		{
			//ref ok
			examples: Examples{
				{
					ExampleRef: "foo",
				},
			},
		},
		{
			//ref not found
			examples: Examples{
				{
					ExampleRef: "bar",
				},
			},
			expectedErrs: 1,
		},
		{
			//full ref ok
			examples: Examples{
				{
					ExampleRef: refComponentsPrefix + tagNameExamples + "/foo",
				},
			},
		},
		{
			//full ref not found
			examples: Examples{
				{
					ExampleRef: refComponentsPrefix + tagNameExamples + "/bar",
				},
			},
			expectedErrs: 1,
		},
		{
			//full ref bad area
			examples: Examples{
				{
					ExampleRef: refComponentsPrefix + tagNameSchemas + "/foo",
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.examples.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestContentTypes_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
			Examples: Examples{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		contentTypes ContentTypes
		expectedErrs int
	}{
		{
			//nothing to check
			contentTypes: ContentTypes{},
		},
		{
			// ref ok
			contentTypes: ContentTypes{
				"text/csv": {
					SchemaRef: "foo",
				},
			},
		},
		{
			// ref not found
			contentTypes: ContentTypes{
				"text/csv": {
					SchemaRef: "bar",
				},
			},
			expectedErrs: 1,
		},
		{
			// full ref ok
			contentTypes: ContentTypes{
				"text/csv": {
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
				},
			},
		},
		{
			// full ref not found
			contentTypes: ContentTypes{
				"text/csv": {
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
				},
			},
			expectedErrs: 1,
		},
		{
			// full ref not found and has schema
			contentTypes: ContentTypes{
				"text/csv": {
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
					Schema:    &Schema{},
				},
			},
			expectedErrs: 2,
		},
		{
			// full ref not found and has schema
			contentTypes: ContentTypes{
				"text/csv": {
					SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
					Schema:    Schema{},
				},
			},
			expectedErrs: 2,
		},
		{
			// example ref ok
			contentTypes: ContentTypes{
				"text/csv": {
					Examples: Examples{
						{
							ExampleRef: "foo",
						},
					},
				},
			},
		},
		{
			// example ref not found
			contentTypes: ContentTypes{
				"text/csv": {
					Examples: Examples{
						{
							ExampleRef: "bar",
						},
					},
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.contentTypes.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestSchema_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		schema       *Schema
		expectedErrs int
		seen         map[string]bool
	}{
		{
			// schema ref ok
			schema: &Schema{
				SchemaRef: "/foo",
			},
		},
		{
			// schema ref not found
			schema: &Schema{
				SchemaRef: "bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref ok
			schema: &Schema{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
			},
		},
		{
			// full schema ref not found
			schema: &Schema{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref invalid area
			schema: &Schema{
				SchemaRef: refComponentsPrefix + tagNameExamples + "/foo",
			},
			expectedErrs: 1,
		},
		{
			// cyclic
			schema: &Schema{
				SchemaRef: "foo",
			},
			seen: map[string]bool{
				refComponentsPrefix + tagNameSchemas + "/foo": true,
			},
			expectedErrs: 1,
		},
		{
			// non-cyclic property
			schema: &Schema{
				Properties: Properties{
					{
						Name: "sub-pty",
						Properties: Properties{
							{
								Name: "sub-sub-pty",
							},
						},
					},
				},
			},
		},
		{
			// cyclic property
			schema: &Schema{
				Properties: Properties{
					{
						Name: "sub-pty",
						Properties: Properties{
							{
								Name:      "sub-sub-pty",
								SchemaRef: "foo",
							},
						},
					},
				},
			},
			seen: map[string]bool{
				refComponentsPrefix + tagNameSchemas + "/foo": true,
			},
			expectedErrs: 1,
		},
		{
			// with discriminator ok
			schema: &Schema{
				Discriminator: &Discriminator{
					Mapping: map[string]string{
						"pty": "foo",
					},
				},
			},
		},
		{
			// with discriminator bad
			schema: &Schema{
				Discriminator: &Discriminator{
					Mapping: map[string]string{
						"pty": "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// with ofs ok
			schema: &Schema{
				Ofs: &Ofs{
					Of: []OfSchema{
						&Of{
							SchemaRef: "foo",
						},
					},
				},
			},
		},
		{
			// with ofs bad
			schema: &Schema{
				Ofs: &Ofs{
					Of: []OfSchema{
						&Of{
							SchemaRef: "bad",
						},
					},
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			seen := tc.seen
			if seen == nil {
				seen = make(map[string]bool)
			}
			errs := tc.schema.checkRefs("", "", d, seen)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestProperty_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		property     Property
		expectedErrs int
		seen         map[string]bool
	}{
		{
			// schema ref ok
			property: Property{
				SchemaRef: "/foo",
			},
		},
		{
			// schema ref not found
			property: Property{
				SchemaRef: "bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref ok
			property: Property{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
			},
		},
		{
			// full schema ref not found
			property: Property{
				SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
			},
			expectedErrs: 1,
		},
		{
			// full schema ref invalid area
			property: Property{
				SchemaRef: refComponentsPrefix + tagNameExamples + "/foo",
			},
			expectedErrs: 1,
		},
		{
			// non-cyclic
			property: Property{
				Properties: Properties{
					{
						Name: "sub-pty",
						Properties: Properties{
							{
								Name: "sub-sub-pty",
							},
						},
					},
				},
			},
		},
		{
			// cyclic
			property: Property{
				Properties: Properties{
					{
						Name: "sub-pty",
						Properties: Properties{
							{
								Name:      "sub-sub-pty",
								SchemaRef: "foo",
							},
						},
					},
				},
			},
			seen: map[string]bool{
				refComponentsPrefix + tagNameSchemas + "/foo": true,
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			seen := tc.seen
			if seen == nil {
				seen = make(map[string]bool)
			}
			errs := tc.property.checkRefs("", "", d, seen)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestDiscriminator_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		discriminator *Discriminator
		expectedErrs  int
	}{
		{
			// empty ok
			discriminator: &Discriminator{},
		},
		{
			// ref ok
			discriminator: &Discriminator{
				Mapping: map[string]string{
					"pty": "foo",
				},
			},
		},
		{
			// ref not found
			discriminator: &Discriminator{
				Mapping: map[string]string{
					"pty": "bar",
				},
			},
			expectedErrs: 1,
		},
		{
			// full ref ok
			discriminator: &Discriminator{
				Mapping: map[string]string{
					"pty": refComponentsPrefix + tagNameSchemas + "/foo",
				},
			},
		},
		{
			// full ref not found
			discriminator: &Discriminator{
				Mapping: map[string]string{
					"pty": refComponentsPrefix + tagNameSchemas + "/bar",
				},
			},
			expectedErrs: 1,
		},
		{
			// full ref invalid area
			discriminator: &Discriminator{
				Mapping: map[string]string{
					"pty": refComponentsPrefix + tagNameExamples + "/foo",
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.discriminator.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func TestOfs_checkRefs(t *testing.T) {
	d := &Definition{
		Components: &Components{
			Schemas: Schemas{
				{
					Name: "foo",
				},
			},
		},
	}
	testCases := []struct {
		ofs          *Ofs
		expectedErrs int
	}{
		{
			// empty ok
			ofs: &Ofs{},
		},
		{
			// ref ok
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaRef: "foo",
					},
				},
			},
		},
		{
			// ref not found
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaRef: "bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// full ref ok
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaRef: refComponentsPrefix + tagNameSchemas + "/foo",
					},
				},
			},
		},
		{
			// full ref not found
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaRef: refComponentsPrefix + tagNameSchemas + "/bar",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// full ref invalid area
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaRef: refComponentsPrefix + tagNameExamples + "/foo",
					},
				},
			},
			expectedErrs: 1,
		},
		{
			// schema ok
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaDef: &Schema{
							SchemaRef: "foo",
						},
					},
				},
			},
		},
		{
			// schema not found
			ofs: &Ofs{
				Of: []OfSchema{
					&Of{
						SchemaDef: &Schema{
							SchemaRef: "bar",
						},
					},
				},
			},
			expectedErrs: 1,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			errs := tc.ofs.checkRefs("", "", d)
			assert.Equal(t, tc.expectedErrs, len(errs))
		})
	}
}

func Test_isInternalRef(t *testing.T) {
	testCases := []struct {
		ref        string
		defArea    string
		expectOk   bool
		expectRef  string
		expectArea string
		expectErr  bool
	}{
		{
			ref: "",
		},
		{
			ref: "/some/external/url",
		},
		{
			ref:       "foo",
			expectOk:  true,
			expectRef: "foo",
		},
		{
			ref:       refComponentsPrefix + "foo",
			expectErr: true,
		},
		{
			ref:        refComponentsPrefix + tagNameSchemas + "/foo",
			defArea:    tagNameSchemas,
			expectOk:   true,
			expectRef:  "foo",
			expectArea: tagNameSchemas,
		},
		{
			ref:       refComponentsPrefix + tagNameSchemas + "/foo",
			defArea:   tagNameParameters,
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			ref, area, ok, err := isInternalRef(tc.ref, tc.defArea)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectOk, ok)
				if tc.expectOk {
					assert.Equal(t, tc.expectRef, ref)
					assert.Equal(t, tc.expectArea, area)
				}
			}
		})
	}
}

func TestRefError_Error(t *testing.T) {
	err := &RefError{
		Msg: "foo",
	}
	require.Error(t, err)
	assert.Equal(t, "foo", err.Error())
}
