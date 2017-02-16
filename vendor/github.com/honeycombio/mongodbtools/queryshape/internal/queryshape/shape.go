package queryshape

import (
	"sort"
	"strings"
)

// rough rules XXX(toshok) needs editing
// 1. if key is not an op:
//    1a. if value is a primitive, set value = 1
//    1b. if value is an aggregate, walk subtree flattening everything but ops and their values if necessary
// 2. if key is an op:
//    2a. if value is a primitive, set value = 1
//    2b. if value is a map, keep map + all keys, and process keys (starting at step 1)
//    2c. if value is a list, walk list.
//        2c1. if all values are primitive, set value = 1
//        2c2. if any values are maps/lists, keep map + all keys, and process keys (starting at step 1)

func GetQueryShape(q map[string]interface{}) string {
	if q_, ok := q["$query"].(map[string]interface{}); ok {
		return GetQueryShape(q_)
	}

	pruned := make(map[string]interface{})
	for k, v := range q {
		if k[0] == '$' {
			pruned[k] = flattenOp(v)
		} else {
			pruned[k] = flatten(v)
		}
	}

	// flatten pruned to a string, sorting keys alphabetically ($ coming before a/A)
	return serializeShape(pruned)
}

func isAggregate(v interface{}) bool {
	if _, ok := v.([]interface{}); ok {
		return true
	} else if _, ok := v.(map[string]interface{}); ok {
		return true
	}
	return false
}

func flattenSlice(slice []interface{}, fromOp bool) interface{} {
	var rv []interface{}
	for _, v := range slice {
		if s, ok := v.([]interface{}); ok {
			sv := flattenSlice(s, false)
			if isAggregate(sv) {
				rv = append(rv, sv)
			}
		} else if m, ok := v.(map[string]interface{}); ok {
			mv := flattenMap(m, fromOp)
			if isAggregate(mv) || fromOp {
				rv = append(rv, mv)
			}
		}
	}
	// if the slice is empty, return 1 (since it's entirely primitives).
	// otherwise return the slice
	if len(rv) == 0 {
		return 1
	}
	return rv
}

func flattenMap(m map[string]interface{}, fromOp bool) interface{} {
	rv := make(map[string]interface{})
	for k, v := range m {
		if k[0] == '$' {
			rv[k] = flattenOp(v)
		} else {
			flattened := flatten(v)
			if isAggregate(flattened) || fromOp {
				rv[k] = flattened
			}
		}
	}
	// if the slice is empty, return 1 (since it's entirely primitives).
	// otherwise return the slice
	if len(rv) == 0 {
		return 1
	}
	return rv
}

func flatten(v interface{}) interface{} {
	if s, ok := v.([]interface{}); ok {
		return flattenSlice(s, false)
	} else if m, ok := v.(map[string]interface{}); ok {
		return flattenMap(m, false)
	} else {
		return 1
	}
}

func flattenOp(v interface{}) interface{} {
	if s, ok := v.([]interface{}); ok {
		return flattenSlice(s, true)
	} else if m, ok := v.(map[string]interface{}); ok {
		return flattenMap(m, true)
	} else {
		return 1
	}
}

func serializeShape(shape interface{}) string {
	// we can't just json marshal, since we need ordered keys
	if m, ok := shape.(map[string]interface{}); ok {
		var keys []string
		var keyAndVal []string
		for k := range m {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			keyAndVal = append(keyAndVal, "\""+k+"\": "+serializeShape(m[k]))
		}

		return "{ " + strings.Join(keyAndVal, ", ") + " }"

	} else if s, ok := shape.([]interface{}); ok {
		var vals []string
		for _, v := range s {
			vals = append(vals, serializeShape(v))
		}
		return "[ " + strings.Join(vals, ", ") + " ]"
	} else {
		return "1"
	}
}
