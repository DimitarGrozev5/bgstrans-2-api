package transformations

import (
	"fmt"

	"github.com/dimitargrozev5/bgstrans-2-api/config"
)

// Type repository
type Repository struct {
	App      *config.App
	ValidCSs map[string]bool
	ValidHSs map[string]bool
	CSGraph  CSTransformationGraph
	HSGraph  HSTransformationGraph
}

// Define repo
var Repo Repository

// Setup repo
func Setup(a *config.App) {

	// Add app to repo
	Repo = Repository{
		App:      a,
		ValidCSs: map[string]bool{},
		ValidHSs: map[string]bool{},
		CSGraph:  CSTransformationGraph{data: a.CsGraph},
		HSGraph:  HSTransformationGraph{data: a.HsGraph, methods: a.HTransformations},
	}

	// Covert valid CSs to Repo
	for _, cs := range a.ValidCSs {
		Repo.ValidCSs[cs] = true
	}

	// Covert valid HSs to Repo
	for _, hs := range a.ValidHSs {
		Repo.ValidHSs[hs] = true
	}
}

// Get transformer
func GetTransformer(ics, ocs, ihs, ohs string) (Transformer, error) {

	// Validate input
	if _, ok := Repo.ValidCSs[ics]; !ok {
		return nil, fmt.Errorf("%s", "Invalid input CS")
	}
	if _, ok := Repo.ValidCSs[ocs]; !ok {
		return nil, fmt.Errorf("%s", "Invalid output CS")
	}
	if _, ok := Repo.ValidHSs[ihs]; !ok {
		return nil, fmt.Errorf("%s", "Invalid input HS")
	}
	if _, ok := Repo.ValidHSs[ohs]; !ok {
		return nil, fmt.Errorf("%s", "Invalid output HS")
	}

	// Set hs targets
	hsTargets := make(map[string]bool)
	hsTargets[ohs] = false

	// Find path from input HS to output HS
	hsPath, found := findPathGraph(Repo.HSGraph.data, ihs, [2]string{ohs})
	if !found {
		return nil, fmt.Errorf("Can't convert from %s to %s", ihs, ohs)
	}

	// Check if a grid transformation is in the path
	storeBgs := false

	// Get first system
	from := ihs

	// Traverse graph
	for {

		// Get node
		to, ok := hsPath[from]

		// Exit if not found
		if !ok {
			break
		}

		// Get params
		params, _ := Repo.HSGraph.Get(from, to[0])

		// If params are not grid base
		if params.Name != "grid" {

			// Update from
			from = to[0]

			// Continue
			continue
		}

		// Will need bgs coordiantes
		storeBgs = true

		// Exit
		break
	}

	// Set cs transformation targets
	csTargets := [2]string{ocs}

	// Add bgs if needed
	if storeBgs {
		csTargets[1] = "bgs-cad"
	}

	// Find path from input CS to output CS, going trough BGS if needed
	csPath, found := findPathGraph(Repo.CSGraph.data, ics, csTargets)
	if !found {
		return nil, fmt.Errorf("Can't convert from %s to %s", ics, ocs)
	}

	return &TransformerOutput{
		csPath:       csPath,
		hsPath:       hsPath,
		includesGrid: storeBgs,
		ics:          ics,
		ihs:          ihs,
		ocs:          ocs,
		ohs:          ohs,
	}, nil
}
