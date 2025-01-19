package transformations

// Find path through graph
// Returns the path to target, not including the start
func getPath[T any](graph map[string]map[string]T, current string, targets map[string]bool, path []string) []string {

	// Mark if on target
	if _, in := targets[current]; in {
		targets[current] = true
	}

	// Check if all targets are reached
	reachedAll := true

	// Iterate over targets
	for _, reached := range targets {

		// Continue if reached
		if reached {
			continue
		}

		// Switch flag
		reachedAll = false

		// Exit loop
		break
	}

	// Return if reached all
	if reachedAll {
		return append(path, current)
	}

	// Track shortest path length
	shortestPathLen := 1000000

	// Track shortest path
	shortestPath := []string{}

	// Iterate over connections
	for nextNode := range graph[current] {

		// Exit if node is already in the path
		visited := false

		// Iterate current path, to find if next node is visited
		for _, node := range path {
			if node == nextNode {
				visited = true
			}
		}

		// Skip if the node is already visited
		if visited {
			continue
		}

		// Get path from this node
		newPath := getPath(graph, nextNode, targets, path)

		// Skip if path is empty
		if len(newPath) == 0 {
			continue
		}

		// If the new path is shorter
		if len(newPath) <= shortestPathLen {
			shortestPathLen = len(newPath)
			shortestPath = newPath
		}
	}

	// Return path
	return shortestPath
}
