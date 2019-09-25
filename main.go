package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {

	// Initialize a router as usual
	router := httprouter.New()
	// Welcome screen
	router.GET("/", Welcome)
	// OAuth check
	router.GET("/login", Login)
	router.GET("/home", CallbackHome)
	router.POST("/process", Process)
	router.GET("/googleccd40724f8c32619.html", googleVerification)

	// temporary router for the static pages
	router.ServeFiles("/page/*filepath", http.Dir("./pages"))

	fmt.Printf("Running server at http://%s \n", HostURL)
	log.Fatal(http.ListenAndServe(":"+PORT, router))
}
