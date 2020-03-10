package main

// Path is used to store information about the footpaths
type Path struct {
	PathID  string   `json:"path-id"`
	Desc    string   `json:"path-description"`
	Surveys []Survey `json:"surveys"`
}

// Image is used to store details of uploaded images
type Image struct {
	ImageID  string     `json:"image-id"`
	PathID   string     `json:"path-id"`
	Filename string     `json:"filename"`
	Desc     string     `json:"image-description"`
	Location Coordinate `json:"image-coordinates"`
}

// Coordinate is used to store lat/long extracted from EXIF data
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Survey is used to store survey data
type Survey struct {
	SurveyID string  `json:"survey-id"`
	PathID   string  `json:"path-id"`
	Date     string  `json:"survey-date"`
	User     string  `json:"survey-submitted-by"`
	Detail   string  `json:"detail"`
	Images   []Image `json:"images"`
	ImageIDs string  `json:"image-ids"`
}
