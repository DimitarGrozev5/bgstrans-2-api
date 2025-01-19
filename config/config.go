package config

import "github.com/dimitargrozev5/bgstrans-2-api/util"

// Setup app
type App struct {
	// Is app in production
	InProduction bool `yaml:"inProduction"`

	// List of valid systems
	ValidCSs []string `yaml:"validCSs"`
	ValidHSs []string `yaml:"validHSs"`

	// Coordinate transformations
	CsGraph map[string]map[string][]CSTransformation `yaml:"csGraph"`

	// Height transformations
	HsGraph          map[string]map[string]HSTransformation `yaml:"hsGraph"`
	HTransformations TransformationMethods                  `yaml:"hTransformations"`
}

// TODO: This is definetly not the place for these types and methods
// CS transformation type
type CSTransformation struct {
	Border []struct {
		X float64 `yaml:"X"`
		Y float64 `yaml:"Y"`
	} `yaml:"Border"`
	X0 float64
	Y0 float64

	A00 float64 `yaml:"A00"`
	A10 float64 `yaml:"A10"`
	A01 float64 `yaml:"A01"`
	A20 float64 `yaml:"A20"`
	A11 float64 `yaml:"A11"`
	A02 float64 `yaml:"A02"`
	A30 float64 `yaml:"A30"`
	A21 float64 `yaml:"A21"`
	A12 float64 `yaml:"A12"`
	A03 float64 `yaml:"A03"`

	B00 float64 `yaml:"B00"`
	B10 float64 `yaml:"B10"`
	B01 float64 `yaml:"B01"`
	B20 float64 `yaml:"B20"`
	B11 float64 `yaml:"B11"`
	B02 float64 `yaml:"B02"`
	B30 float64 `yaml:"B30"`
	B21 float64 `yaml:"B21"`
	B12 float64 `yaml:"B12"`
	B03 float64 `yaml:"B03"`
}

// Check if point is in tranformation zone
func (p *CSTransformation) InZone(x, y float64) bool {

	// Track intersections
	intersections := 0

	for i := 0; i < len(p.Border); i++ {

		// Get first polygon segment point
		x1 := p.Border[i].X
		y1 := p.Border[i].Y

		// Get second polygon segment point
		var x2, y2 float64

		// If last segment
		if i+1 == len(p.Border) {
			x2 = p.Border[0].X
			y2 = p.Border[0].Y
		} else {
			x2 = p.Border[i+1].X
			y2 = p.Border[i+1].Y
		}

		// Calculate polygon line equation
		var k, m, xInt, yInt float64

		// If not horizontal
		if x2-x1 != 0 {

			// Calculate line parameters
			k = (y2 - y1) / (x2 - x1)
			m = y1 - k/x1

			// Calculate intersection point
			yInt = y
			xInt = (y - m) / k
		} else {
			yInt = y
			xInt = x
		}

		// Calculate helper distances
		d12 := util.Dist(x1, y1, x2, y2)
		d1i := util.Dist(x1, y1, xInt, yInt)
		d2i := util.Dist(x2, y2, xInt, yInt)

		// Check if the intersection point is on the polygon segment
		if d1i < d12 && d2i <= d12 && xInt > x {
			intersections++
		}
	}

	// Return result
	return intersections%2 == 1
}

// HS tranformation
type HSTransformation struct {
	Type      string  `yaml:"Type"`
	Name      string  `yaml:"Name"`
	Direction float64 `yaml:"Direction"`
}

// HS Tranformation methods
type TransformationMethods struct {
	Grid  map[string]HGridTransformation  `yaml:"gridTransformations"`
	Plane map[string]HPlaneTransformation `yaml:"planeTransformations"`
}

// Grid transformation
type HGridTransformation struct {
	DB       string  `yaml:"DB"`
	X0       float64 `yaml:"X0"`
	Y0       float64 `yaml:"Y0"`
	GridSize float64 `yaml:"GridSize"`
}

// Flat Plane tranformation
type HPlaneTransformation struct {
	X0 float64 `yaml:"X0"`
	Y0 float64 `yaml:"Y0"`
	A  float64 `yaml:"A"`
	B  float64 `yaml:"B"`
	C  float64 `yaml:"C"`
}
