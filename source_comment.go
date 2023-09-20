package chioas

import (
	"fmt"
	"runtime"
	"strings"
)

// SourceComment is a utility function that can be used to place a comment in the spec yaml informing
// where, in the source code, a particular spec item was defined.
// Example:
//
//		var myApi = chioas.Definition{
//		  Comment:         chioas.SourceComment("this is a static comment"),
//	   ...
func SourceComment(comments ...string) string {
	static := ""
	if len(comments) > 0 {
		static = strings.Join(comments, "\n") + "\n"
	}
	pcs := make([]uintptr, 1)
	runtime.Callers(2, pcs)
	fms := runtime.CallersFrames(pcs)
	fm, _ := fms.Next()
	fn := ""
	if fnp := strings.Split(fm.File, "/"); len(fnp) > 0 {
		fn = fnp[len(fnp)-1]
	}
	return static + fmt.Sprintf("source: %s:%d", fn, fm.Line)
}
