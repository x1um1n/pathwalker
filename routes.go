package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/x1um1n/checkerr"
)

// AddRoutes adds api routes and assigns handlers for them
func AddRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/add", AddSurvey)
	// router.Post("/update/{surveyID}", UpdateSurvey)
	// router.Post("/delete/{surveyID}", DeleteSurvey)
	// router.Get("/list/{pathID}", ListSurveysForPath)
	// router.Get("/list/survey/{surveyID}", ListSurvey)
	// router.Get("/list/images/{pathID}", ListImages)
	return router
}

// AddSurvey is a handler function which adds a survey
func AddSurvey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body) //parse json in request
	var s Survey
	err := decoder.Decode(&s)
	checkerr.Check(err, "Error decoding json")

	log.Println("Inserting new survey for path " + s.PathID)
	stmt,es := DB.Prepare("INSERT INTO surveys (`survey_id`,`path_id`,`survey_date`,`survey_submitted_by`,`detail`,`image_ids`) VALUES (UUID(),?,?,?,?,?)")
	checkerr.Check500(es, w, "Error preparing INSERT")
	defer stmt.Close()

  // create a CSL of imageIDs to be attached to the report
  var imageIDs string
  for i, img := range s.Images {
    if i == 0 {
      imageIDs = img.ImageID+","
    } else if i == len(s.Images) {
      imageIDs += img.ImageID
    } else {
      imageIDs += img.ImageID+","
    }
  }

	_, err = stmt.Exec(s.PathID, s.Date, s.User, s.Detail, imageIDs)
	checkerr.Check500(err, w, "Error inserting action:", s.PathID, s.Date, s.User, s.Detail, imageIDs)

	response := make(map[string]string)
	if err != nil {
		response["message"] = "Error adding action - " + err.Error()
	} else {
		response["message"] = "Successfully added action"
	}
	render.JSON(w, r, response)
}
