//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

var trace func(msg string, args ...interface{})

func init() {
	trace = func(msg string, args ...interface{}) {}
}

func SetTraceFunction(fn func(msg string, args ...interface{})) {
	trace = fn
}
