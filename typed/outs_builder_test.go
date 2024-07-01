package typed

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewOutsBuilder(t *testing.T) {
	type someStruct struct{}
	testCases := []struct {
		fn                     any
		expectErr              string
		expectErrArg           int
		expectStatusCodeArg    int
		expectMarshableArg     int
		expectMarshableHandler any
	}{
		{
			fn:                  func() {},
			expectErrArg:        -1,
			expectStatusCodeArg: -1,
			expectMarshableArg:  -1,
		},
		{
			fn:                  func() error { return nil },
			expectErrArg:        0,
			expectStatusCodeArg: -1,
			expectMarshableArg:  -1,
		},
		{
			fn:                  func() ApiError { return nil },
			expectErrArg:        0,
			expectStatusCodeArg: -1,
			expectMarshableArg:  -1,
		},
		{
			fn:                  func() int { return 0 },
			expectErrArg:        -1,
			expectStatusCodeArg: 0,
			expectMarshableArg:  -1,
		},
		{
			fn:                     func() JsonResponse { return JsonResponse{} },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: jsonResponseHandler,
		},
		{
			fn:                     func() *JsonResponse { return nil },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: jsonResponsePtrHandler,
		},
		{
			fn:                     func() []byte { return nil },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: bytesResponseHandler,
		},
		{
			fn:                     func() []uint8 { return nil },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: bytesResponseHandler,
		},
		{
			fn:                     func() someStruct { return someStruct{} },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: marshalerHandler,
		},
		{
			fn:                     func() *someStruct { return nil },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: marshalerPtrHandler,
		},
		{
			fn:                     func() ResponseMarshaler { return nil },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: responseMarshalerHandler,
		},
		{
			fn:                     func() any { return nil },
			expectErrArg:           -1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: anyOrErrorHandler,
		},
		{
			fn:                     func() (any, error) { return nil, nil },
			expectErrArg:           1,
			expectStatusCodeArg:    -1,
			expectMarshableArg:     0,
			expectMarshableHandler: anyHandler,
		},
		{
			fn:        func() (any, any) { return nil, nil },
			expectErr: errMultiMarshable,
		},
		{
			fn:        func() (any, any, any, any) { return nil, nil, nil, nil },
			expectErr: errTooMany,
		},
		{
			fn:        func() (error, error) { return nil, nil },
			expectErr: errMultiErrs,
		},
		{
			fn:        func() (error, ApiError) { return nil, nil },
			expectErr: errMultiErrs,
		},
		{
			fn:        func() (int, int) { return 0, 0 },
			expectErr: errMultiStatus,
		},
		{
			fn:        func() (*someStruct, *someStruct) { return nil, nil },
			expectErr: errMultiMarshable,
		},
		{
			fn:        func() (JsonResponse, JsonResponse) { return JsonResponse{}, JsonResponse{} },
			expectErr: errMultiMarshable,
		},
		{
			fn:        func() (*JsonResponse, *JsonResponse) { return nil, nil },
			expectErr: errMultiMarshable,
		},
		{
			fn:        func() ([]byte, []uint8) { return nil, nil },
			expectErr: errMultiMarshable,
		},
		{
			fn:        func() (ResponseMarshaler, ResponseMarshaler) { return nil, nil },
			expectErr: errMultiMarshable,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			mf := reflect.ValueOf(tc.fn)
			ob, err := newOutsBuilder(mf)
			if tc.expectErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, ob)
				assert.Equal(t, tc.expectErrArg, ob.errArg)
				assert.Equal(t, tc.expectStatusCodeArg, ob.statusCodeArg)
				assert.Equal(t, tc.expectMarshableArg, ob.marshableArg)
				if tc.expectMarshableArg != -1 {
					chkpt := reflect.ValueOf(tc.expectMarshableHandler).Pointer()
					hpt := reflect.ValueOf(ob.marshableHandler).Pointer()
					assert.Equal(t, chkpt, hpt)
				}
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
				assert.Nil(t, ob)
			}
		})
	}
}

func TestJsonResponseHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	jr := JsonResponse{}
	assert.True(t, jsonResponseHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	res = httptest.NewRecorder()
	jr = JsonResponse{Error: NewApiError(http.StatusNotImplemented, "")}
	assert.True(t, jsonResponseHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNotImplemented, res.Result().StatusCode)

	res = httptest.NewRecorder()
	jr = JsonResponse{Error: errors.New("")}
	assert.True(t, jsonResponseHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)
}

func TestJsonResponsePtrHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	jr := &JsonResponse{}
	assert.True(t, jsonResponsePtrHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	res = httptest.NewRecorder()
	jr = nil
	assert.True(t, jsonResponsePtrHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	res = httptest.NewRecorder()
	jr = nil
	assert.True(t, jsonResponsePtrHandler(reflect.ValueOf(jr), b, nil, http.StatusAccepted, res, req))
	assert.Equal(t, http.StatusAccepted, res.Result().StatusCode)

	res = httptest.NewRecorder()
	jr = &JsonResponse{Error: NewApiError(http.StatusNotImplemented, "")}
	assert.True(t, jsonResponsePtrHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNotImplemented, res.Result().StatusCode)

	res = httptest.NewRecorder()
	jr = &JsonResponse{Error: errors.New("")}
	assert.True(t, jsonResponsePtrHandler(reflect.ValueOf(jr), b, nil, 0, res, req))
	assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)
}

func TestBytesResponseHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	data := make([]byte, 0)
	assert.True(t, bytesResponseHandler(reflect.ValueOf(data), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	res = httptest.NewRecorder()
	data = nil
	assert.True(t, bytesResponseHandler(reflect.ValueOf(data), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	res = httptest.NewRecorder()
	data = nil
	assert.True(t, bytesResponseHandler(reflect.ValueOf(data), b, nil, http.StatusAccepted, res, req))
	assert.Equal(t, http.StatusAccepted, res.Result().StatusCode)

	res = httptest.NewRecorder()
	data = []byte("null")
	assert.True(t, bytesResponseHandler(reflect.ValueOf(data), b, nil, 0, res, req))
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)

	res = httptest.NewRecorder()
	data = []byte("null")
	assert.True(t, bytesResponseHandler(reflect.ValueOf(data), b, nil, http.StatusCreated, res, req))
	assert.Equal(t, http.StatusCreated, res.Result().StatusCode)
}

func TestResponseMarshalerHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	rm := &testResponseMarshaler{}
	assert.True(t, responseMarshalerHandler(reflect.ValueOf(rm), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	res = httptest.NewRecorder()
	rm = &testResponseMarshaler{err: NewApiError(http.StatusNotFound, "")}
	assert.True(t, responseMarshalerHandler(reflect.ValueOf(rm), b, nil, 0, res, req))
	assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)

	res = httptest.NewRecorder()
	rm = &testResponseMarshaler{data: []byte("null"), hdrs: [][2]string{{hdrContentType, contentTypeXml}}}
	assert.True(t, responseMarshalerHandler(reflect.ValueOf(rm), b, nil, 0, res, req))
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeXml, res.Result().Header.Get(hdrContentType))

	res = httptest.NewRecorder()
	rm = &testResponseMarshaler{data: []byte("null"), hdrs: [][2]string{{hdrContentType, contentTypeXml}}}
	assert.True(t, responseMarshalerHandler(reflect.ValueOf(rm), b, nil, http.StatusCreated, res, req))
	assert.Equal(t, http.StatusCreated, res.Result().StatusCode)
	assert.Equal(t, contentTypeXml, res.Result().Header.Get(hdrContentType))

	var rmv ResponseMarshaler = nil
	res = httptest.NewRecorder()
	assert.False(t, responseMarshalerHandler(reflect.ValueOf(rmv), b, nil, 0, res, req))
}

func TestMarshalerHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	s := struct{}{}
	assert.True(t, marshalerHandler(reflect.ValueOf(s), b, nil, 0, res, req))
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeJson, res.Result().Header.Get(hdrContentType))

	res = httptest.NewRecorder()
	s2 := &testUnmarshalble{}
	assert.True(t, marshalerHandler(reflect.ValueOf(s2), b, nil, 0, res, req))
	assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)

	res = httptest.NewRecorder()
	s2 = nil
	assert.True(t, marshalerHandler(reflect.ValueOf(s2), b, nil, 0, res, req))
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)

	res = httptest.NewRecorder()
	s2 = nil
	assert.True(t, marshalerHandler(reflect.ValueOf(s2), b, nil, http.StatusNoContent, res, req))
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)
}

func TestMarshalerPtrHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	type someStruct struct{}
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	var returnVal *someStruct

	res := httptest.NewRecorder()
	assert.False(t, marshalerPtrHandler(reflect.ValueOf(returnVal), b, nil, 0, res, req))
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)

	res = httptest.NewRecorder()
	returnVal = &someStruct{}
	assert.True(t, marshalerPtrHandler(reflect.ValueOf(returnVal), b, nil, 0, res, req))
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)

	res = httptest.NewRecorder()
	returnVal = &someStruct{}
	assert.True(t, marshalerPtrHandler(reflect.ValueOf(returnVal), b, nil, http.StatusCreated, res, req))
	assert.Equal(t, http.StatusCreated, res.Result().StatusCode)
}

func TestAnyHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	type myStruct struct{}
	testCases := []struct {
		fn           any
		expectStatus int
	}{
		{
			fn: func() (any, error) {
				return []byte{}, nil
			},
			expectStatus: http.StatusNoContent,
		},
		{
			fn: func() (any, error) {
				return nil, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				return JsonResponse{StatusCode: http.StatusCreated}, nil
			},
			expectStatus: http.StatusCreated,
		},
		{
			fn: func() (any, error) {
				return &JsonResponse{StatusCode: http.StatusCreated}, nil
			},
			expectStatus: http.StatusCreated,
		},
		{
			fn: func() (any, error) {
				return &testResponseMarshaler{statusCode: http.StatusCreated}, nil
			},
			expectStatus: http.StatusCreated,
		},
		{
			fn: func() (any, error) {
				return struct{}{}, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				var anyErr error
				return anyErr, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				return errors.New("fooey"), nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				var emptyMy *myStruct = nil
				return emptyMy, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				return myStruct{}, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				return "", nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				str := ""
				return &str, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() (any, error) {
				return false, nil
			},
			expectStatus: http.StatusOK,
		},
	}
	chkpt := reflect.ValueOf(anyHandler).Pointer()
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			mf := reflect.ValueOf(tc.fn)
			ob, err := newOutsBuilder(mf)
			assert.NoError(t, err)
			hpt := reflect.ValueOf(ob.marshableHandler).Pointer()
			assert.Equal(t, chkpt, hpt)
			out := mf.Call([]reflect.Value{})
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()
			ob.handleReturnArgs(out, b, nil, res, req)
			assert.Equal(t, tc.expectStatus, res.Result().StatusCode)
		})
	}
}

func TestAnyOrErrorHandler(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	type myStruct struct{}
	testCases := []struct {
		fn           any
		expectStatus int
	}{
		{
			fn: func() any {
				return []byte{}
			},
			expectStatus: http.StatusNoContent,
		},
		{
			fn: func() any {
				return nil
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() any {
				return JsonResponse{StatusCode: http.StatusCreated}
			},
			expectStatus: http.StatusCreated,
		},
		{
			fn: func() any {
				return &JsonResponse{StatusCode: http.StatusCreated}
			},
			expectStatus: http.StatusCreated,
		},
		{
			fn: func() any {
				return &testResponseMarshaler{statusCode: http.StatusCreated}
			},
			expectStatus: http.StatusCreated,
		},
		{
			fn: func() any {
				return struct{}{}
			},
			expectStatus: http.StatusOK,
		},
		{
			fn: func() any {
				return errors.New("fooey")
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			fn: func() any {
				return NewApiError(http.StatusBadGateway, "")
			},
			expectStatus: http.StatusBadGateway,
		},
	}
	chkpt := reflect.ValueOf(anyOrErrorHandler).Pointer()
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			mf := reflect.ValueOf(tc.fn)
			ob, err := newOutsBuilder(mf)
			assert.NoError(t, err)
			hpt := reflect.ValueOf(ob.marshableHandler).Pointer()
			assert.Equal(t, chkpt, hpt)
			out := mf.Call([]reflect.Value{})
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()
			ob.handleReturnArgs(out, b, nil, res, req)
			assert.Equal(t, tc.expectStatus, res.Result().StatusCode)
		})
	}
}

func TestHandleReturnArgs_NotHandled(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	returnSc := 0
	fn := func() int {
		return returnSc
	}
	mf := reflect.ValueOf(fn)
	ob, err := newOutsBuilder(mf)
	assert.NoError(t, err)
	assert.NotNil(t, ob)

	out := mf.Call([]reflect.Value{})
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	ob.handleReturnArgs(out, b, nil, res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)

	returnSc = http.StatusAccepted
	out = mf.Call([]reflect.Value{})
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	ob.handleReturnArgs(out, b, nil, res, req)
	assert.Equal(t, http.StatusAccepted, res.Result().StatusCode)
}

func TestHandleReturnArgs(t *testing.T) {
	b := NewTypedMethodsHandlerBuilder().(*builder)
	var returnErr error
	returnSc := 0
	returnData := make([]byte, 0)
	fn := func() (int, []byte, error) {
		return returnSc, returnData, returnErr
	}
	mf := reflect.ValueOf(fn)
	ob, err := newOutsBuilder(mf)
	assert.NoError(t, err)
	assert.NotNil(t, ob)

	out := mf.Call([]reflect.Value{})
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	ob.handleReturnArgs(out, b, nil, res, req)
	assert.Equal(t, http.StatusNoContent, res.Result().StatusCode)

	returnSc = http.StatusPaymentRequired
	out = mf.Call([]reflect.Value{})
	res = httptest.NewRecorder()
	ob.handleReturnArgs(out, b, nil, res, req)
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)

	returnErr = NewApiError(http.StatusNotImplemented, "")
	out = mf.Call([]reflect.Value{})
	res = httptest.NewRecorder()
	ob.handleReturnArgs(out, b, nil, res, req)
	assert.Equal(t, http.StatusNotImplemented, res.Result().StatusCode)
}
