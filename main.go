package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dimitargrozev5/bgstrans-2-api/config"
	"github.com/dimitargrozev5/bgstrans-2-api/transformations"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gopkg.in/yaml.v3"
)

// App config
var app config.App

// Main func
func main() {

	// Setup app
	setup()

	// Create router
	mux := chi.NewRouter()

	// Setup recoverer
	mux.Use(middleware.Recoverer)

	// TODO: handle CSRF protection and same site origin protection
	// mux.Use(NoSurf)
	// mux.Use(SessionLoad)

	// Setup main transformation route
	mux.Post("/transform", func(w http.ResponseWriter, r *http.Request) {

		// Close response body
		defer r.Body.Close()

		// Check Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		var data TransfomrationRequest
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Failed to parse JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get CS names
		inputCS := fmt.Sprintf("%s-%s", data.InputCS, data.InputCSVariant)
		outputCS := fmt.Sprintf("%s-%s", data.OutputCS, data.OutputCSVariant)

		// Get transformer
		transformer, err := transformations.GetTransformer(inputCS, outputCS, data.InputHS, data.OutputHS)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Store output
		results := map[int]*transformations.PointResult{}

		// Iterate over data
		for i, line := range data.Data {

			fmt.Println(line)

			// Store output for current line
			var o transformations.PointResult
			results[i] = &o

			// Add empty line
			if len(line) == 0 {
				continue
			}

			// Store comment/point name
			if len(line) == 1 || len(line) > 2 {
				o.Name = line[0]

				// Exit if only comment
				if len(line) == 1 {
					continue
				}
			}

			// Continue if len is less than 3
			if len(line) < 2 {
				continue
			}

			// Get X index
			xIndex := 0

			// Update index if there is a point name
			if len(line) > 2 {
				xIndex = 1
			}

			// Get X
			o.X, err = strconv.ParseFloat(line[xIndex], 64)
			if err != nil {
				// TODO: add better error
				o.XYErr = fmt.Sprintf("Error parsing '%s' as number", line[xIndex])
				continue
			}

			// Parse Y
			o.Y, err = strconv.ParseFloat(line[xIndex+1], 64)
			if err != nil {
				// TODO: add better error
				o.XYErr = fmt.Sprintf("Error parsing '%s' as number", line[xIndex+1])
				continue
			}

			// If there is an H
			if len(line) > 3 {

				// Parse H
				o.H, err = strconv.ParseFloat(line[3], 64)
				if err != nil {
					// TODO: add better error
					o.HErr = fmt.Sprintf("Error parsing '%s' as number", line[3])
					continue
				}
				o.HasH = true
			}

			// Get other fields
			if len(line) > 4 {
				o.Var = line[4:]
			}

			// Add point for tranformation
			transformer.Add(i, &o)
		}

		// Transform data
		transResults, err := transformer.TransformBatch()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Store api result
		var apiResult [][]string

		// Iterate over points
		for i := range data.Data {

			// Try to get point from results
			pt, ok := results[i]

			// If not found, get from trans results
			if !ok {
				pt = transResults[i]
			}

			// Create output row
			var row []string

			// Add name to row
			if len(pt.Name) > 0 {
				row = append(row, pt.Name)
			}

			// Add coordinates or error
			if len(pt.XYErr) > 0 {
				row = append(row, pt.XYErr)
			} else {
				row = append(row, fmt.Sprintf("%.3f", pt.X))
				row = append(row, fmt.Sprintf("%.3f", pt.Y))
			}

			// Add height or error
			if len(pt.HErr) > 0 {
				row = append(row, pt.HErr)
			} else {
				row = append(row, fmt.Sprintf("%.3f", pt.H))
			}

			// Add other fields
			row = append(row, pt.Var...)

			// Add row to api output
			apiResult = append(apiResult, row)
		}

		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")

		// Set the status code
		w.WriteHeader(http.StatusOK)

		// Write to response
		json.NewEncoder(w).Encode(TransformationResponse{Data: apiResult})
	})

	// Starting server
	fmt.Println("Starting server on port :3000")

	// Start server
	http.ListenAndServe(":3000", mux)
}

// Config type
type Config struct {
	InProduction bool `yaml:"inProduction"`
}

// Setup function
func setup() {

	// Open the YAML file
	file, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("Error opening config file: %v\n", err)
		return
	}
	defer file.Close()

	// Load yaml config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&app); err != nil {
		log.Fatalf("Error decoding config file: %v\n", err)
		return
	}

	// Setup tranformations
	transformations.Setup(&app)
}

// Tranformation request format
type TransfomrationRequest struct {
	// Input systems
	InputCS        string `json:"ics"`
	InputCSVariant string `json:"icsv"`
	InputHS        string `json:"ihs"`

	// Output systems
	OutputCS        string `json:"ocs"`
	OutputCSVariant string `json:"ocsv"`
	OutputHS        string `json:"ohs"`

	// Raw data row, that can contain multiple string fields
	// 0: No data
	// 1: Comment
	// 2: X, Y
	// 3: N, X, Y
	// >= 4: N, X, Y, H, (Various string fields)
	Data [][]string `json:"d"`
}

// Transformation response format
type TransformationResponse struct {
	Data [][]string `json:"d"`
}
