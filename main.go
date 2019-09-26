package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	var err error
	tmpl, err = tmpl.ParseFiles("pages/base.gohtml", "pages/home.gohtml")
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	// Initialize a router as usual
	router := httprouter.New()
	// Welcome screen
	router.GET("/", Welcome)
	router.POST("/", Welcome)
	// OAuth check
	router.GET("/login", Login)
	router.POST("/process", Process)
	router.GET("/googleccd40724f8c32619.html", googleVerification)

	fmt.Printf("Running server at http://%s \n", HostURL)
	log.Fatal(http.ListenAndServe(":"+PORT, router))
}
