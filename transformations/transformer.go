package transformations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Transformer interface
type Transformer interface {
	Add(id int, pt *PointResult)
	TransformBatch() (map[int]*PointResult, error)
}

// Store transformation intermediate steps
type PointResult struct {
	Name string

	X     float64
	Y     float64
	XYErr string

	H    float64
	HasH bool
	HErr string

	Xbgs float64
	Ybgs float64

	Var []string
}

// Transform output type
type TransformerOutput struct {
	csPath       map[string][]string
	hsPath       map[string][]string
	includesGrid bool
	ics          string
	ihs          string
	ocs          string
	ohs          string
	points       map[int]*PointResult
}

// Add points to transformation batch
func (t *TransformerOutput) Add(id int, pt *PointResult) {

	// Add point to batch
	t.points[id] = pt
}

// Trasnform batch
func (t *TransformerOutput) TransformBatch() (map[int]*PointResult, error) {

	// Get base cs
	from := t.ics

	// Iterate points
	for key, pt := range t.points {

		// Store intermediate results
		type IntRes struct {
			CS string
			X  float64
			Y  float64
		}

		// Track walk nodes
		fromNodes := []IntRes{
			{
				CS: t.ics,
				X:  pt.X,
				Y:  pt.Y,
			},
		}

		// Store results
		res := map[string][2]float64{}

		// Walk the graph
	graphLoop:
		for len(fromNodes) > 0 {

			nextNodes := []IntRes{}

			// Iterate nodes
			for _, node := range fromNodes {

				// If current node is of interest, store values
				if node.CS == t.ocs {
					res[t.ocs] = [2]float64{node.X, node.Y}
				}
				if t.includesGrid && node.CS == "bgs-cad" {
					res["bgs-cad"] = [2]float64{node.X, node.Y}
				}

				// Find connections
				connections, found := t.csPath[node.CS]

				// If no connections, continue
				if !found {
					continue
				}

				// Iterate connections
				for _, to := range connections {

					// Create next node
					nextNode := IntRes{
						CS: to,
						X:  node.X,
						Y:  node.Y,
					}

					// Get CS trasnformation parameters
					zones, ok := Repo.CSGraph.Get(from, to)
					if !ok {
						return nil, errors.ErrUnsupported
					}

					// Track if point is transformed
					transformed := false

					// Iterate over zones
					for _, zone := range zones {

						// Check if point is in zone
						if !zone.InZone(nextNode.X, nextNode.Y) {
							continue
						}

						// Helper values
						dx := nextNode.X - zone.X0
						dy := nextNode.Y - zone.Y0

						// Transform point
						nextNode.X = zone.A00 +
							zone.A10*dx +
							zone.A01*dy +
							zone.A20*dx*dx +
							zone.A11*dx*dy +
							zone.A02*dy*dy +
							zone.A30*dx*dx*dx +
							zone.A21*dx*dx*dy +
							zone.A12*dx*dy*dy +
							zone.A03*dy*dy*dy
						nextNode.Y = zone.B00 +
							zone.B10*dx +
							zone.B01*dy +
							zone.B20*dx*dx +
							zone.B11*dx*dy +
							zone.B02*dy*dy +
							zone.B30*dx*dx*dx +
							zone.B21*dx*dx*dy +
							zone.B12*dx*dy*dy +
							zone.B03*dy*dy*dy

						// Mark as tranformed
						transformed = true

						// Exit loop
						break
					}

					// Return error if not transformed
					if !transformed {
						pt.XYErr = "point out of transformation bounds"
						break graphLoop
					}

					// Add next node
					nextNodes = append(nextNodes, nextNode)
				}
			}

			// Update from nodes
			fromNodes = nextNodes
		}

		// If there is a transformation error
		if len(pt.XYErr) > 0 {

			// Update point
			t.points[key] = pt

			// Continue to next point
			continue
		}

		// Update point coordinates
		pt.X = res[t.ocs][0]
		pt.Y = res[t.ocs][1]
		if t.includesGrid {
			pt.Xbgs = res["bgs-cad"][0]
			pt.Ybgs = res["bgs-cad"][1]
		}
		t.points[key] = pt
	}

	// Get base hs
	from = t.ihs

	// Iterate over HS transformation path
	for _, to := range t.hsPath {

		// Get CS trasnformation parameters
		params, ok := Repo.HSGraph.Get(from, to[0])
		if !ok {
			return nil, errors.ErrUnsupported
		}

		// Update from
		from = to[0]

		// If grid type
		if params.Type == "grid" {

			// Get grid params
			gridParams, ok := Repo.HSGraph.methods.Grid[params.Name]
			if !ok {
				return nil, errors.ErrUnsupported
			}

			// Store grid vertices
			verticesList := []string{}

			// Iterate over points
			for _, pt := range t.points {

				// Skip if H is missing or if there is a XY err
				if !(pt.HasH && len(pt.XYErr) == 0) {
					continue
				}

				// Build vertex name
				name := makeOndulationVertexNameFromXY(gridParams, pt.Xbgs, pt.Ybgs)

				// Add to points
				verticesList = append(verticesList, name)
			}

			// Open DB
			// TODO: Make path configurable
			db, err := sql.Open("sqlite3", fmt.Sprintf("/grid-models/%s", gridParams.DB))
			if err != nil {
				return nil, err
			}

			// Close db on end
			defer db.Close()

			// Ping DB
			if err = db.Ping(); err != nil {
				return nil, err
			}

			const maxOpenDbConn = 10
			const maxIdleDbConn = 5
			const maxDbLifetime = 5 * time.Minute

			// Configure connection
			db.SetMaxOpenConns(maxOpenDbConn)
			db.SetMaxIdleConns(maxIdleDbConn)
			db.SetConnMaxLifetime(maxDbLifetime)

			// Define context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Build query
			query := fmt.Sprintf("SELECT * FROM undulation_points WHERE id IN (%s);", strings.Join(verticesList, ", "))

			// Get rows
			rows, err := db.QueryContext(ctx, query)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			// Store cached data
			vertices := make(map[string]float64)

			// Scan rows
			for rows.Next() {

				// Define base models
				c := struct {
					ID string
					H  float64
				}{}

				err = rows.Scan(
					&c.ID,
					&c.H,
				)
				if err != nil {
					return nil, err
				}

				// Add to accounts
				vertices[c.ID] = c.H
			}

			// Check for errors
			err = rows.Err()
			if err != nil {
				return nil, err
			}

			// Iterate over points
			for key, pt := range t.points {

				// Get height
				hr, err := gridInterpolation(gridParams, pt.Xbgs, pt.Ybgs, pt.H, params.Direction, vertices)
				if err != nil {
					return nil, err
				}

				// Update H
				pt.H = hr
				t.points[key] = pt
			}

			// Continue
			continue

			// If Plane
		} else if params.Type == "plane" {

			// Get grid params
			planeParams, ok := Repo.HSGraph.methods.Plane[params.Name]
			if !ok {
				return nil, errors.ErrUnsupported
			}

			// Iterate over points
			for key, pt := range t.points {

				// Get height
				hr, err := planeInterpolation(planeParams, pt.X, pt.Y, pt.H, params.Direction)
				if err != nil {
					return nil, err
				}

				// Update H
				pt.H = hr
				t.points[key] = pt
			}

			// Continue
			continue
		}

		// Unsuported method
		return nil, errors.ErrUnsupported
	}

	return t.points, nil
}
