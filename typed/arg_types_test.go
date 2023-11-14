package typed

import (
	"github.com/go-andiamo/chioas"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMultipartFormArgBuilder(t *testing.T) {
	var storedForm *multipart.Form
	called := false
	mdef := chioas.Method{
		Handler: func(mf *multipart.Form) {
			called = true
			storedForm = mf
		},
	}
	b := NewTypedMethodsHandlerBuilder()
	_, err := b.BuildHandler("/", http.MethodGet, mdef, nil)
	assert.Error(t, err)
	assert.Equal(t, "error building in args (path: /, method: GET) - cannot determine arg 0", err.Error())

	b = NewTypedMethodsHandlerBuilder(NewMultipartFormArgSupport(10000, false))
	hf, err := b.BuildHandler("/", http.MethodGet, mdef, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)

	assert.False(t, called)
	assert.Nil(t, storedForm)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(hdrContentType, "multipart/form-data")
	res := httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.False(t, called)

	b = NewTypedMethodsHandlerBuilder(NewMultipartFormArgSupport(10000, true))
	hf, err = b.BuildHandler("/", http.MethodGet, mdef, nil)
	assert.NoError(t, err)
	assert.NotNil(t, hf)
	assert.False(t, called)
	assert.Nil(t, storedForm)
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(hdrContentType, "multipart/form-data")
	res = httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.True(t, called)
	assert.Nil(t, storedForm)

	const bodyForm = `--xxx
Content-Disposition: form-data; name="field1"

value1
--xxx
Content-Disposition: form-data; name="field2"

value2
--xxx
Content-Disposition: form-data; name="file"; filename="file"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
`
	req, _ = http.NewRequest(http.MethodGet, "/", io.NopCloser(strings.NewReader(bodyForm)))
	req.Header.Set(hdrContentType, "multipart/form-data; boundary=xxx")
	res = httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.True(t, called)
	assert.NotNil(t, storedForm)
	assert.Equal(t, 2, len(storedForm.Value))
	assert.Equal(t, 1, len(storedForm.File))
}
