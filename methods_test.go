package chioas

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethod_GetHandler_WithHandlerSet(t *testing.T) {
	m := Method{
		Handler: func(writer http.ResponseWriter, request *http.Request) {

		},
	}
	mh := m.getHandler(nil)
	assert.NotNil(t, mh)
}

func TestMethod_GetHandler_WithMethodNameSet(t *testing.T) {
	m := Method{
		MethodName: "Foo",
	}
	assert.Panics(t, func() {
		_ = m.getHandler(nil)
	})

	type dummyStruct struct{}
	assert.Panics(t, func() {
		_ = m.getHandler(&dummyStruct{})
	})

	dummy := &dummyWithMethod{}
	mh := m.getHandler(dummy)
	assert.NotNil(t, mh)
	assert.False(t, dummy.called)
	req, err := http.NewRequest(http.MethodGet, "/example", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	mh.ServeHTTP(res, req)
	assert.True(t, dummy.called)
}

type dummyWithMethod struct {
	called bool
}

func (d *dummyWithMethod) Foo(writer http.ResponseWriter, request *http.Request) {
	d.called = true
}

func TestMethod_GetHandler_NoHandlerOrMethodNameSet(t *testing.T) {
	m := Method{}
	assert.Panics(t, func() {
		_ = m.getHandler(nil)
	})
}

func TestMethods_WriteYaml(t *testing.T) {
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
	w := newYamlWriter(nil)
	ms.writeYaml("", w)
	data, err := w.bytes()
	require.NoError(t, err)
	const expect = `get:
head:
options:
post:
put:
patch:
delete:
trace:
`
	assert.Equal(t, expect, string(data))
}

func TestMethod_WriteYaml(t *testing.T) {
	m := Method{
		Description: "test desc",
		Summary:     "test summary",
		OperationId: "testOp",
		Tag:         "",
	}
	w := newYamlWriter(nil)
	m.writeYaml("foo", http.MethodGet, w)
	data, err := w.bytes()
	require.NoError(t, err)
	const expect = `get:
  summary: "test summary"
  description: "test desc"
  operationId: "testOp"
  tags:
    - "foo"
`
	assert.Equal(t, expect, string(data))
}
