package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	// blank import required for DB package
	_ "github.com/go-sql-driver/mysql"

	"github.com/x1um1n/checkerr"
)

// DB is the global DB handler
var DB *sql.DB

// InitDB creates the DB handler and a test connection to confirm it works
func InitDB() {
	log.Printf("Connecting to the %s database on %s", K.String("dbschema"), K.String("dbhost"))
	connstr := K.String("dbuser") + ":" + K.String("dbpasswd") + "@tcp(" + K.String("dbhost") + ":" + K.String("dbport") + ")/" + K.String("dbschema")
	var err error
	DB, err = sql.Open("mysql", connstr)
	checkerr.CheckFatal(err, "Error creating database object")

	for i := 0; i < K.Int("dbretries"); i++ {
		err = DB.Ping()

		if err == nil {
			log.Println("Database connection established")
			break
		} //get out of the loop
		if i < K.Int("dbretries") {
			checkerr.Check(err, "Error connecting to database... retrying in 5s")
			time.Sleep(5 * time.Second)
		} else {
			checkerr.CheckFatal(err, "Error connecting to database :(")
		}
	}
}

// getSurvey retrieves a specified survey from the db and returns
func getSurvey(sid string) (s Survey, e error) {
	log.Println("Getting survey " + sid)

	qry, e := DB.Prepare("SELECT * FROM surveys WHERE survey_id = '" + sid + "'")
	defer qry.Close()
	if !checkerr.Check(e, "Error preparing SELECT from surveys") {
		row := qry.QueryRow()
		switch e = row.Scan(&s.SurveyID, &s.PathID, &s.Date, &s.User, &s.Detail, &s.ImageIDs); e {
		case sql.ErrNoRows:
			checkerr.Check(e, "No survey found for "+sid)
		case nil:
			s.Images = getImages(s.ImageIDs)
		default:
			checkerr.Check(e, "Error reading survey data for "+sid)
		}
	}

	return
}

// getImages takes a CSL of imageIDs and returns a slice containing details of the relevant images
func getImages(imageIDs string) (ims []Image) {
	img := strings.Split(imageIDs, ",")
	for _, im := range img {
		qry, e := DB.Prepare("SELECT * FROM images WHERE image_id = '" + im + "'")
		checkerr.Check(e, "Error preparing SELECT from images")
		defer qry.Close()

		var i Image
		row := qry.QueryRow()
		switch e = row.Scan(&i.ImageID, &i.PathID, &i.Filename, &i.Desc, &i.Location.Latitude, &i.Location.Longitude); e {
		case sql.ErrNoRows:
			log.Println("No images found for imageID " + im)
		case nil:
			ims = append(ims, i)
		default:
			checkerr.Check(e, "Error reading image data for "+im)
		}
	}
	return
}

// makeImageCSL extracts the IDs from a slice of Images and returns a CSL
func makeImageCSL(ims []Image) (s string) {
	for i, img := range ims {
		if i == 0 {
			s = img.ImageID + ","
		} else if i == len(ims) {
			s += img.ImageID
		} else {
			s += img.ImageID + ","
		}
	}
	return
}

// putSurvey writes a survey struct to the DB
// if a survey record already exists with the supplied ID, it will be overwritten
func putSurvey(s Survey) (e error) {
	_, err := getSurvey(s.SurveyID)
	if err == sql.ErrNoRows {
		log.Println("Inserting new survey for path " + s.PathID)
		stmt, es := DB.Prepare("INSERT INTO surveys (`survey_id`,`path_id`,`survey_date`,`survey_submitted_by`,`survey_detail`,`image_ids`) VALUES (UUID(),?,?,?,?,?)")
		checkerr.Check(es, "Error preparing INSERT")
		defer stmt.Close()

		s.ImageIDs = makeImageCSL(s.Images)

		_, err = stmt.Exec(s.PathID, s.Date, s.User, s.Detail, s.ImageIDs)
		checkerr.Check(err, "Error inserting survey:", s.PathID, s.Date, s.User, s.Detail, s.ImageIDs)
	} else {
		log.Println("Updating survey " + s.SurveyID)
		if s.ImageIDs == "" {
			s.ImageIDs = makeImageCSL(s.Images)
		}
		stmt, es := DB.Prepare("UPDATE surveys SET `path_id` = " + s.PathID + ",`survey_date` = " + s.Date + ",`survey_submitted_by` = " + s.User + ",`survey_detail` = " + s.Detail + ",`image_ids` = " + s.ImageIDs)
		checkerr.Check(es, "Error preparing UPDATE")
		defer stmt.Close()

		_, err = stmt.Exec()
		checkerr.Check(err, "Error updating survey:", s.PathID, s.Date, s.User, s.Detail, s.ImageIDs)
	}

	return
}
