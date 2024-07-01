package typed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/urit"
	"github.com/go-chi/chi/v5"
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
	raw, ok := mhb.(*builder)
	assert.True(t, ok)
	assert.Equal(t, 0, len(raw.argBuilders))
	assert.Nil(t, raw.errorHandler)
	assert.NotNil(t, raw.unmarshaler)

	mhb = NewTypedMethodsHandlerBuilder(defaultErrorHandler, nil, &testAdditional{}, nil)
	raw, ok = mhb.(*builder)
	assert.True(t, ok)
	assert.Equal(t, 1, len(raw.argBuilders))
	assert.NotNil(t, raw.errorHandler)
	assert.NotNil(t, raw.unmarshaler)

	um := &testUnmarshaler{}
	mhb = NewTypedMethodsHandlerBuilder(defaultErrorHandler, nil, &testAdditional{}, nil, um, um, nil)
	raw, ok = mhb.(*builder)
	assert.True(t, ok)
	assert.Equal(t, 1, len(raw.argBuilders))
	assert.NotNil(t, raw.errorHandler)
	assert.NotNil(t, raw.unmarshaler)
	assert.Equal(t, um, raw.unmarshaler)

	mhb = NewTypedMethodsHandlerBuilder(&testApiWithResponseHandler{})
	raw, ok = mhb.(*builder)
	assert.True(t, ok)
	assert.Nil(t, raw.errorHandler)
	assert.NotNil(t, raw.responseHandler)
}

func TestNewTypedMethodsHandlerBuilder_BadOption(t *testing.T) {
	type badOption struct{}
	mhb := NewTypedMethodsHandlerBuilder(nil, &badOption{}, &badOption{})
	assert.NotNil(t, mhb)
	raw, ok := mhb.(*builder)
	assert.True(t, ok)
	assert.Error(t, raw.initErr)

	_, err := mhb.BuildHandler("/", http.MethodGet, chioas.Method{}, nil)
	assert.Error(t, err)
	assert.Equal(t, "invalid option passed to NewTypedMethodsHandlerBuilder", err.Error())
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
	_, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)

	m = chioas.Method{
		Handler: true,
	}
	_, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
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
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)

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

func TestTypedMethodsHandlerBuilder_HandlerFor_OutArgsErrors(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder()
	_, err := tmhb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: func() (error, error) {
		return nil, nil
	}}, nil)
	assert.Error(t, err)
	assert.Equal(t, "error building return args (path: /, method: GET) - has multiple error return args", err.Error())

	_, err = tmhb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: func(req *http.Request) (error, error) {
		return nil, nil
	}}, nil)
	assert.Error(t, err)
	assert.Equal(t, "error building return args (path: /, method: GET) - has multiple error return args", err.Error())
}

type testErrorringArgBuilder struct {
}

func (t *testErrorringArgBuilder) IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool) {
	return argType.Kind() == reflect.Bool, false
}

func (t *testErrorringArgBuilder) BuildValue(argType reflect.Type, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
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
	assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)
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
				Handler: func() ([]uint8, int) {
					return []byte{'n', 'u', 'l', 'l'}, http.StatusAccepted
				},
			},
			expectStatus: http.StatusAccepted,
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
				Handler: func() (int, []byte) {
					return http.StatusAccepted, nil
				},
			},
			expectStatus: http.StatusAccepted,
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
		{
			mdef: chioas.Method{
				Handler: func() (int, error) {
					return http.StatusAccepted, nil
				},
			},
			expectStatus: http.StatusAccepted,
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
			assert.Equal(t, tc.expectStatus, res.Result().StatusCode)
			if tc.expectBody != "" {
				assert.Equal(t, tc.expectBody, res.Body.String())
			}
		})
	}
}

func TestTypedMethodsHandlerBuilder_WithMethodExpressions(t *testing.T) {
	tmhb := NewTypedMethodsHandlerBuilder()
	dummy := &dummyWithMethods{}
	m := chioas.Method{
		Handler: (*dummyWithMethods).Foo,
	}
	mh, err := tmhb.BuildHandler("/", http.MethodGet, m, dummy)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	assert.False(t, dummy.fooCalled)
	req, err := http.NewRequest(http.MethodGet, "/example", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	mh.ServeHTTP(res, req)
	assert.True(t, dummy.fooCalled)

	_, err = tmhb.BuildHandler("/", http.MethodGet, m, nil)
	assert.Error(t, err)
	assert.Equal(t, "cannot use method expressions when thisApi not supplied (path: /, method: GET)", err.Error())

	m = chioas.Method{
		Handler: (*dummyWithMethods).private,
	}
	_, err = tmhb.BuildHandler("/", http.MethodGet, m, dummy)
	assert.Error(t, err)
	assert.Equal(t, "supplied thisApi does not have public method 'private' (path: /, method: GET)", err.Error())

	m = chioas.Method{
		Handler: (*dummyWithMethods).Baz,
	}
	mh, err = tmhb.BuildHandler("/", http.MethodGet, m, dummy)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	assert.False(t, dummy.bazCalled)
	req, err = http.NewRequest(http.MethodGet, "/example", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/example", mh)
	r.ServeHTTP(res, req)
	assert.True(t, dummy.bazCalled)
}

func TestTypedMethodsHandlerBuilder_WithResponseHandler(t *testing.T) {
	dummy := &testApiWithResponseHandler{
		statusCode: http.StatusPaymentRequired,
	}
	tmhb := NewTypedMethodsHandlerBuilder(dummy)
	m := chioas.Method{
		Handler: (*testApiWithResponseHandler).Foo,
	}
	mh, err := tmhb.BuildHandler("/", http.MethodGet, m, dummy)
	assert.NoError(t, err)
	assert.NotNil(t, mh)

	req, _ := http.NewRequest(http.MethodGet, "/example", nil)
	res := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/example", mh)
	r.ServeHTTP(res, req)
	assert.True(t, dummy.called)
	assert.True(t, dummy.writeCalled)
	assert.False(t, dummy.writeErrCalled)
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)

	dummy.called = false
	dummy.writeCalled = false
	dummy.statusCode = http.StatusBadRequest
	dummy.err = errors.New("fooey")
	req, _ = http.NewRequest(http.MethodGet, "/example", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.True(t, dummy.called)
	assert.True(t, dummy.writeErrCalled)
	assert.False(t, dummy.writeCalled)
	assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode)
}

func TestNewTypedMethodsHandlerBuilder_GetErrorHandler(t *testing.T) {
	b := &builder{}
	eh := b.getErrorHandler(nil)
	assert.Equal(t, defaultErrorHandler, eh)

	b = &builder{
		errorHandler: &testErrorHandler{},
	}
	eh = b.getErrorHandler(nil)
	assert.Equal(t, b.errorHandler, eh)

	api := &testErrorHandler{}
	b = &builder{}
	eh = b.getErrorHandler(api)
	assert.Equal(t, api, eh)
}

func TestNewTypedMethodsHandlerBuilder_GetResponseHandler(t *testing.T) {
	b := &builder{}
	rh := b.getResponseHandler(nil)
	assert.Nil(t, rh)

	b = &builder{
		responseHandler: &testApiWithResponseHandler{},
	}
	rh = b.getResponseHandler(nil)
	assert.Equal(t, b.responseHandler, rh)

	b = &builder{}
	api := &testApiWithResponseHandler{}
	rh = b.getResponseHandler(api)
	assert.Equal(t, api, rh)
}

type testErrorHandler struct{}

var _ ErrorHandler = &testErrorHandler{}

func (t *testErrorHandler) HandleError(writer http.ResponseWriter, request *http.Request, err error) {
}

type testApiWithResponseHandler struct {
	called         bool
	writeCalled    bool
	writeErrCalled bool
	statusCode     int
	err            error
	response       any
}

func (t *testApiWithResponseHandler) WriteResponse(writer http.ResponseWriter, request *http.Request, value any, statusCode int, thisApi any) {
	t.writeCalled = true
	writer.WriteHeader(t.statusCode)
}

func (t *testApiWithResponseHandler) WriteErrorResponse(writer http.ResponseWriter, request *http.Request, err error, thisApi any) {
	t.writeErrCalled = true
	writer.WriteHeader(t.statusCode)
}

func (t *testApiWithResponseHandler) Foo() (any, int, error) {
	t.called = true
	return t.response, t.statusCode, t.err
}

var _ ResponseHandler = &testApiWithResponseHandler{}

type testUnmarshalble struct {
}

func (t *testUnmarshalble) MarshalJSON() ([]byte, error) {
	return nil, errors.New("foo")
}

type dummyWithMethods struct {
	fooCalled bool
	barCalled bool
	bazCalled bool
}

func (d *dummyWithMethods) private(writer http.ResponseWriter, request *http.Request) {
	d.fooCalled = true
}

func (d *dummyWithMethods) Foo(writer http.ResponseWriter, request *http.Request) {
	d.fooCalled = true
}

func (d *dummyWithMethods) Bar(pathParam ...string) json.RawMessage {
	d.barCalled = true
	return nil
}

func (d *dummyWithMethods) Baz(ctx context.Context, writer http.ResponseWriter) error {
	d.bazCalled = true
	return nil
}

func (d *dummyWithMethods) Errs() (json.RawMessage, error) {
	return nil, errors.New("foo")
}

func (d *dummyWithMethods) HandleError(writer http.ResponseWriter, request *http.Request, err error) {
	writer.WriteHeader(http.StatusBadGateway)
}
