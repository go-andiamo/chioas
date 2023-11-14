package typed

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultErrorHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	defaultErrorHandler.HandleError(res, req, errors.New("foo"))
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "foo", res.Body.String())

	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	defaultErrorHandler.HandleError(res, req, &testApiError{"fooey", http.StatusPaymentRequired})
	assert.Equal(t, http.StatusPaymentRequired, res.Code)
	assert.Equal(t, "fooey", res.Body.String())
	assert.Equal(t, "", res.Result().Header.Get(hdrContentType))

	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	defaultErrorHandler.HandleError(res, req, &testApiJsonError{"fooey", http.StatusPaymentRequired, false})
	assert.Equal(t, http.StatusPaymentRequired, res.Code)
	assert.Equal(t, "{\"$error\":\"fooey\"}", res.Body.String())
	assert.Equal(t, contentTypeJson, res.Result().Header.Get(hdrContentType))

	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	defaultErrorHandler.HandleError(res, req, &testApiJsonError{"fooey", http.StatusPaymentRequired, true})
	assert.Equal(t, http.StatusPaymentRequired, res.Code)
	assert.Equal(t, "fooey\nfailed to marshal", res.Body.String())
	assert.Equal(t, "", res.Result().Header.Get(hdrContentType))
}

type testApiError struct {
	errMsg     string
	statusCode int
}

func (e *testApiError) Error() string {
	return e.errMsg
}

func (e *testApiError) StatusCode() int {
	return e.statusCode
}

func (e *testApiError) Wrapped() error {
	return nil
}

type testApiJsonError struct {
	errMsg     string
	statusCode int
	errors     bool
}

func (e *testApiJsonError) Error() string {
	return e.errMsg
}

func (e *testApiJsonError) StatusCode() int {
	return e.statusCode
}

func (e *testApiJsonError) Wrapped() error {
	return nil
}

func (e *testApiJsonError) MarshalJSON() ([]byte, error) {
	if e.errors {
		return nil, errors.New("failed to marshal")
	}
	return json.Marshal(map[string]any{"$error": e.errMsg})
}
