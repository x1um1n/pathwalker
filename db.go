package main

import (
	"database/sql"
	"log"
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
