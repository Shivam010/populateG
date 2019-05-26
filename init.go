package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	oauth22 "google.golang.org/api/oauth2/v2"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

var (
	PORT             string
	seed             *rand.Rand
	client           *http.Client
	config           *oauth2.Config
	oauthStateString string
)

func init() {
	// Randomised seed
	seed = rand.New(rand.NewSource(time.Now().UnixNano()))

	// oauth2 state string randomisation
	oauthStateString = RandomString(10)

	// Environment Variables values
	HostURL := os.Getenv("POPULATE_HOST_URL")
	ClientID := os.Getenv("GOOGLE_CLIENT_ID")
	ClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	PORT = os.Getenv("PORT")

	config = &oauth2.Config{
		RedirectURL:  "https://" + HostURL + "/home",
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
