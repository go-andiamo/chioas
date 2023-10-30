package chioas

import (
	"errors"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-andiamo/urit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMethod_GetHandler_WithHandlerSet(t *testing.T) {
	m := Method{
		Handler: func(writer http.ResponseWriter, request *http.Request) {},
	}
	mh, err := m.getHandler(root, http.MethodGet, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	var hf http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {}
	m = Method{
		Handler: hf,
	}
	mh, err = m.getHandler(root, http.MethodGet, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)
}

func TestMethod_GetHandler_WithGetHandler(t *testing.T) {
	m := Method{
		Handler: func(path string, method string, thisApi any) (http.HandlerFunc, error) {
			return nil, errors.New("foo")
		},
	}
	mh, err := m.getHandler(root, http.MethodGet, nil)
	assert.Error(t, err)
	assert.Nil(t, mh)
}

func TestMethod_GetHandler_WithMethodNameSet(t *testing.T) {
	m := Method{
		Handler: "Foo",
	}
	_, err := m.getHandler(root, http.MethodGet, nil)
	assert.Error(t, err)

	type dummyStruct struct{}
	_, err = m.getHandler(root, http.MethodGet, &dummyStruct{})
	assert.Error(t, err)

	dummy := &dummyWithMethod{}
	mh, err := m.getHandler(root, http.MethodGet, dummy)
	assert.NoError(t, err)
	assert.NotNil(t, mh)
	assert.False(t, dummy.called)
	req, err := http.NewRequest(http.MethodGet, "/example", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	mh.ServeHTTP(res, req)
	assert.True(t, dummy.called)

	m = Method{
		Handler: "Bar",
	}
	_, err = m.getHandler(root, http.MethodGet, dummy)
	assert.Error(t, err)

	m = Method{
		Handler: false,
	}
	_, err = m.getHandler(root, http.MethodGet, dummy)
	assert.Error(t, err)
}

type dummyWithMethod struct {
	called bool
}

func (d *dummyWithMethod) Foo(writer http.ResponseWriter, request *http.Request) {
	d.called = true
}

func (d *dummyWithMethod) Bar() {
	d.called = true
}

func TestMethod_GetHandler_NoHandlerOrMethodNameSet(t *testing.T) {
	m := Method{}
	_, err := m.getHandler(root, http.MethodGet, nil)
	assert.Error(t, err)
}

func TestMethods_WriteYaml(t *testing.T) {
	opts := &DocOptions{}
	ms := Methods{
		http.MethodHead:    {},
		http.MethodPost:    {},
		http.MethodPut:     {},
		http.MethodTrace:   {},
		http.MethodPatch:   {},
		http.MethodDelete:  {},
		http.MethodOptions: {},
		http.MethodGet:     {},
	}
	w := yaml.NewWriter(nil)
	ms.writeYaml(opts, false, nil, nil, "", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
head:
  responses:
    200:
      description: "OK"
options:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
post:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
put:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
patch:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
delete:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
trace:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestMethods_HasVisibleMethods(t *testing.T) {
	opts := &DocOptions{
		HideHeadMethods: true,
	}
	ms := Methods{}
	assert.False(t, ms.hasVisibleMethods(opts))

	ms = Methods{
		http.MethodHead: {},
	}
	assert.False(t, ms.hasVisibleMethods(opts))

	ms = Methods{
		http.MethodHead: {},
		http.MethodGet:  {},
	}
	assert.True(t, ms.hasVisibleMethods(opts))

	ms = Methods{
		http.MethodHead: {},
		http.MethodGet:  {HideDocs: true},
	}
	assert.False(t, ms.hasVisibleMethods(opts))
}

func TestMethods_WriteYaml_HideMethod(t *testing.T) {
	opts := &DocOptions{}
	ms := Methods{
		http.MethodHead: {HideDocs: true},
		http.MethodGet:  {},
	}
	w := yaml.NewWriter(nil)
	ms.writeYaml(opts, false, nil, nil, "", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestMethods_WriteYaml_AutoHead(t *testing.T) {
	opts := &DocOptions{}
	ms := Methods{
		http.MethodGet: {},
	}
	w := yaml.NewWriter(nil)
	ms.writeYaml(opts, true, nil, nil, "", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
head:
  responses:
    200:
      description: "OK"
`
	assert.Equal(t, expect, string(data))
}

func TestMethod_WriteYaml(t *testing.T) {
	opts := &DocOptions{}
	m := Method{
		Description: "test desc",
		Summary:     "test summary",
		OperationId: "testOp",
		Tag:         "",
		Request: &Request{
			SchemaRef: "foo",
		},
		Responses: Responses{
			http.StatusCreated: {
				SchemaRef: "foo",
			},
		},
		Additional: &testAdditional{},
		Comment:    "test comment",
	}
	w := yaml.NewWriter(nil)
	m.writeYaml(opts, nil, nil, nil, "foo", http.MethodPost, w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `post:
  #test comment
  summary: "test summary"
  description: "test desc"
  operationId: "testOp"
  tags:
    - "foo"
  requestBody:
    required: false
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/foo"
  responses:
    201:
      description: "Created"
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/foo"
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}

func TestMethod_WriteYaml_WithDefaultResponses(t *testing.T) {
	opts := &DocOptions{
		DefaultResponses: Responses{
			http.StatusOK: {},
		},
	}
	m := Method{
		Description: "test desc",
		Summary:     "test summary",
		OperationId: "testOp",
	}
	w := yaml.NewWriter(nil)
	m.writeYaml(opts, nil, nil, nil, "foo", http.MethodGet, w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  summary: "test summary"
  description: "test desc"
  operationId: "testOp"
  tags:
    - "foo"
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestMethod_WriteYaml_WithOperationIdentifier(t *testing.T) {
	opts := &DocOptions{
		OperationIdentifier: func(method Method, methodName string, path string, parentTag string) string {
			return methodName + strings.ReplaceAll(path, "/", "_")
		},
		DefaultResponses: Responses{
			http.StatusOK: {},
		},
	}
	m := Method{
		Description: "test desc",
		Summary:     "test summary",
		OperationId: "testOp",
	}
	pathTemplate := urit.MustCreateTemplate("/root/foo")
	w := yaml.NewWriter(nil)
	m.writeYaml(opts, pathTemplate, nil, nil, "foo", http.MethodGet, w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  summary: "test summary"
  description: "test desc"
  operationId: "GET_root_foo"
  tags:
    - "foo"
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestMethod_WriteYaml_Deprecated(t *testing.T) {
	opts := &DocOptions{}
	m := Method{
		Description: "test desc",
		Deprecated:  true,
	}
	pathTemplate := urit.MustCreateTemplate("/root/foo")
	w := yaml.NewWriter(nil)
	m.writeYaml(opts, pathTemplate, nil, nil, "", http.MethodGet, w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  description: "test desc"
  deprecated: true
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestMethod_WriteYaml_OptionalSecurity(t *testing.T) {
	opts := &DocOptions{}
	m := Method{
		Description:      "test desc",
		OptionalSecurity: true,
	}
	pathTemplate := urit.MustCreateTemplate("/root/foo")
	w := yaml.NewWriter(nil)
	m.writeYaml(opts, pathTemplate, nil, nil, "", http.MethodGet, w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `get:
  description: "test desc"
  security:
    - {}
  responses:
    200:
      description: "OK"
      content:
        application/json:
          schema:
            type: "object"
`
	assert.Equal(t, expect, string(data))
}
