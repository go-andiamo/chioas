package chioas

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"math"
)

// OasVersion is the default OAS version for docs
var OasVersion = "3.0.3"

// ApplyMiddlewares is a function that returns middlewares to be applied to a Path or the api root
//
// can be used on Path.ApplyMiddlewares and Definition.ApplyMiddlewares (for api root)
type ApplyMiddlewares func(thisApi any) chi.Middlewares

// DisablerFunc is a function that can be used by Path.Disabled
type DisablerFunc func() bool

const (
	root = "/"
)

func nilString(v string) (result any) {
	result = v
	if v == "" {
		result = nil
	}
	return
}

func nilBool(v bool) (result any) {
	result = v
	if !v {
		result = nil
	}
	return
}

func nilNumber(n json.Number) (result any) {
	if n != "" {
		if i, err := n.Int64(); err == nil {
			result = i
		} else if f, err := n.Float64(); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			result = n
		}
	}
	return
}

func nilUint(v uint) (result any) {
	result = v
	if v == 0 {
		result = nil
	}
	return
}
