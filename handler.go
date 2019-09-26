package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func Welcome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")

	state := r.FormValue("state")
	code := r.FormValue("code")
	data := ViewData{}

	// user is not logged in
	if r.Method == http.MethodGet && state == "" && code == "" {
		render(w, data)
		return
	}

	ctx := r.Context()

	// user just got redirected from logging in
	if client == nil {
		err := GetOauthConfig(ctx, state, code)
		if err != nil {
			render(w, ViewData{})
			log.Printf("oauth error: %v", err)
			return
		}

		per, err := GetUserInfo(ctx, client)
		if err != nil {
			log.Printf("could not get user info: %v", err)
			render(w, ViewData{})
			return
		}

		data.Name = per.Name
		data.Authenticated = true
		render(w, data)
		return
	}

	// user just completed populating a template
	per, err := GetUserInfo(ctx, client)
	if err != nil {
		log.Printf("could not get user info: %v", err)
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
	p, err := FilPopulateObject(r.FormValue("docID"), r.FormValue("sheetID"), r.FormValue("ent"), r.FormValue("cols"))
	if err != nil {
		data := ViewData{
			Authenticated: true,
			Errors: []string{
				fmt.Sprintf("failed to parse form: %v", err),
			},
		}
		render(w, data)
		return
	}
	if err := p.Process(); err != nil {
		data := ViewData{
			Authenticated: true,
			Errors: []string{
				fmt.Sprintf("failed to populate: %v", err),
			},
		}
		render(w, data)
		return
	}

	data := ViewData{
		Authenticated: true,
		Success:       []string{"Succesfully created N documents"},
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
