package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"text/template"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	oauth22 "google.golang.org/api/oauth2/v2"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

var (
	HostURL, PORT string

	seed             *rand.Rand
	client           *http.Client
	config           *oauth2.Config
	oauthStateString string
	tmpl             *template.Template
)

func init() {
	var err error
	tmpl, err = template.ParseFiles("pages/base.gohtml", "pages/form.gohtml")
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	log.SetFlags(0)
	// Randomised seed
	seed = rand.New(rand.NewSource(time.Now().UnixNano()))

	// oauth2 state string randomisation
	oauthStateString = RandomString(10)

	// Environment Variables values
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "80"
	}
	HostURL = os.Getenv("POPULATE_HOST_URL")
	if HostURL == "" {
		HostURL = "localhost:" + PORT
	}
	ClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if ClientID == "" {
		log.Println("[Warning] ClientID is not provided. Google Oauth flow will not work, use the env variable GOOGLE_CLIENT_ID to set it")
	}
	ClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if ClientSecret == "" {
		log.Println("[Warning] ClientSecret is not provided. Google Oauth flow will not work, use the env variable GOOGLE_CLIENT_SECRET to set it")
	}

	config = &oauth2.Config{
		RedirectURL:  "http://" + HostURL + "/",
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Scopes: []string{
			docs.DriveScope,
			oauth22.UserinfoEmailScope,
			oauth22.UserinfoProfileScope,
		},
		Endpoint: google.Endpoint,
	}
}
