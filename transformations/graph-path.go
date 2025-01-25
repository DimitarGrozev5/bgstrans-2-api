package transformations

import (
	"slices"
)

// Generate distance graph
func distGraph[T any](graph map[string]map[string]T, from string) map[string]int {

	// Create resulting graph
	result := make(map[string]int)

	// Track visited nodes
	visited := map[string]bool{from: true}

	// Track current distance
	dist := 0

	// Track a list of nodes to visit next
	nodes := []string{from}

	// Infinate loop
	for {

		// Track next nodes
		var nextNodes []string

		// Iterate over nodes
		for _, node := range nodes {

			// Update node distance
			result[node] = dist

			// Iterate node connections
			for next := range graph[node] {

				// Skip if already visited
				if visited[next] {
					continue
				}

				// Add next node
				nextNodes = append(nextNodes, next)

				// Mark as visited
				visited[next] = true
			}
		}

		// If no next nodes, exit
		if len(nextNodes) == 0 {
			break
		}

		// Set next as current
		nodes = nextNodes

		// Update dist
		dist++
	}

	// Return
	return result
}

// Find path through distance graph
func findPath[T any](graph map[string]map[string]T, dists map[string]int, from, to string, path []string) ([]string, bool) {

	// If from is target, return
	if from == to {
		return path, true
	}

	// Get current distance
	dist := dists[from]

	// Iterate from connections
	for node := range graph[from] {

		// Skip if distance is not larger
		if dists[node] <= dist {
			continue
		}

		// Get path
		newPath, found := findPath(graph, dists, node, to, append(path, node))

		// Return if found
		if found {
			return newPath, true
		}
	}

	// Return path not found
	return path, false
}

// Find graph path, including up to two targets
func findPathGraph[T any](graph map[string]map[string]T, from string, targets [2]string) (map[string][]string, bool) {

	// If only one target
	if targets[1] == "" {

		// Build dist graph
		dg := distGraph(graph, from)

		// Build path
		path, found := findPath(graph, dg, from, targets[0], []string{})

		// If not found
		if !found {
			return nil, false
		}

		// Convert path to graph
		var res = make(map[string][]string)

		// Get start
		cn := from

		// Iterate path
		for _, node := range path {

			// Add link
			res[cn] = append(res[cn], node)

			// Update current node
			cn = node
		}

		// Return result
		return res, true
	}

	// If more than one target, find anchor point
	combinedDist := make(map[string][]int)
	var dg = [3]map[string]int{}

	// Build distance graph for the three points
	// We are assuming that all vertices in the graph are bi-directional
	for i, node := range []string{from, targets[0], targets[1]} {

		// Build dist graph
		dg[i] = distGraph(graph, node)

		// Iterate nodes
		for dgNode, val := range dg[i] {

			// Add value
			combinedDist[dgNode] = append(combinedDist[dgNode], val)
		}
	}

	// Track min dist
	minDist := 100000000
	minDistNode := ""

	// Iterate over dist nodes
	for node, dists := range combinedDist {

		// Get dist
		dist := dists[0] + dists[1] + dists[2]

		// Update min
		if dist < minDist {
			minDist = dist
			minDistNode = node
		}
	}

	// Find path from start, to minDistNode
	path, found := findPath(graph, dg[0], from, minDistNode, []string{})
	if !found {
		return nil, false
	}

	// Convert path to graph
	res := make(map[string][]string)

	// Get start
	cn := from

	// Iterate path
	for _, node := range path {

		// Add link
		res[cn] = append(res[cn], node)

		// Update current node
		cn = node
	}

	// Find path from target 1, to minDistNode
	path, found = findPath(graph, dg[1], targets[0], minDistNode, []string{})
	if !found {
		return nil, false
	}

	// If there is a path
	if len(path) > 0 {

		// Reverse path
		slices.Reverse(path)

		// Convert path to graph
		// Get start
		cn = minDistNode

		// Iterate path
		// Remove first node (minDistNode) and add the target to the end
		for _, node := range append(path[1:], targets[0]) {

			// Add link
			res[cn] = append(res[cn], node)

			// Update current node
			cn = node
		}
	}

	// Find path from target 1, to minDistNode
	path, found = findPath(graph, dg[2], targets[1], minDistNode, []string{})
	if !found {
		return nil, false
	}

	// If there is a path
	if len(path) > 0 {

		// Reverse path
		slices.Reverse(path)

		// Convert path to graph
		// Get start
		cn = minDistNode

		// Iterate path
		// Remove first node (minDistNode) and add the target to the end
		for _, node := range append(path[1:], targets[1]) {

			// Add link
			res[cn] = append(res[cn], node)

			// Update current node
			cn = node
		}
	}

	return res, true
}
