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
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestNewInsBuilder_ErrorsWithBadPath(t *testing.T) {
	_, err := newInsBuilder(reflect.ValueOf(func() {}), "", "", nil)
	assert.Error(t, err)
}

func TestInsBuilder_Build(t *testing.T) {
	fn := func(req *http.Request, w http.ResponseWriter, pathParams ...string) {}
	path := "/foo/{fooid}/bar/{barid}"
	inb, err := newInsBuilder(reflect.ValueOf(fn), path, "", &builder{unmarshaler: defaultUnmarshaler})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/foo/1/bar/2", nil)
	args, err := inb.build(w, req)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(args))
	assert.Equal(t, req, args[0].Interface())
	assert.Equal(t, w, args[1].Interface())
	assert.Equal(t, "1", args[2].Interface())
	assert.Equal(t, "2", args[3].Interface())
}

func TestInsBuilder_Build_ErrorsWithVaradicBuilderError(t *testing.T) {
	inb := &insBuilder{
		len:          1,
		isVaradic:    true,
		pathTemplate: urit.MustCreateTemplate("/"),
		argTypes:     []reflect.Type{dummyType},
		valueBuilders: []inValueBuilder{
			func(arg reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
				return reflect.Value{}, errors.New("foo")
			},
		},
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name":1}`))
	_, err := inb.build(w, req)
	assert.Error(t, err)
}

func TestInsBuilder_Build_ErrorsWithBadRequest(t *testing.T) {
	fn := func(req testRequest) {}
	path := "/"
	inb, err := newInsBuilder(reflect.ValueOf(fn), path, "", &builder{unmarshaler: defaultUnmarshaler})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name":1}`))
	_, err = inb.build(w, req)
	assert.Error(t, err)
	assert.Equal(t, "json: cannot unmarshal number into Go struct field testRequest.name of type string", err.Error())
}

func TestInsBuilder_Build_ErrorsWithBadPath(t *testing.T) {
	fn := func(req *http.Request, w http.ResponseWriter, pathParams ...string) {}
	path := "/foo/{fooid}/bar/{barid}"
	inb, err := newInsBuilder(reflect.ValueOf(fn), path, "", &builder{unmarshaler: defaultUnmarshaler})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	_, err = inb.build(w, req)
	assert.Error(t, err)
}

func TestInsBuilder_MakeBuilders(t *testing.T) {
	testCases := []struct {
		fn             any
		expectErr      string
		expectBuilders int
		expectVaradic  bool
		checkTypes     []any
		method         string
	}{
		{
			fn: func() {},
		},
		{
			fn:             func(req testRequest) {},
			expectBuilders: 1,
		},
		{
			fn:             func(req *testRequest) {},
			expectBuilders: 1,
		},
		{
			fn:             func(req []testRequest) {},
			expectBuilders: 1,
		},
		{
			fn:        func(req testRequest, req2 testRequest) {},
			expectErr: "multiple args could be from request.Body",
		},
		{
			fn:             func(pathParam string) {},
			expectBuilders: 1,
		},
		{
			fn:             func(pathParams ...string) {},
			expectBuilders: 1,
			expectVaradic:  true,
		},
		{
			fn:             func(pathParams []string) {},
			expectBuilders: 1,
		},
		{
			fn:        func(unknown int) {},
			expectErr: "cannot determine arg 0",
		},
		{
			fn:        func(unknown []int) {},
			expectErr: "cannot determine arg 0",
		},
		{
			fn:             func(body json.RawMessage) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderJsonRawMessage},
		},
		{
			fn:             func(body []byte) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderByteBody},
		},
		{
			fn:             func(body []uint8) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderByteBody},
		},
		{
			fn:             func(w http.ResponseWriter) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderHttpResponseWriter},
		},
		{
			fn:             func(r *http.Request) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderHttpRequest},
		},
		{
			fn:             func(ctx context.Context) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderContext},
		},
		{
			fn:             func(ctx chi.Context) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderChiContext},
		},
		{
			fn:             func(ctx *chi.Context) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderChiContextPtr},
		},
		{
			fn:             func(hdrs http.Header) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderHttpHeader},
		},
		{
			fn:             func(cookies []*http.Cookie) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderCookies},
		},
		{
			fn:             func(url *url.URL) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderUrl},
		},
		{
			fn:             func(hdrs Headers) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderTypedHeaders},
		},
		{
			fn:             func(hdrs map[string][]string) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderMapHeaders},
		},
		{
			fn:             func(pathParam PathParams) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderTypedPathParams},
		},
		{
			fn:             func(queryParams QueryParams) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderQueryParams},
		},
		{
			fn:             func(rawQuery RawQuery) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderRawQuery},
		},
		{
			fn:        func(req *http.Request, unknown int) {},
			expectErr: "cannot determine arg 1",
		},
		{
			fn:        func(res http.Response) {},
			expectErr: "cannot determine arg 0", // because package http is excluded!
		},
		{
			fn:        func(res *http.Response) {},
			expectErr: "cannot determine arg 0", // because package http is excluded!
		},
		{
			fn:        func(res []*http.Response) {},
			expectErr: "cannot determine arg 0", // because package http is excluded!
		},
		{
			fn:        func(mf multipart.Form) {},
			expectErr: "cannot determine arg 0", // because package multipart is excluded!
		},
		{
			fn:        func(mf *multipart.Form) {},
			expectErr: "cannot determine arg 0", // because package multipart is excluded!
		},
		{
			fn:             func(f PostForm) {},
			expectBuilders: 1,
			method:         http.MethodPost,
			checkTypes:     []any{commonBuilderPostForm},
		},
		{
			fn:        func(f PostForm, f2 PostForm) {},
			method:    http.MethodPost,
			expectErr: "multiple args could be from request.Body",
		},
		{
			fn:             func(f PostForm) {},
			expectBuilders: 1,
			method:         http.MethodPut,
			checkTypes:     []any{commonBuilderPostForm},
		},
		{
			fn:             func(f PostForm) {},
			expectBuilders: 1,
			method:         http.MethodPatch,
			checkTypes:     []any{commonBuilderPostForm},
		},
		{
			fn:             func(f PostForm) {},
			expectBuilders: 1,
			method:         http.MethodGet,
			checkTypes:     []any{commonBuilderPostFormEmpty},
		},
		{
			fn:             func(f PostForm, f2 PostForm) {},
			expectBuilders: 2,
			method:         http.MethodGet,
			checkTypes:     []any{commonBuilderPostFormEmpty, commonBuilderPostFormEmpty},
		},
		{
			fn:             func(ba BasicAuth) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderBasicAuth},
		},
		{
			fn:             func(ba *BasicAuth) {},
			expectBuilders: 1,
			checkTypes:     []any{commonBuilderBasicAuthPtr},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			fv := reflect.ValueOf(tc.fn)
			if !fv.IsValid() || fv.Type().Kind() != reflect.Func {
				t.Fatalf("test must be a func")
			}
			inb, err := newInsBuilder(fv, "/", tc.method, &builder{unmarshaler: defaultUnmarshaler})
			if tc.expectErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectBuilders, inb.len)
				assert.Equal(t, tc.expectBuilders, len(inb.valueBuilders))
				assert.Equal(t, tc.expectVaradic, inb.isVaradic)
				for j, chkT := range tc.checkTypes {
					if chkT != nil {
						vbpt := reflect.ValueOf(inb.valueBuilders[j]).Pointer()
						chkpt := reflect.ValueOf(chkT).Pointer()
						assert.Equal(t, chkpt, vbpt)
					}
				}
			}
		})
	}
}

func TestInsBuilder_MakeBuilders_WithAdditionals(t *testing.T) {
	wasGet := false
	fn := func(isGet bool, pathParams ...string) {
		wasGet = isGet
	}
	mf := reflect.ValueOf(fn)
	_, err := newInsBuilder(mf, "/", "", &builder{unmarshaler: defaultUnmarshaler})
	assert.Error(t, err)

	additional := &testAdditional{}
	inb, err := newInsBuilder(mf, "/", "", &builder{unmarshaler: defaultUnmarshaler, argBuilders: []ArgBuilder{additional}})
	assert.NoError(t, err)
	assert.Equal(t, 2, inb.len)
	assert.True(t, additional.applicableCalled)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	args, err := inb.build(w, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(args))
	_ = mf.Call(args)
	assert.True(t, additional.buildValueCalled)
	assert.False(t, wasGet)

	additional.buildValueCalled = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	args, err = inb.build(w, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(args))
	_ = mf.Call(args)
	assert.True(t, additional.buildValueCalled)
	assert.True(t, wasGet)
}

type testAdditional struct {
	applicableCalled bool
	buildValueCalled bool
}

func (t *testAdditional) IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool) {
	t.applicableCalled = true
	return argType.Kind() == reflect.Bool, true
}
func (t *testAdditional) BuildValue(argType reflect.Type, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	t.buildValueCalled = true
	result := request.Method == http.MethodGet
	return reflect.ValueOf(result), nil
}

var dummyType = reflect.TypeOf("")

func TestCommonBuilderHttpResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderHttpResponseWriter(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, w, v.Interface())
}

func TestCommonBuilderHttpRequest(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderHttpRequest(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, req, v.Interface())
}

func TestCommonBuilderContext(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderContext(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, req.Context(), v.Interface())
}

func TestCommonBuilderChiContextPtr(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderChiContextPtr(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Nil(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/foo", nil)
	router := chi.NewRouter()
	router.Get("/foo", func(writer http.ResponseWriter, request *http.Request) {
		v, err = commonBuilderChiContextPtr(dummyType, writer, request, nil)
	})
	router.ServeHTTP(w, req)
	assert.NoError(t, err)
	assert.NotNil(t, v.Interface())
	ctx, ok := v.Interface().(*chi.Context)
	assert.True(t, ok)
	assert.Equal(t, http.MethodGet, ctx.RouteMethod)
	assert.Equal(t, []string{"/foo"}, ctx.RoutePatterns)
}

func TestCommonBuilderChiContext(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderChiContext(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, chi.Context{}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/foo", nil)
	router := chi.NewRouter()
	router.Get("/foo", func(writer http.ResponseWriter, request *http.Request) {
		v, err = commonBuilderChiContext(dummyType, writer, request, nil)
	})
	router.ServeHTTP(w, req)
	assert.NoError(t, err)
	assert.NotNil(t, v.Interface())
	ctx, ok := v.Interface().(chi.Context)
	assert.True(t, ok)
	assert.Equal(t, http.MethodGet, ctx.RouteMethod)
	assert.Equal(t, []string{"/foo"}, ctx.RoutePatterns)
}

func TestCommonBuilderHttpHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(hdrContentType, contentTypeJson)
	v, err := commonBuilderHttpHeader(dummyType, w, req, nil)
	assert.NoError(t, err)
	expect := http.Header{
		hdrContentType: []string{contentTypeJson},
	}
	assert.Equal(t, expect, v.Interface())
}

func TestCommonBuilderCookies(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderCookies(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok := v.Interface().([]*http.Cookie)
	assert.True(t, ok)
	assert.Equal(t, 0, len(av))

	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "foo",
		Value: "bar",
	})
	v, err = commonBuilderCookies(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok = v.Interface().([]*http.Cookie)
	assert.True(t, ok)
	assert.Equal(t, 1, len(av))
}

func TestCommonBuilderUrl(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/foo/bar?foo=bar", nil)
	v, err := commonBuilderUrl(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok := v.Interface().(*url.URL)
	assert.True(t, ok)
	assert.Equal(t, "/foo/bar", av.Path)
	assert.Equal(t, "foo=bar", av.RawQuery)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	req.URL = nil
	v, err = commonBuilderUrl(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok = v.Interface().(*url.URL)
	assert.True(t, ok)
	assert.Nil(t, av)
}

func TestCommonBuilderTypedHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(hdrContentType, contentTypeJson)
	v, err := commonBuilderTypedHeaders(dummyType, w, req, nil)
	assert.NoError(t, err)
	expect := Headers{
		hdrContentType: []string{contentTypeJson},
	}
	assert.Equal(t, expect, v.Interface())
}

func TestCommonBuilderMapHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(hdrContentType, contentTypeJson)
	v, err := commonBuilderMapHeaders(dummyType, w, req, nil)
	assert.NoError(t, err)
	expect := map[string][]string{
		hdrContentType: []string{contentTypeJson},
	}
	assert.Equal(t, expect, v.Interface())
}

func TestCommonBuilderTypedPathParams(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	params := []urit.PathVar{
		{
			Name:  "foo",
			Value: "bar1",
		},
		{
			Name:  "bar",
			Value: "1",
		},
		{
			Name:  "foo",
			Value: "bar2",
		},
	}
	v, err := commonBuilderTypedPathParams(dummyType, w, req, params)
	assert.NoError(t, err)
	expect := PathParams{
		"foo": []string{"bar1", "bar2"},
		"bar": []string{"1"},
	}
	assert.Equal(t, expect, v.Interface())
}

func TestCommonBuilderQueryParams(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderQueryParams(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Empty(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/xxx?foo=bar&bar=1", nil)
	v, err = commonBuilderQueryParams(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, v.Interface())
	assert.Equal(t, QueryParams{"foo": []string{"bar"}, "bar": []string{"1"}}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	req.URL = nil
	v, err = commonBuilderQueryParams(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok := v.Interface().(QueryParams)
	assert.True(t, ok)
	assert.Empty(t, av)
}

func TestCommonBuilderRawQuery(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderRawQuery(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Empty(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/xxx?foo=bar&bar=1", nil)
	v, err = commonBuilderRawQuery(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, v.Interface())
	assert.Equal(t, RawQuery("foo=bar&bar=1"), v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	req.URL = nil
	v, err = commonBuilderRawQuery(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok := v.Interface().(RawQuery)
	assert.True(t, ok)
	assert.Empty(t, av)
}

func TestCommonBuilderJsonRawMessage(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderJsonRawMessage(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Empty(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader("foo"))
	v, err = commonBuilderJsonRawMessage(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, v.Interface())
	assert.Equal(t, json.RawMessage{'f', 'o', 'o'}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/", &testErrorReader{})
	_, err = commonBuilderJsonRawMessage(dummyType, w, req, nil)
	assert.Error(t, err)
}

func TestCommonBuilderPostForm(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderPostForm(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Empty(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/", nil)
	_, err = commonBuilderPostForm(dummyType, w, req, nil)
	assert.Error(t, err)
	assert.Equal(t, "missing form body", err.Error())
}

func TestCommonBuilderPostFormEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderPostFormEmpty(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Empty(t, v.Interface())
}

const (
	basicGood = "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
	basicBad  = "Basic 123"
)

func TestCommonBuilderBasicAuth(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderBasicAuth(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok := v.Interface().(BasicAuth)
	assert.True(t, ok)
	assert.False(t, av.Ok)
	assert.Equal(t, "", av.Username)
	assert.Equal(t, "", av.Password)

	req.Header.Set("Authorization", basicBad)
	v, err = commonBuilderBasicAuth(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok = v.Interface().(BasicAuth)
	assert.True(t, ok)
	assert.False(t, av.Ok)
	assert.Equal(t, "", av.Username)
	assert.Equal(t, "", av.Password)

	req.Header.Set("Authorization", basicGood)
	v, err = commonBuilderBasicAuth(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok = v.Interface().(BasicAuth)
	assert.True(t, ok)
	assert.True(t, av.Ok)
	assert.Equal(t, "username", av.Username)
	assert.Equal(t, "password", av.Password)
}

func TestCommonBuilderBasicAuthPtr(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderBasicAuthPtr(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Nil(t, v.Interface())

	req.Header.Set("Authorization", "not basic")
	v, err = commonBuilderBasicAuthPtr(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Nil(t, v.Interface())

	req.Header.Set("Authorization", basicBad)
	v, err = commonBuilderBasicAuthPtr(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok := v.Interface().(*BasicAuth)
	assert.True(t, ok)
	assert.False(t, av.Ok)
	assert.Equal(t, "", av.Username)
	assert.Equal(t, "", av.Password)

	req.Header.Set("Authorization", basicGood)
	v, err = commonBuilderBasicAuthPtr(dummyType, w, req, nil)
	assert.NoError(t, err)
	av, ok = v.Interface().(*BasicAuth)
	assert.True(t, ok)
	assert.True(t, av.Ok)
	assert.Equal(t, "username", av.Username)
	assert.Equal(t, "password", av.Password)
}

func TestCommonBuilderByteBody(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := commonBuilderByteBody(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Empty(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader("foo"))
	v, err = commonBuilderByteBody(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, v.Interface())
	assert.Equal(t, []byte{'f', 'o', 'o'}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/", &testErrorReader{})
	_, err = commonBuilderByteBody(dummyType, w, req, nil)
	assert.Error(t, err)
}

func TestNewPathParamBuilder(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	fn := newPathParamBuilder(0)
	v, err := fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", v.Interface())

	params := []urit.PathVar{
		{
			Name:  "foo",
			Value: "bar",
		},
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	fn = newPathParamBuilder(0)
	v, err = fn(dummyType, w, req, params)
	assert.NoError(t, err)
	assert.Equal(t, "bar", v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	fn = newPathParamBuilder(1)
	v, err = fn(dummyType, w, req, params)
	assert.NoError(t, err)
	assert.Equal(t, "", v.Interface())
}

func TestNewPathParamsBuilder(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	fn := newPathParamsBuilder(0)
	v, err := fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, v.Interface())

	params := []urit.PathVar{
		{
			Name:  "foo",
			Value: "1",
		},
		{
			Name:  "bar",
			Value: "2",
		},
		{
			Name:  "baz",
			Value: "3",
		},
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	fn = newPathParamsBuilder(0)
	v, err = fn(dummyType, w, req, params)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	fn = newPathParamsBuilder(1)
	v, err = fn(dummyType, w, req, params)
	assert.NoError(t, err)
	assert.Equal(t, []string{"2", "3"}, v.Interface())
}

func TestNewStructParamBuilder(t *testing.T) {
	arg := reflect.TypeOf(testRequest{})
	fn := newStructParamBuilder(arg, defaultUnmarshaler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, testRequest{}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name": "foo"}`))
	v, err = fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, testRequest{Name: "foo"}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name": 1}`))
	_, err = fn(dummyType, w, req, nil)
	assert.Error(t, err)
}

func TestNewStructPtrParamBuilder(t *testing.T) {
	arg := reflect.TypeOf(&testRequest{})
	fn := newStructPtrParamBuilder(arg, defaultUnmarshaler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Nil(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name": "foo"}`))
	v, err = fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, &testRequest{Name: "foo"}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`null`))
	v, err = fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, &testRequest{}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"name": 1}`))
	_, err = fn(dummyType, w, req, nil)
	assert.Error(t, err)
}

func TestNewSliceParamBuilder(t *testing.T) {
	arg := reflect.TypeOf([]testRequest{})
	fn := newSliceParamBuilder(arg, defaultUnmarshaler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	v, err := fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Nil(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`[{"name": "foo"}]`))
	v, err = fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Equal(t, []testRequest{{Name: "foo"}}, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`null`))
	v, err = fn(dummyType, w, req, nil)
	assert.NoError(t, err)
	assert.Nil(t, v.Interface())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`[{"name": 1}]`))
	_, err = fn(dummyType, w, req, nil)
	assert.Error(t, err)
}

func TestChiRouteParams(t *testing.T) {
	const path = "/foo/{id1}/bar/{id2}/{id3}"
	var rcvdPathParams PathParams
	fn := func(pps PathParams) {
		rcvdPathParams = pps
	}
	b := NewTypedMethodsHandlerBuilder()
	hf, err := b.BuildHandler(path, http.MethodGet, chioas.Method{Handler: fn}, nil)
	assert.NoError(t, err)

	// test with urit fallback...
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/foo/1/bar/2/3", nil)
	hf.ServeHTTP(w, r)
	assert.Equal(t, 3, len(rcvdPathParams))

	// test with chi url params...
	router := chi.NewRouter()
	router.Get(path, hf)
	w = httptest.NewRecorder()
	r, _ = http.NewRequest(http.MethodGet, "/foo/1/bar/2/3", nil)
	rcvdPathParams = PathParams{}
	router.ServeHTTP(w, r)
	assert.Equal(t, 3, len(rcvdPathParams))
}

type testRequest struct {
	Name string `json:"name"`
}

type testErrorReader struct {
}

func (t *testErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("foo")
}
