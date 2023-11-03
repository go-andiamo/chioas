package chioas

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMethodHandlerBuilder(t *testing.T) {
	mhb := getMethodHandlerBuilder(nil)
	assert.Equal(t, defaultMethodHandlerBuilder, mhb)

	dummy := &dummyMethodHandlerBuilder{}
	mhb = getMethodHandlerBuilder(dummy)
	assert.Equal(t, dummy, mhb)
}

type dummyMethodHandlerBuilder struct {
}

func (d *dummyMethodHandlerBuilder) BuildHandler(path string, method string, mdef Method, thisApi any) (http.HandlerFunc, error) {
	return nil, nil
}

func TestDefaultMethodHandlerBuilder_Build_NoHandlerOrMethodNameSet(t *testing.T) {
	m := Method{}
	_, err := defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, nil)
	assert.Error(t, err)
}

func TestDefaultMethodHandlerBuilder_Build_WithGetHandler(t *testing.T) {
	m := Method{
		Handler: func(path string, method string, thisApi any) (http.HandlerFunc, error) {
			return nil, errors.New("foo")
		},
	}
	mh, err := defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, nil)
	assert.Error(t, err)
	assert.Nil(t, mh)
}

func TestDefaultMethodHandlerBuilder_Build_WithHandlerSet(t *testing.T) {
	m := Method{
		Handler: func(writer http.ResponseWriter, request *http.Request) {},
	}
	mh, err := defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	var hf http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {}
	m = Method{
		Handler: hf,
	}
	mh, err = defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)
}

func TestDefaultMethodHandlerBuilder_Build_WithMethodNameSet(t *testing.T) {
	m := Method{
		Handler: "Foo",
	}
	_, err := defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, nil)
	assert.Error(t, err)

	type dummyStruct struct{}
	_, err = defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, &dummyStruct{})
	assert.Error(t, err)

	dummy := &dummyWithMethod{}
	mh, err := defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, dummy)
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
	_, err = defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, dummy)
	assert.Error(t, err)

	m = Method{
		Handler: false,
	}
	_, err = defaultMethodHandlerBuilder.BuildHandler(root, http.MethodGet, m, dummy)
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
