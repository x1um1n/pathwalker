package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	// blank import required for DB package
	_ "github.com/go-sql-driver/mysql"

	"github.com/google/uuid"
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
		stmt, es := DB.Prepare("UPDATE surveys SET `path_id` = '" + s.PathID + "',`survey_date` = '" + s.Date + "',`survey_submitted_by` = '" + s.User + "',`survey_detail` = '" + s.Detail + "',`image_ids` = '" + s.ImageIDs + "'")
		checkerr.Check(es, "Error preparing UPDATE")
		defer stmt.Close()

		_, err = stmt.Exec()
		checkerr.Check(err, "Error updating survey:", s.PathID, s.Date, s.User, s.Detail, s.ImageIDs)
	}

	return
}

// delSurvey deletes a survey record from the DB and purges any associated images
func delSurvey(sid string) (e error) {
	log.Println("Finding survey to delete: " + sid)
	_, e = getSurvey(sid)
	if !checkerr.Check(e, "Survey not found in database") {
		stmt, e := DB.Prepare("DELETE FROM surveys WHERE `survey_id` = ?")
		checkerr.Check(e, "Error preparing DELETE")
		defer stmt.Close()

		_, e = stmt.Exec(sid)
		if !checkerr.Check(e, "Error deleting survey: ", sid) {
			log.Printf("Survey %s successfully deleted\n", sid)
			//fixme: purge images
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
		switch e = row.Scan(&i.ImageID, &i.Filename, &i.Location.Latitude, &i.Location.Longitude); e {
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

// putImage inserts an image record corresponding to an uploaded file
func putImage(im Image) (id string, e error) {
	log.Println("Inserting new image for path " + im.S3Path)
	stmt, e := DB.Prepare("INSERT INTO images (`image_id`,`filename`,`s3-location`,`image_latitude`,`image_longitude`) VALUES (?,?,?,?,?)")
	checkerr.Check(e, "Error preparing INSERT")
	defer stmt.Close()

	id = uuid.New().String()

	_, e = stmt.Exec(id, im.Filename, im.S3Path, im.Location.Latitude, im.Location.Longitude)
	if checkerr.Check(e, "Error inserting image record:", id, im.Filename, im.S3Path, strconv.FormatFloat(im.Location.Latitude, 'E', -1, 64), strconv.FormatFloat(im.Location.Longitude, 'E', -1, 64)) {
		id = "" //blank out the id if the insert fails
	}

	return
}
