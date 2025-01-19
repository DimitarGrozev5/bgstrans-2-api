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
	hsPath := getPath(Repo.HSGraph.data, ihs, hsTargets, []string{})

	// Check if a grid transformation is in the path
	storeBgs := false

	// Get first system
	from := ihs

	// Iterate over hsPath
	for _, to := range hsPath {

		// Get params
		params, _ := Repo.HSGraph.Get(from, to)

		// Update from
		from = to

		// If params are not grid base, continue
		if params.Name != "grid" {
			continue
		}

		// Will need bgs coordiantes
		storeBgs = true

		// Exit
		break
	}

	// Set cs transformation targets
	csTargets := make(map[string]bool)
	csTargets[ocs] = false

	// Add bgs if needed
	if storeBgs {
		csTargets["bgs-cad"] = false
	}

	// Find path from input CS to output CS, going trough BGS if needed
	csPath := getPath(Repo.CSGraph.data, ics, csTargets, []string{})

	return &TransformerOutput{
		csPath:       csPath,
		hsPath:       hsPath,
		includesGrid: storeBgs,
		ics:          ics,
		ihs:          ihs,
	}, nil
}
