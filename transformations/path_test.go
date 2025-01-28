package transformations

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/dimitargrozev5/bgstrans-2-api/config"
)

/*
 * Define graphs for testing traversing
 * Have the following structure for height graph
 * hs1---hs2---[-]---hs3
 * .......|...........|.
 * ......hs4...hs5...hs6
 * ........\....|..../..
 * .........\--hs7-[/]..
 */
var mockPathState config.App = config.App{
	ValidCSs: []string{"cs1", "cs2"},
	ValidHSs: []string{"hs1", "hs2"},
	CsGraph: map[string]map[string][]config.CSTransformation{
		"cs1": {
			"cs2": []config.CSTransformation{
				{
					Border: []struct {
						X float64 `yaml:"X"`
						Y float64 `yaml:"Y"`
					}{
						{0, 0},
						{0, 10},
						{10, 10},
						{10, 0},
					},
					X0: 10,
					Y0: 10,
				},
			},
		},
		"cs2": {
			"cs1": []config.CSTransformation{
				{
					Border: []struct {
						X float64 `yaml:"X"`
						Y float64 `yaml:"Y"`
					}{
						{10, 10},
						{10, 20},
						{20, 20},
						{20, 10},
					},
					X0: -10,
					Y0: -10,
				},
			},
		},
	},
	HsGraph: map[string]map[string]config.HSTransformation{
		"hs1": {
			"hs2": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr12",
				Direction: 1,
			},
		},
		"hs2": {
			"hs1": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr12",
				Direction: -1,
			},
			"hs3": config.HSTransformation{
				Type:      "grid",
				Name:      "gtr23",
				Direction: 1,
			},
			"hs4": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr24",
				Direction: 1,
			},
		},
		"hs3": {
			"hs2": config.HSTransformation{
				Type:      "grid",
				Name:      "gtr23",
				Direction: -1,
			},
			"hs6": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr36",
				Direction: 1,
			},
		},
		"hs4": {
			"hs2": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr24",
				Direction: -1,
			},
			"hs7": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr47",
				Direction: 1,
			},
		},
		"hs5": {
			"hs7": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr57",
				Direction: 1,
			},
		},
		"hs6": {
			"hs3": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr36",
				Direction: -1,
			},
			"hs7": config.HSTransformation{
				Type:      "grid",
				Name:      "gtr67",
				Direction: 1,
			},
		},
		"hs7": {
			"hs4": config.HSTransformation{
				Type:      "plane",
				Name:      "ptr47",
				Direction: -1,
			},
			"hs5": config.HSTransformation{
				Type:      "grid",
				Name:      "gtr57",
				Direction: -1,
			},
			"hs6": config.HSTransformation{
				Type:      "grid",
				Name:      "gtr67",
				Direction: -1,
			},
		},
	},
	HTransformations: config.TransformationMethods{
		Grid: map[string]config.HGridTransformation{
			"gtr23": defaultGrid,
			"gtr67": defaultGrid,
		},
		Plane: map[string]config.HPlaneTransformation{
			"ptr12": defaultPlanar,
			"ptr24": defaultPlanar,
			"ptr36": defaultPlanar,
			"ptr47": defaultPlanar,
			"ptr57": defaultPlanar,
		},
	},
}

// Default planar transformation
var defaultPlanar = config.HPlaneTransformation{
	X0: 0,
	Y0: 0,
	A:  0,
	B:  0,
	C:  0,
}

// Default planar transformation
var defaultGrid = config.HGridTransformation{
	DB:       "dbpath",
	X0:       0,
	Y0:       0,
	GridSize: 100,
}

func expect(t *testing.T, from, to string, val, res int) {
	if val != res {
		t.Errorf("Expected %d, received %d on dist from %s to %s", res, val, from, to)
	}
}

// Test distance builder
func TestDist(t *testing.T) {

	// Generate dist map
	res := distGraph(mockPathState.HsGraph, "hs1")

	// Test
	expect(t, "hs1", "hs1", res["hs1"], 0)
	expect(t, "hs1", "hs2", res["hs2"], 1)
	expect(t, "hs1", "hs3", res["hs3"], 2)
	expect(t, "hs1", "hs4", res["hs4"], 2)
	expect(t, "hs1", "hs6", res["hs6"], 3)
	expect(t, "hs1", "hs7", res["hs7"], 3)
	expect(t, "hs1", "hs5", res["hs5"], 4)

	// Generate dist map
	res = distGraph(mockPathState.HsGraph, "hs2")

	// Test
	expect(t, "hs2", "hs1", res["hs1"], 1)
	expect(t, "hs2", "hs3", res["hs3"], 1)
	expect(t, "hs2", "hs4", res["hs4"], 1)
	expect(t, "hs2", "hs6", res["hs6"], 2)
	expect(t, "hs2", "hs7", res["hs7"], 2)
	expect(t, "hs2", "hs5", res["hs5"], 3)

	// Generate dist map
	res = distGraph(mockPathState.HsGraph, "hs3")

	// Test
	expect(t, "hs3", "hs2", res["hs2"], 1)
	expect(t, "hs3", "hs6", res["hs6"], 1)
	expect(t, "hs3", "hs1", res["hs1"], 2)
	expect(t, "hs3", "hs4", res["hs4"], 2)
	expect(t, "hs3", "hs7", res["hs7"], 2)
	expect(t, "hs3", "hs5", res["hs5"], 3)

	// Generate dist map
	res = distGraph(mockPathState.HsGraph, "hs4")

	// Test
	expect(t, "hs4", "hs2", res["hs2"], 1)
	expect(t, "hs4", "hs7", res["hs7"], 1)
	expect(t, "hs4", "hs1", res["hs1"], 2)
	expect(t, "hs4", "hs3", res["hs3"], 2)
	expect(t, "hs4", "hs5", res["hs5"], 2)
	expect(t, "hs4", "hs6", res["hs6"], 2)

	// Generate dist map
	res = distGraph(mockPathState.HsGraph, "hs5")

	// Test
	expect(t, "hs5", "hs7", res["hs7"], 1)
	expect(t, "hs5", "hs4", res["hs4"], 2)
	expect(t, "hs5", "hs6", res["hs6"], 2)
	expect(t, "hs5", "hs2", res["hs2"], 3)
	expect(t, "hs5", "hs3", res["hs3"], 3)
	expect(t, "hs5", "hs1", res["hs1"], 4)

	// Generate dist map
	res = distGraph(mockPathState.HsGraph, "hs6")

	// Test
	expect(t, "hs6", "hs3", res["hs3"], 1)
	expect(t, "hs6", "hs7", res["hs7"], 1)
	expect(t, "hs6", "hs2", res["hs2"], 2)
	expect(t, "hs6", "hs4", res["hs4"], 2)
	expect(t, "hs6", "hs5", res["hs5"], 2)
	expect(t, "hs6", "hs1", res["hs1"], 3)

	// Generate dist map
	res = distGraph(mockPathState.HsGraph, "hs7")

	// Test
	expect(t, "hs7", "hs4", res["hs4"], 1)
	expect(t, "hs7", "hs5", res["hs5"], 1)
	expect(t, "hs7", "hs6", res["hs6"], 1)
	expect(t, "hs7", "hs2", res["hs2"], 2)
	expect(t, "hs7", "hs3", res["hs3"], 2)
	expect(t, "hs7", "hs1", res["hs1"], 3)
}

// Test path
func TestPath(t *testing.T) {

	// Define test cases
	type Case struct {
		start    string
		target   string
		expected []string
	}
	cases := []Case{
		{
			start:    "hs1",
			target:   "hs1",
			expected: []string{},
		},
		{
			start:    "hs1",
			target:   "hs2",
			expected: []string{"hs2"},
		},
		{
			start:    "hs1",
			target:   "hs3",
			expected: []string{"hs2", "hs3"},
		},
		{
			start:    "hs1",
			target:   "hs4",
			expected: []string{"hs2", "hs4"},
		},
		{
			start:    "hs1",
			target:   "hs5",
			expected: []string{"hs2", "hs4", "hs7", "hs5"},
		},
		{
			start:    "hs1",
			target:   "hs6",
			expected: []string{"hs2", "hs3", "hs6"},
		},
		{
			start:    "hs1",
			target:   "hs7",
			expected: []string{"hs2", "hs4", "hs7"},
		},
		{
			start:    "hs3",
			target:   "hs1",
			expected: []string{"hs2", "hs1"},
		},
		{
			start:    "hs3",
			target:   "hs2",
			expected: []string{"hs2"},
		},
		{
			start:    "hs3",
			target:   "hs4",
			expected: []string{"hs2", "hs4"},
		},
		{
			start:    "hs3",
			target:   "hs5",
			expected: []string{"hs6", "hs7", "hs5"},
		},
		{
			start:    "hs3",
			target:   "hs6",
			expected: []string{"hs6"},
		},
		{
			start:    "hs3",
			target:   "hs7",
			expected: []string{"hs6", "hs7"},
		},
		{
			start:    "hs7",
			target:   "hs1",
			expected: []string{"hs4", "hs2", "hs1"},
		},
		{
			start:    "hs7",
			target:   "hs2",
			expected: []string{"hs4", "hs2"},
		},
		{
			start:    "hs7",
			target:   "hs3",
			expected: []string{"hs6", "hs3"},
		},
		{
			start:    "hs7",
			target:   "hs4",
			expected: []string{"hs4"},
		},
		{
			start:    "hs7",
			target:   "hs5",
			expected: []string{"hs5"},
		},
		{
			start:    "hs7",
			target:   "hs6",
			expected: []string{"hs6"},
		},
	}

	// Run cases
	for _, c := range cases {

		// Get dist graph
		dg := distGraph(mockPathState.HsGraph, c.start)

		// Run case
		res, found := findPath(mockPathState.HsGraph, dg, c.start, c.target, []string{})
		if !found {
			t.Errorf("Path not found from %s to %s", c.start, c.target)
			return
		}

		// Distance be the same
		if len(res) != len(c.expected) {
			t.Errorf("From %s to %s; Expected %s; Received %s", c.start, c.target, strings.Join(c.expected, ", "), strings.Join(res, ", "))
			return

		}

		// Iterate results
		for i := 0; i < len(res); i++ {

			// If different
			if res[i] != c.expected[i] {
				t.Errorf("From %s to %s; Expected %s; Received %s", c.start, c.target, strings.Join(c.expected, ", "), strings.Join(res, ", "))
				return
			}
		}
	}
}

// Test path graph
func TestPathGraph(t *testing.T) {

	// Define test cases
	type Case struct {
		start    string
		target   [2]string
		expected map[string][]string
	}
	cases := []Case{
		{
			start:    "hs1",
			target:   [2]string{"hs1"},
			expected: map[string][]string{},
		},
		{
			start:  "hs1",
			target: [2]string{"hs2"},
			expected: map[string][]string{
				"hs1": {"hs2"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs6"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs3"},
				"hs3": {"hs6"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs5"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs4"},
				"hs4": {"hs7"},
				"hs7": {"hs5"},
			},
		},
		{
			start:  "hs6",
			target: [2]string{"hs4"},
			expected: map[string][]string{
				"hs6": {"hs7"},
				"hs7": {"hs4"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs2", "hs3"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs3"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs2", "hs5"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs4"},
				"hs4": {"hs7"},
				"hs7": {"hs5"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs7", "hs6"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs3"},
				"hs3": {"hs6"},
				"hs6": {"hs7"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs3", "hs4"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs3", "hs4"},
			},
		},
		{
			start:  "hs1",
			target: [2]string{"hs4", "hs5"},
			expected: map[string][]string{
				"hs1": {"hs2"},
				"hs2": {"hs4"},
				"hs4": {"hs7"},
				"hs7": {"hs5"},
			},
		},
		{
			start:  "hs4",
			target: [2]string{"hs5", "hs3"},
			expected: map[string][]string{
				"hs4": {"hs7", "hs2"},
				"hs7": {"hs5"},
				"hs2": {"hs3"},
			},
		},
		{
			start:  "hs2",
			target: [2]string{"hs2", "hs3"},
			expected: map[string][]string{
				"hs2": {"hs3"},
			},
		},
	}

	/*
	 * Define graphs for testing traversing
	 * Have the following structure for height graph
	 * hs1---hs2---------hs3
	 * .......|...........|.
	 * ......hs4...hs5...hs6
	 * ........\....|..../..
	 * .........\--hs7--/...
	 */

	// Run cases
	for _, c := range cases {

		// Run case
		res, found := findPathGraph(mockPathState.HsGraph, c.start, c.target)
		if !found {
			t.Errorf("Path not found from %s to %s", c.start, c.target)
			return
		}

		// Check if the lengths of the maps are different
		if len(res) != len(c.expected) {
			t.Errorf("From %s to %s; Expected %v; Received %v", c.start, c.target, c.expected, res)
			return
		}

		// Iterate over keys in res
		for key, slice1 := range res {

			// Get slice
			slice2, exists := c.expected[key]

			// Check if the key exists in c.expected
			if !exists {
				t.Errorf("From %s to %s; Expected %v; Received %v", c.start, c.target, c.expected, res)
				return
			}

			// Check if the slices have the same length
			if len(slice1) != len(slice2) {
				t.Errorf("From %s to %s; Expected %v; Received %v", c.start, c.target, c.expected, res)
				return
			}

			// Sort the slices to compare them
			sortedSlice1 := append([]string{}, slice1...) // Create a copy to avoid modifying the original
			sortedSlice2 := append([]string{}, slice2...)
			sort.Strings(sortedSlice1)
			sort.Strings(sortedSlice2)

			// Check if the sorted slices are equal
			if !reflect.DeepEqual(sortedSlice1, sortedSlice2) {
				t.Errorf("From %s to %s; Expected %v; Received %v", c.start, c.target, c.expected, res)
				return
			}
		}
	}
}
