//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw/pkg/context"
)

const (
	UNKNOWN = "unknown"
)

type closer interface {
	Close() error
}

func SafeClose(c closer) {
	err := c.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to close: %s", err)
	}
}

// Reads a supplied filepath and converts the contents to an integer. Returns
// -1 if there were file permissions or existence errors or if the contents
// could not be successfully converted to an integer. In any error, a warning
// message is printed to STDERR and -1 is returned.
func SafeIntFromFile(ctx *context.Context, path string) int {
	msg := "failed to read int from file: %s\n"
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		ctx.Warn(msg, err)
		return -1
	}
	contents := strings.TrimSpace(string(buf))
	res, err := strconv.Atoi(contents)
	if err != nil {
		ctx.Warn(msg, err)
		return -1
	}
	return res
}

// ConcatStrings concatenate strings in a larger one. This function
// addresses a very specific ghw use case. For a more general approach,
// just use strings.Join()
func ConcatStrings(items ...string) string {
	return strings.Join(items, "")
}
