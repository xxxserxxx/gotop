package devices

import (
	"fmt"
	"strings"
)

// makeName creates a prometheus metric name in the gotop space
// This function doesn't have to be very efficient because it's only
// called at init time, and only a few dozen times... and it isn't
// (very efficient).
func makeName(parts ...interface{}) string {
	args := make([]string, len(parts))
	for i, v := range parts {
		args[i] = fmt.Sprintf("%v", v)
	}
	rv := strings.Join(args, "_")
	rv = strings.ReplaceAll(rv, "-", ":")
	rv = strings.ReplaceAll(rv, " ", ":")
	return rv
}
