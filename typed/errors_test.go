package typed

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewApiError(t *testing.T) {
	err := NewApiError(0, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	assert.Equal(t, "Internal Server Error", err.Error())
	assert.Nil(t, err.Wrapped())

	err = NewApiError(http.StatusPaymentRequired, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "Payment Required", err.Error())
	assert.Nil(t, err.Wrapped())

	err = NewApiError(http.StatusPaymentRequired, "fooey")
	assert.Error(t, err)
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "fooey", err.Error())
	assert.Nil(t, err.Wrapped())
}

func TestNewApiErrorf(t *testing.T) {
	err := NewApiErrorf(0, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	assert.Equal(t, "Internal Server Error", err.Error())
	assert.Nil(t, err.Wrapped())

	err = NewApiErrorf(http.StatusPaymentRequired, "")
	assert.Error(t, err)
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "Payment Required", err.Error())
	assert.Nil(t, err.Wrapped())

	err = NewApiErrorf(http.StatusPaymentRequired, "fooey")
	assert.Error(t, err)
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "fooey", err.Error())
	assert.Nil(t, err.Wrapped())

	err = NewApiErrorf(http.StatusPaymentRequired, "fooey %d", 16)
	assert.Error(t, err)
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "fooey 16", err.Error())
	assert.Nil(t, err.Wrapped())
}

func TestWrapApiError(t *testing.T) {
	err := WrapApiError(0, nil)
	assert.NoError(t, err)

	err = WrapApiError(0, errors.New("fooey"))
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	assert.Equal(t, "fooey", err.Error())
	assert.NotNil(t, err.Wrapped())

	err = WrapApiError(http.StatusPaymentRequired, errors.New(""))
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "Payment Required", err.Error())
	assert.NotNil(t, err.Wrapped())
}

func TestWrapApiErrorMsg(t *testing.T) {
	err := WrapApiErrorMsg(0, nil, "whoops")
	assert.NoError(t, err)

	err = WrapApiErrorMsg(0, errors.New("fooey"), "whoops")
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	assert.Equal(t, "whoops", err.Error())
	assert.NotNil(t, err.Wrapped())

	err = WrapApiErrorMsg(http.StatusPaymentRequired, errors.New(""), "whoops")
	assert.Equal(t, http.StatusPaymentRequired, err.StatusCode())
	assert.Equal(t, "whoops", err.Error())
	assert.NotNil(t, err.Wrapped())
}
