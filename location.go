// functions for reading EXIF data from images to capture the location photos were taken
package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/gosexy/exif"
	"github.com/x1um1n/checkerr"
)

// getLocation opens a local file, extracts the lat/long from EXIF data & returns it as a coordinate struct
func getLocation(path string) (c Coordinate, e error) {
	log.Println("Reading EXIF data from " + path)
	parser := exif.New()
	e = parser.Open(path)
	if !checkerr.Check(e, "Error parsing EXIF data from ", path) {
		if parser.Tags["Latitude"] != "" {
			if parser.Tags["North or South Latitude"] != "" {
				if parser.Tags["Longitude"] != "" {
					if parser.Tags["East or West Longitude"] != "" {
						c = parseCoord(parser.Tags["Latitude"], parser.Tags["North or South Latitude"], parser.Tags["Longitude"], parser.Tags["East or West Longitude"])
					} else {
						log.Println("Location either missing or incomplete in EXIF data")
					}
				} else {
					log.Println("Location either missing or incomplete in EXIF data")
				}
			} else {
				log.Println("Location either missing or incomplete in EXIF data")
			}
		} else {
			log.Println("Location either missing or incomplete in EXIF data")
		}
	}
	return
}

// parseCoordString takes an EXIF lat/long string and converts it to a coordinate
func parseCoordString(val string) float64 {
	chunks := strings.Split(val, ",")
	hours, _ := strconv.ParseFloat(strings.TrimSpace(chunks[0]), 64)
	minutes, _ := strconv.ParseFloat(strings.TrimSpace(chunks[1]), 64)
	seconds, _ := strconv.ParseFloat(strings.TrimSpace(chunks[2]), 64)

	return hours + (minutes / 60) + (seconds / 3600)
}

// parseCoord takes location values from EXIF data and returns a Coordinate struct
func parseCoord(latVal, latRef, lngVal, lngRef string) (c Coordinate) {
	lat := parseCoordString(latVal)
	lng := parseCoordString(lngVal)

	if latRef == "S" { // N is "+", S is "-"
		lat *= -1
	}

	if lngRef == "W" { // E is "+", W is "-"
		lng *= -1
	}

	c = Coordinate{lat, lng}

	return
}
