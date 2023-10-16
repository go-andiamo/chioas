package chioas

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNilString(t *testing.T) {
	v := nilString("")
	assert.Nil(t, v)
	v = nilString("foo")
	assert.NotNil(t, v)
}

func TestNilBool(t *testing.T) {
	v := nilBool(false)
	assert.Nil(t, v)
	v = nilBool(true)
	assert.NotNil(t, v)
}

func TestNilNumber(t *testing.T) {
	v := nilNumber("")
	assert.Nil(t, v)
	v = nilNumber("NaN")
	assert.Nil(t, v)
	v = nilNumber("Inf")
	assert.Nil(t, v)
	v = nilNumber("1")
	assert.NotNil(t, v)
	assert.Equal(t, int64(1), v)
	v = nilNumber("1.1")
	assert.NotNil(t, v)
	assert.Equal(t, json.Number("1.1"), v)
}

func TestNilUint(t *testing.T) {
	v := nilUint(0)
	assert.Nil(t, v)
	v = nilUint(1)
	assert.NotNil(t, v)
	assert.Equal(t, uint(1), v)
}
