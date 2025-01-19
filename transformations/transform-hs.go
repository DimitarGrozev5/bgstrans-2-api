package transformations

import (
	"errors"
	"fmt"
	"math"

	"github.com/dimitargrozev5/bgstrans-2-api/config"
)

// Transform between evrs and balt
func planeInterpolation(params config.HPlaneTransformation, x, y, h float64, sign float64) (float64, error) {

	// Calculate correction
	u := params.A + (x-params.X0)*params.B + (y-params.Y0)*params.C

	// Return result
	return h + sign*u, nil
}

// Transform grid interpolation
func gridInterpolation(p config.HGridTransformation, x, y, h float64, sign float64, dataPoints map[string]float64) (float64, error) {

	// Store result
	xr := x
	yr := y

	// Get points from cache
	a, ok1 := dataPoints[makeOndulationVertexNameFromXY(p, x, y)]
	b, ok2 := dataPoints[makeOndulationVertexNameFromXY(p, x+p.GridSize, y)]
	c, ok3 := dataPoints[makeOndulationVertexNameFromXY(p, x, y+p.GridSize)]
	d, ok4 := dataPoints[makeOndulationVertexNameFromXY(p, x+p.GridSize, y+p.GridSize)]

	// Check if all points are found
	if !(ok1 && ok2 && ok3 && ok4) {
		return 0, errors.New("outside of bound")
	}

	// Determine grid square base point
	x0 := math.Floor((x-p.X0)/p.GridSize)*p.GridSize + p.X0
	y0 := math.Floor((y-p.Y0)/p.GridSize)*p.GridSize + p.Y0

	// Normalize coordinates
	xr = (xr - x0) / 100
	yr = (yr - y0) / 100

	// Calculate undulation
	u := a*(1-xr)*(1-yr) + b*xr*(1-yr) + c*(1-xr)*yr + d*xr*yr

	// Return result
	return h + sign*u, nil
}

// Calculate undulation vertice name from coordinates
func makeOndulationVertexNameFromXY(p config.HGridTransformation, x, y float64) string {
	return fmt.Sprintf(
		"%.0f/%.0f",
		math.Floor((x-p.X0)/p.GridSize),
		math.Floor((y-p.Y0)/p.GridSize),
	)
}
