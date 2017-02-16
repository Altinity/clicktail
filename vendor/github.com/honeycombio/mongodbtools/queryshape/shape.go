package queryshape

import "github.com/honeycombio/mongodbtools/queryshape/internal/queryshape"

// GetQueryShape takes a query map (provided by the logparser) and returns the query shape serialized as a string
func GetQueryShape(q map[string]interface{}) string {
	return queryshape.GetQueryShape(q)
}
