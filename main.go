package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
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
