package transformations

import "github.com/dimitargrozev5/bgstrans-2-api/config"

// CS transformation graph
type CSTransformationGraph struct {
	data map[string]map[string][]config.CSTransformation
}

// Get params
func (g *CSTransformationGraph) Get(from, to string) ([]config.CSTransformation, bool) {
	res, ok := g.data[from][to]
	return res, ok
}

// HS transformation graph
type HSTransformationGraph struct {
	data    map[string]map[string]config.HSTransformation
	methods config.TransformationMethods
}

// Get params
func (g *HSTransformationGraph) Get(from, to string) (config.HSTransformation, bool) {
	res, ok := g.data[from][to]
	return res, ok
}
