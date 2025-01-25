package transformations

import (
	"testing"

	"github.com/dimitargrozev5/bgstrans-2-api/config"
)

// Test validation
func TestValidation(t *testing.T) {

	// Setup app state
	app := config.App{
		ValidCSs: []string{"cs1", "cs2"},
		ValidHSs: []string{"hs1", "hs2"},
	}

	// Setup transformations
	Setup(&app)

	// Pass valid cs and hs
	_, err := GetTransformer("cs1", "cs2", "hs1", "hs2")
	if err != nil {
		t.Error("error when all systems are correct")
	}

	// Pass invalid ics
	_, err = GetTransformer("cs3", "cs2", "hs1", "hs2")
	if err == nil {
		t.Error("expected error for invalid ics")
	}

	// Pass invalid ocs
	_, err = GetTransformer("cs1", "cs3", "hs1", "hs2")
	if err == nil {
		t.Error("expected error for invalid ocs")
	}

	// Pass invalid ihs
	_, err = GetTransformer("cs1", "cs2", "hs3", "hs2")
	if err == nil {
		t.Error("expected error for invalid ihs")
	}

	// Pass invalid ohs
	_, err = GetTransformer("cs1", "cs1", "hs1", "hs3")
	if err == nil {
		t.Error("expected error for invalid ohs")
	}
}
