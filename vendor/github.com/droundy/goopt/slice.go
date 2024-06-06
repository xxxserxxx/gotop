package goopt

// Here we have some utility slice routines

// append appends an element to a slice, in-place if possible, and
// expanding if needed.
func append(slice *[]string, val string) {
	length := len(*slice)
	if cap(*slice) == length {
		// we need to expand
		newsl := make([]string, length, 2*(length+1))
		for i, v := range *slice {
			newsl[i] = v
		}
		*slice = newsl
	}
	*slice = (*slice)[0 : length+1]
	(*slice)[length] = val
}

// cat concatenates two slices, expanding if needed.
func cat(slices ...[]string) []string {
	return cats(slices)
}

// cats concatenates several slices, expanding if needed.
func cats(slices [][]string) []string {
	lentot := 0
	for _, sl := range slices {
		lentot += len(sl)
	}
	out := make([]string, lentot)
	i := 0
	for _, sl := range slices {
		for _, v := range sl {
			out[i] = v
			i++
		}
	}
	return out
}

func any(f func(string) bool, slice []string) bool {
	for _, v := range slice {
		if f(v) {
			return true
		}
	}
	return false
}
