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
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

func Welcome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")

	state := r.FormValue("state")
	code := r.FormValue("code")
	data := ViewData{}
	var token *oauth2.Token
	if cookie, _ := r.Cookie("token"); cookie != nil {
		token, _ = ParseToken(cookie.Value)
	}

	// user is not logged in
	if r.Method == http.MethodGet && ((state == "" && code == "") || token == nil) {
		render(w, data)
		return
	}

	ctx := r.Context()

	// user just got redirected from logging in
	if token == nil {
		var err error
		token, err = GetOauthConfig(ctx, state, code)
		if err != nil {
			render(w, ViewData{})
			log.Println("Oauth error:", err)
			return
		}
	}
	val, _ := json.Marshal(token)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    string(val),
		Domain:   HostURL,
		HttpOnly: true,
		Secure:   HostURL != "localhost:"+PORT,
	})

	client := config.Client(ctx, token)
	per, err := GetUserInfo(ctx, client)
	if err != nil {
		log.Println("Could not get user info:", err)
		render(w, ViewData{})
		return
	}

	data.Name = per.Name
	data.Authenticated = true
	render(w, data)
}

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// oauth2 state string randomisation
	oauthStateString = RandomString(10)

	url := config.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Process(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")
	var token *oauth2.Token
	if cookie, _ := r.Cookie("token"); cookie != nil {
		token, _ = ParseToken(cookie.Value)
	}
	if token == nil {
		data := ViewData{
			Authenticated: true,
			Errors: []string{
				fmt.Sprintf("Client token expired redirecting to login page"),
			},
		}
		render(w, data)
		return
	}
	p, err := FilPopulateObject(r.FormValue("docID"), r.FormValue("sheetID"), r.FormValue("ent"), r.FormValue("cols"))
	if err != nil {
		data := ViewData{
			Authenticated: true,
			Errors: []string{
				fmt.Sprintf("Sorry, %v", err),
			},
		}
		render(w, data)
		return
	}

	client := config.Client(r.Context(), token)
	list, err := p.Process(client)
	if err != nil {
		data := ViewData{
			Authenticated: true,
			Errors: []string{
				fmt.Sprintf("Sorry, failed to populate: %v", err),
			},
		}
		if err.Error() == "client expired" {
			data.Authenticated = false
		}
		render(w, data)
		return
	}

	docCreated := int(p.Entries) - len(list)
	suc := []string{
		fmt.Sprintf("Successfully created %v documents", docCreated),
	}
	if docCreated == 0 {
		suc = []string{}
	}
	errs := make([]string, 0, len(list))
	for _, res := range list {
		errs = append(errs, fmt.Sprintf("Sorry, failed to populate Document-%v: %v", res.DocNo, res.ErrorMessage))
	}
	data := ViewData{
		Authenticated: true,
		Success:       suc,
		Errors:        errs,
	}
	render(w, data)
}

func googleVerification(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if _, err := fmt.Fprintf(w, `google-site-verification: googleccd40724f8c32619.html`); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("error on welcome page: %v", err)
	}
}

func render(w http.ResponseWriter, data ViewData) {
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("error on welcome page: %v", err)
	}
}
