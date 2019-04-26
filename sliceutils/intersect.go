package sliceutils

import "reflect"

// Eqaute checks if two inputs are same
type Eqaute func(interface{}, interface{}) bool

// Intersect between two slices
// Complexity: O(n^2)
// Input a, b must be array otherwise it panicss
func Intersect(a, b interface{}, fn Eqaute) interface{} {
	set := make([]interface{}, 0)
	av := reflect.ValueOf(a)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		if contains(b, el, fn) {
			set = append(set, el)
		}
	}

	return set
}

func contains(a interface{}, e interface{}, fn Eqaute) bool {
	v := reflect.ValueOf(a)

	for i := 0; i < v.Len(); i++ {
		if fn(v.Index(i).Interface(), e) {
			return true
		}
	}
	return false
}
