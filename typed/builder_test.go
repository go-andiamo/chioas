package typed

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/urit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewTypedMethodsHandlerBuilder(t *testing.T) {
	mhb := NewTypedMethodsHandlerBuilder()
	assert.NotNil(t, mhb)
	raw, ok := mhb.(*typedMethodsHandlerBuilder)
	assert.True(t, ok)
	assert.Equal(t, 0, len(raw.argBuilders))
	assert.Nil(t, raw.errorHandler)
	assert.NotNil(t, raw.unmarshaler)

	mhb = NewTypedMethodsHandlerBuilder(defaultErrorHandler, nil, &testAdditional{}, nil)
	raw, ok = mhb.(*typedMethodsHandlerBuilder)
	assert.True(t, ok)
	assert.Equal(t, 1, len(raw.argBuilders))
	assert.NotNil(t, raw.errorHandler)
	assert.NotNil(t, raw.unmarshaler)

	um := &testUnmarshaler{}
	mhb = NewTypedMethodsHandlerBuilder(defaultErrorHandler, nil, &testAdditional{}, nil, um, um, nil)
	raw, ok = mhb.(*typedMethodsHandlerBuilder)
	assert.True(t, ok)
	assert.Equal(t, 1, len(raw.argBuilders))
	assert.NotNil(t, raw.errorHandler)
	assert.NotNil(t, raw.unmarshaler)
	assert.Equal(t, um, raw.unmarshaler)
}

type testUnmarshaler struct {
}

func (t *testUnmarshaler) Unmarshal(request *http.Request, v any) error {
	return nil
}

func TestTypedMethodsHandlerBuilder_Build_NoHandlerOrMethodNameSet(t *testing.T) {
	m := chioas.Method{}
	tmhb := NewTypedMethodsHandlerBuilder()
	_, err := tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)
}

func TestTypedMethodsHandlerBuilder_Build_WithGetHandler(t *testing.T) {
	m := chioas.Method{
		Handler: func(path string, method string, thisApi any) (http.HandlerFunc, error) {
			return nil, errors.New("foo")
		},
	}
	tmhb := NewTypedMethodsHandlerBuilder()
	mh, err := tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)
	assert.Nil(t, mh)
}

func TestTypedMethodsHandlerBuilder_Build_WithHandlerSet(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder()

	m := chioas.Method{
		Handler: func(request *http.Request) json.RawMessage {
			return nil
		},
	}
	mh, err := tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	var hf = func(request *http.Request) json.RawMessage {
		return nil
	}
	m = chioas.Method{
		Handler: hf,
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	m = chioas.Method{
		Handler: func(writer http.ResponseWriter, request *http.Request) {},
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	var hf2 http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {}
	m = chioas.Method{
		Handler: hf2,
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	m = chioas.Method{
		Handler: func(unknown bool) json.RawMessage {
			return nil
		},
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)

	m = chioas.Method{
		Handler: true,
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)
}

func TestTypedMethodsHandlerBuilder_Build_WithMethodNameSet(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder()

	m := chioas.Method{
		Handler: "Foo",
	}
	_, err := tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)

	type dummyStruct struct{}
	_, err = tmhb.BuildHandler("/", http.MethodGet, m, &dummyStruct{})
	assert.Error(t, err)

	dummy := &dummyWithMethods{}
	mh, err := tmhb.BuildHandler("/", http.MethodGet, m, dummy)
	assert.NoError(t, err)
	assert.NotNil(t, mh)
	assert.False(t, dummy.fooCalled)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	mh.ServeHTTP(res, req)
	assert.True(t, dummy.fooCalled)

	m = chioas.Method{
		Handler: "Bar",
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, dummy)
	assert.NoError(t, err)
	assert.NotNil(t, mh)
	assert.False(t, dummy.barCalled)
	req, err = http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	mh.ServeHTTP(res, req)
	assert.True(t, dummy.barCalled)
}

func TestTypedMethodsHandlerBuilder_HandlerFor_ZeroInOut(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder()
	called := false

	hf, err := tmhb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: func() {
		called = true
	}}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.True(t, called)
}

func TestTypedMethodsHandlerBuilder_HandlerFor_ZeroInSomeOut(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder()
	called := false

	hf, err := tmhb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: func() (map[string]any, error) {
		called = true
		return nil, NewApiError(http.StatusPaymentRequired, "")
	}}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)
	assert.False(t, called)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusPaymentRequired, res.Code)

	called = false
	hf, err = tmhb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: func() json.RawMessage {
		called = true
		return nil
	}}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)
	assert.False(t, called)
	req, err = http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.True(t, called)
}

type testErrorringArgBuilder struct {
}

func (t *testErrorringArgBuilder) IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool) {
	return argType.Kind() == reflect.Bool, false
}

func (t *testErrorringArgBuilder) BuildValue(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.Value{}, errors.New("fooey")
}

func TestTypedMethodsHandlerBuilder_HandlerFor_SomeIn_BuilderErrors(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder(&testErrorringArgBuilder{})

	hf, err := tmhb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: func(test bool) (map[string]any, error) {
		return nil, NewApiError(http.StatusPaymentRequired, "")
	}}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

type testResponseMarshaler struct {
	data       []byte
	statusCode int
	hdrs       [][2]string
	err        error
}

func (t *testResponseMarshaler) Marshal(request *http.Request) (data []byte, statusCode int, hdrs [][2]string, err error) {
	return t.data, t.statusCode, t.hdrs, t.err
}

func TestTypedMethodsHandlerBuilder_ResponseTypes(t *testing.T) {
	testCases := []struct {
		mdef         chioas.Method
		thisApi      any
		expectStatus int
		expectBody   string
		errorHandler ErrorHandler
	}{
		{
			mdef: chioas.Method{
				Handler: func() ResponseMarshaler {
					return &testResponseMarshaler{}
				},
			},
			expectStatus: http.StatusNoContent,
		},
		{
			mdef: chioas.Method{
				Handler: func() ResponseMarshaler {
					return &testResponseMarshaler{
						hdrs: [][2]string{{hdrContentType, contentTypeJson}},
					}
				},
			},
			expectStatus: http.StatusNoContent,
		},
		{
			mdef: chioas.Method{
				Handler: func() ResponseMarshaler {
					return &testResponseMarshaler{
						data: []byte{'n', 'u', 'l', 'l'},
					}
				},
			},
			expectStatus: http.StatusOK,
			expectBody:   "null",
		},
		{
			mdef: chioas.Method{
				Handler: func() ResponseMarshaler {
					return &testResponseMarshaler{
						statusCode: http.StatusPaymentRequired,
					}
				},
			},
			expectStatus: http.StatusPaymentRequired,
		},
		{
			mdef: chioas.Method{
				Handler: func() ResponseMarshaler {
					return &testResponseMarshaler{
						err: NewApiError(http.StatusNotImplemented, ""),
					}
				},
			},
			expectStatus: http.StatusNotImplemented,
		},
		{
			mdef: chioas.Method{
				Handler: func() JsonResponse {
					return JsonResponse{
						Body: map[string]any{"foo": "bar"},
					}
				},
			},
			expectStatus: http.StatusOK,
			expectBody:   `{"foo":"bar"}`,
		},
		{
			mdef: chioas.Method{
				Handler: func() JsonResponse {
					return JsonResponse{
						Error: NewApiError(http.StatusNotImplemented, ""),
					}
				},
			},
			expectStatus: http.StatusNotImplemented,
		},
		{
			mdef: chioas.Method{
				Handler: func() *JsonResponse {
					return nil
				},
			},
			expectStatus: http.StatusNoContent,
		},
		{
			mdef: chioas.Method{
				Handler: func() *JsonResponse {
					return &JsonResponse{
						StatusCode: http.StatusPaymentRequired,
						Body:       map[string]any{"foo": "bar"},
					}
				},
			},
			expectStatus: http.StatusPaymentRequired,
			expectBody:   `{"foo":"bar"}`,
		},
		{
			mdef: chioas.Method{
				Handler: func() *JsonResponse {
					return &JsonResponse{
						Error: NewApiError(http.StatusNotImplemented, ""),
					}
				},
			},
			expectStatus: http.StatusNotImplemented,
		},
		{
			mdef: chioas.Method{
				Handler: func() []byte {
					return []byte{'n', 'u', 'l', 'l'}
				},
			},
			expectStatus: http.StatusOK,
			expectBody:   `null`,
		},
		{
			mdef: chioas.Method{
				Handler: func() []uint8 {
					return []byte{'n', 'u', 'l', 'l'}
				},
			},
			expectStatus: http.StatusOK,
			expectBody:   `null`,
		},
		{
			mdef: chioas.Method{
				Handler: func() []byte {
					return nil
				},
			},
			expectStatus: http.StatusNoContent,
		},
		{
			mdef: chioas.Method{
				Handler: func() any {
					return &testUnmarshalble{}
				},
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			mdef: chioas.Method{
				Handler: func() any {
					return &testUnmarshalble{}
				},
			},
			errorHandler: defaultErrorHandler,
			expectStatus: http.StatusInternalServerError,
		},
		{
			mdef: chioas.Method{
				Handler: "Foo",
			},
			thisApi:      &dummyWithMethods{},
			expectStatus: http.StatusOK,
		},
		{
			mdef: chioas.Method{
				Handler: "Errs",
			},
			thisApi:      &dummyWithMethods{},
			expectStatus: http.StatusBadGateway,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			tmhb := NewTypedMethodsHandlerBuilder(tc.errorHandler)
			hf, err := tmhb.BuildHandler("/", http.MethodGet, tc.mdef, tc.thisApi)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)
			res := httptest.NewRecorder()
			hf.ServeHTTP(res, req)
			assert.Equal(t, tc.expectStatus, res.Code)
			if tc.expectBody != "" {
				assert.Equal(t, tc.expectBody, res.Body.String())
			}
		})
	}
}

type testUnmarshalble struct {
}

func (t *testUnmarshalble) MarshalJSON() ([]byte, error) {
	return nil, errors.New("foo")
}

type dummyWithMethods struct {
	fooCalled bool
	barCalled bool
}

func (d *dummyWithMethods) Foo(writer http.ResponseWriter, request *http.Request) {
	d.fooCalled = true
}

func (d *dummyWithMethods) Bar(pathParam ...string) json.RawMessage {
	d.barCalled = true
	return nil
}

func (d *dummyWithMethods) Errs() (json.RawMessage, error) {
	return nil, errors.New("foo")
}

func (d *dummyWithMethods) HandleError(writer http.ResponseWriter, request *http.Request, err error) {
	writer.WriteHeader(http.StatusBadGateway)
}
