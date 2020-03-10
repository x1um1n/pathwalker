package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/x1um1n/checkerr"
)

// AddRoutes adds api routes and assigns handlers for them
func AddRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/upload", UploadImage)
	router.Post("/add", AddSurvey)
	router.Post("/update/{survey-id}", UpdateSurvey)
	// router.Post("/delete/{survey-id}", DeleteSurvey)
	router.Get("/list/{path-id}", ListSurveysForPath)
	router.Get("/fetch/survey/{survey-id}", ListSurvey)
	// router.Get("/list/images/{path-id}", ListImages)
	return router
}

// UploadImage is the handler that writes uploaded files to the temp-images dir
func UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Println("File upload endpoint hit")

	err := r.ParseMultipartForm(200000)
	checkerr.Check(err, "Error parsing form data")

	formdata := r.MultipartForm

	files := formdata.File["images"] //field name for upload form, should be a multiple file input

	for _, f := range files {
		file, err := f.Open()
		defer file.Close()
		checkerr.Check500(err, w, "Error reading : "+f.Filename)

		out, err := os.Create("temp-images/" + f.Filename)
		defer out.Close()
		checkerr.Check500(err, w, "Unable to create the file for writing : "+f.Filename)

		_, err = io.Copy(out, file)
		checkerr.Check500(err, w, "Error reading : "+f.Filename)

		fmt.Fprintf(w, "Files uploaded successfully : ")
		fmt.Fprintf(w, f.Filename+"\n")
	}
}

// AddSurvey is a handler function which adds a survey
func AddSurvey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body) //parse json in request
	var s Survey
	err := decoder.Decode(&s)
	checkerr.Check(err, "Error decoding json")

	log.Println("Inserting new survey for path " + s.PathID)
	stmt, es := DB.Prepare("INSERT INTO surveys (`survey_id`,`path_id`,`survey_date`,`survey_submitted_by`,`survey_detail`,`image_ids`) VALUES (UUID(),?,?,?,?,?)")
	checkerr.Check500(es, w, "Error preparing INSERT")
	defer stmt.Close()

	// create a CSL of imageIDs to be attached to the report
	var imageIDs string
	for i, img := range s.Images {
		if i == 0 {
			imageIDs = img.ImageID + ","
		} else if i == len(s.Images) {
			imageIDs += img.ImageID
		} else {
			imageIDs += img.ImageID + ","
		}
	}

	_, err = stmt.Exec(s.PathID, s.Date, s.User, s.Detail, imageIDs)
	checkerr.Check500(err, w, "Error inserting survey:", s.PathID, s.Date, s.User, s.Detail, imageIDs)

	response := make(map[string]string)
	if err != nil {
		response["message"] = "Error adding survey - " + err.Error()
	} else {
		response["message"] = "Successfully added survey"
	}
	render.JSON(w, r, response)
}

// UpdateSurvey updates a survey record with data sent in POST
func UpdateSurvey(w http.ResponseWriter, r *http.Request) {

}

// ListSurveysForPath returns all the surveys completed for a given path
func ListSurveysForPath(w http.ResponseWriter, r *http.Request) {
	pathID := chi.URLParam(r, "path-id")
	log.Println("Getting all surveys completed for path " + pathID)

	qry, err := DB.Prepare("SELECT survey_id, survey_date, survey_submitted_by FROM surveys WHERE path_id = '" + pathID + "'")
	checkerr.Check500(err, w, "Error preparing SELECT")
	defer qry.Close()

	rows, err := qry.Query()
	checkerr.Check500(err, w, "Error executing query")
	defer rows.Close()

	var surveys []Survey
	for rows.Next() {
		var s Survey
		err = rows.Scan(&s.SurveyID, &s.Date, &s.User)
		checkerr.Check500(err, w, "Error reading results from query")
		s.PathID = pathID
		surveys = append(surveys, s)
	}

	render.JSON(w, r, surveys)
}

// ListSurvey returns the entire survey record requested
func ListSurvey(w http.ResponseWriter, r *http.Request) {
	surveyID := chi.URLParam(r, "survey-id")
	log.Println("Getting survey " + surveyID)

	qry, err := DB.Prepare("SELECT * FROM surveys WHERE survey_id = '" + surveyID + "'")
	checkerr.Check500(err, w, "Error preparing SELECT from surveys")
	defer qry.Close()

	var s Survey
	var imageIDs string
	row := qry.QueryRow()
	switch err = row.Scan(&s.SurveyID, &s.PathID, &s.Date, &s.User, &s.Detail, &imageIDs); err {
	case sql.ErrNoRows:
		checkerr.Check500(err, w, "No survey found for "+surveyID)
	case nil:
		img := strings.Split(imageIDs, ",")
		for _, im := range img {
			qry, err = DB.Prepare("SELECT * FROM images WHERE image_id = '" + im + "'")
			checkerr.Check500(err, w, "Error preparing SELECT from images")
			defer qry.Close()

			var i Image
			row = qry.QueryRow()
			switch err = row.Scan(&i.ImageID, &i.PathID, &i.Filename, &i.Desc, &i.Location.Latitude, &i.Location.Longitude); err {
			case sql.ErrNoRows:
				log.Println("No images found for imageID " + im)
			case nil:
				s.Images = append(s.Images, i)
			default:
				checkerr.Check500(err, w, "Error reading image data for "+im)
			}
		}
		render.JSON(w, r, s)
	default:
		checkerr.Check500(err, w, "Error reading survey data for "+surveyID)
	}
}
