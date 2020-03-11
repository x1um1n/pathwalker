// api handler functions
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/x1um1n/checkerr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AddRoutes adds api routes and assigns handlers for them
func AddRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/upload", UploadImage)
	router.Post("/add", AddSurvey)
	router.Post("/update/{survey-id}", UpdateSurvey)
	router.Post("/delete/{survey-id}", DeleteSurvey)
	router.Get("/list/{path-id}", ListSurveysForPath)
	router.Get("/fetch/survey/{survey-id}", FetchSurvey)
	return router
}

// UploadImage is the handler that writes uploaded files to the temp-images dir
func UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Println("File upload endpoint hit")

	err := r.ParseMultipartForm(200000)
	checkerr.Check(err, "Error parsing form data")
	formdata := r.MultipartForm

	files := formdata.File["images"] //field name for upload form, should be a multiple file input

	var imageIDs []string
	for i := range files {
		file, err := files[i].Open()
		defer file.Close()

		var img Image
		img.Filename = files[i].Filename

		if !checkerr.Check500(err, w, "Error reading : "+img.Filename) {
			log.Println("Creating temp file")
			out, err := os.Create("temp-images/" + img.Filename)
			defer out.Close()

			if !checkerr.Check500(err, w, "Unable to create the file for writing : "+img.Filename) {
				log.Println("Writing data to temp file")
				_, err = io.Copy(out, file)

				if !checkerr.Check500(err, w, "Error reading : "+img.Filename) {
					fmt.Fprintf(w, "Temp file written successfully : ")
					fmt.Fprintf(w, img.Filename+"\n")

					// extract EXIF data
					// fixme: shouldn't really need to write the file out to do this
					img.Location, err = getLocation("temp-images/" + img.Filename)
					checkerr.Check(err, "Error extracting location data from image file")

					file.Seek(0, io.SeekStart) // go back to the start of the file

					// Upload the file to S3
					uploader := s3manager.NewUploader(sess)
					u, err := uploader.Upload(&s3manager.UploadInput{
						Bucket: aws.String(K.String("images_bucket")),
						Key:    aws.String(img.Filename),
						Body:   file,
					})
					if !checkerr.Check500(err, w, "Failed to upload file to S3") {
						fmt.Fprintf(w, "Successfully uploaded %s to %s\n", files[i].Filename, K.String("images_bucket"))
						img.S3Path = u.Location

						id, err := putImage(img)
						if !checkerr.Check500(err, w, "Error storing image data in database") {
							imageIDs = append(imageIDs, id)
						}
					}
				}
			}
			err = os.Remove("temp-images/" + img.Filename) //tidy up temp file
			checkerr.Check(err, "Error deleting temp file")
		}
	}

	if len(imageIDs) != 0 {
		// build a CSL of all the imageIDs
		var IDs string
		for i, im := range imageIDs {
			switch i {
			case 0:
				IDs = im
			default:
				IDs = IDs + "," + im
			}
		}

		response := make(map[string]string)
		response["message"] = "Successfully added images"
		response["image-ids"] = IDs
		render.JSON(w, r, response)
	} else {
		response := make(map[string]string)
		response["message"] = "Failed to add any images"
		render.JSON(w, r, response)
	}
}

// AddSurvey is a handler function which adds a survey
func AddSurvey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body) //parse json in request
	var s Survey
	err := decoder.Decode(&s)
	checkerr.Check500(err, w, "Error decoding json")

	err = putSurvey(s)
	imageIDs := makeImageCSL(s.Images)
	response := make(map[string]string)
	if checkerr.Check(err, "Error inserting survey: ", s.PathID, s.Date, s.User, s.Detail, imageIDs) {
		response["message"] = "Error adding survey - " + err.Error()
	} else {
		response["message"] = "Successfully added survey "
	}
	render.JSON(w, r, response)
}

// UpdateSurvey updates a survey record with data sent in POST
// blank fields are ignored, non-blank fields overwrite DB record
func UpdateSurvey(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "survey-id")
	s, err := getSurvey(sid)
	checkerr.Check500(err, w, "Error retrieving survey "+sid)

	decoder := json.NewDecoder(r.Body) //parse json in request
	var s2 Survey
	err = decoder.Decode(&s2)
	checkerr.Check500(err, w, "Error decoding json")

	if s2.PathID != "" {
		s.PathID = s2.PathID
	}
	if s2.Date != "" {
		s.Date = s2.Date
	}
	if s2.User != "" {
		s.User = s2.User
	}
	if s2.Detail != "" {
		s.Detail = s2.Detail
	}

	s.ImageIDs = makeImageCSL(s.Images)

	err = putSurvey(s)
	checkerr.Check500(err, w, "Failed to write updated record to the database")
	render.JSON(w, r, s)
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

// FetchSurvey returns the entire survey record requested
func FetchSurvey(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "survey-id")
	s, err := getSurvey(sid)
	checkerr.Check500(err, w, "Error retrieving survey "+sid)
	render.JSON(w, r, s)
}

// DeleteSurvey deletes the survey from the database
func DeleteSurvey(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "survey-id")
	err := delSurvey(sid)

	response := make(map[string]string)

	if checkerr.Check500(err, w, "Error deleting survey "+sid) {
		response["message"] = "Error deleting survey - " + err.Error()
	} else {
		response["message"] = "Successfully deleted survey " + sid
	}
	render.JSON(w, r, response)
}
