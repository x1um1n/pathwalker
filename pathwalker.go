// pathwalker is a restful webservice that allows users to submit reports and
// images of the state of public footpaths.  This should be integrated into the
// website of the council responsible for the paths, so no frontend is included
// in the project.
package main

import (
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/heptiolabs/healthcheck"
	"github.com/x1um1n/checkerr"
)

// Routes creates a router and calls routes.Routes to add in the api routes
func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	)

	router.Route("/v1", func(r chi.Router) {
		r.Mount("/api", AddRoutes())
	})

	return router
}

// defines and starts the healthcheck
func startHealth() {
	h := healthcheck.NewHandler()

	log.Println("Adding database check")
	h.AddReadinessCheck("database", healthcheck.DatabasePingCheck(DB, 1*time.Second))
	h.AddLivenessCheck("database", healthcheck.DatabasePingCheck(DB, 1*time.Second))
	go http.ListenAndServe("0.0.0.0:9080", h)
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

func main() {
	LoadKoanf() //read in the config
	InitDB()    //initialise the database
	defer DB.Close()
	sess = connectAWS() //initialise the connection to S3

	log.Println("Creating REST api routes")
	router := Routes()

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) //walk and print out all routes
		return nil
	}
	err := chi.Walk(router, walkFunc)
	checkerr.CheckFatal(err, "Error walking api routes") //panic if there's an error

	go startHealth()

	log.Fatal(http.ListenAndServe(":8080", router))
}
