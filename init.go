/*
MIT License

Copyright © 2020 Shivam Rathore

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
		RedirectURL:  "https://" + HostURL + "/",
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
