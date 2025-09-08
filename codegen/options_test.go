package codegen

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGoGenOpts_varName(t *testing.T) {
	testCases := []struct {
		options Options
		kind    string
		name    string
		expect  string
	}{
		{
			options: Options{},
			expect:  "x",
		},
		{
			options: Options{PublicVars: true},
			expect:  "X",
		},
		{
			options: Options{},
			name:    "0",
			expect:  "x0",
		},
		{
			options: Options{PublicVars: true},
			name:    "0",
			expect:  "X0",
		},
		{
			options: Options{},
			kind:    "request",
			name:    "foo",
			expect:  "requestFoo",
		},
		{
			options: Options{PublicVars: true},
			kind:    "request",
			name:    "foo",
			expect:  "RequestFoo",
		},
		{
			options: Options{},
			kind:    "request",
			name:    "",
			expect:  "requestX",
		},
		{
			options: Options{PublicVars: true},
			kind:    "request",
			name:    "",
			expect:  "RequestX",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := tc.options.varName(tc.kind, tc.name)
			assert.Equal(t, tc.expect, n)
		})
	}
}

func TestGoGenOpts_topVarName(t *testing.T) {
	testCases := []struct {
		options Options
		expect  string
	}{
		{
			options: Options{},
			expect:  "definition",
		},
		{
			options: Options{PublicVars: true},
			expect:  "Definition",
		},
		{
			options: Options{VarName: "foo"},
			expect:  "foo",
		},
		{
			options: Options{VarName: "foo", PublicVars: true},
			expect:  "foo",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := tc.options.topVarName()
			assert.Equal(t, tc.expect, n)
		})
	}
}

func Test_toPascal(t *testing.T) {
	testCases := []struct {
		name   string
		expect string
	}{
		{
			name:   "",
			expect: "",
		},
		{
			name:   "_",
			expect: "",
		},
		{
			name:   "foo",
			expect: "Foo",
		},
		{
			name:   "foo bar",
			expect: "FooBar",
		},
		{
			name:   "foo_bar",
			expect: "FooBar",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := toPascal(tc.name)
			assert.Equal(t, tc.expect, n)
		})
	}
}

func TestGoGenOpts_translateMethod(t *testing.T) {
	testCases := []struct {
		options Options
		method  string
		expect  string
	}{
		{
			options: Options{},
			method:  http.MethodGet,
			expect:  "\"GET\"",
		},
		{
			options: Options{UseHttpConsts: true},
			method:  http.MethodGet,
			expect:  "http.MethodGet",
		},
		{
			options: Options{UseHttpConsts: true},
			method:  "foo",
			expect:  "\"FOO\"",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := tc.options.translateMethod(tc.method)
			assert.Equal(t, tc.expect, n)
		})
	}
}

func TestGoGenOpts_translateStatus(t *testing.T) {
	testCases := []struct {
		options Options
		status  int
		expect  string
	}{
		{
			options: Options{},
			status:  http.StatusOK,
			expect:  "200",
		},
		{
			options: Options{UseHttpConsts: true},
			status:  http.StatusOK,
			expect:  "http.StatusOK",
		},
		{
			options: Options{UseHttpConsts: true},
			status:  999,
			expect:  "999",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			n := tc.options.translateStatus(tc.status)
			assert.Equal(t, tc.expect, n)
		})
	}
}
