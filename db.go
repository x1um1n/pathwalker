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
		var imageIDs string
		row := qry.QueryRow()
		switch e = row.Scan(&s.SurveyID, &s.PathID, &s.Date, &s.User, &s.Detail, &imageIDs); e {
		case sql.ErrNoRows:
			checkerr.Check(e, "No survey found for "+sid)
		case nil:
			img := strings.Split(imageIDs, ",")
			for _, im := range img {
				qry, e = DB.Prepare("SELECT * FROM images WHERE image_id = '" + im + "'")
				checkerr.Check(e, "Error preparing SELECT from images")
				defer qry.Close()

				var i Image
				row = qry.QueryRow()
				switch e = row.Scan(&i.ImageID, &i.PathID, &i.Filename, &i.Desc, &i.Location.Latitude, &i.Location.Longitude); e {
				case sql.ErrNoRows:
					log.Println("No images found for imageID " + im)
				case nil:
					s.Images = append(s.Images, i)
				default:
					checkerr.Check(e, "Error reading image data for "+im)
				}
			}
		default:
			checkerr.Check(e, "Error reading survey data for "+sid)
		}
	}

	return
}

// putSurvey writes a survey struct to the DB
// if a survey record already exists with the supplied ID, it will be overwritten
func putSurvey(s Survey) (e error) {
	//fixme
	return
}
