package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func Welcome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := fmt.Fprintf(w, `<html>
<body>
Welcome to PopulateG | A Plugin To Populate Google Docs Template <br> <a href="/login">Log In</a>
</body>
<br>
For more info, visit: <a href="https://github.com/Shivam010/populateg">github.com/Shivam010</a>
</html>`)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("error on welcome page: %v", err)
	}
}

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// oauth2 state string randomisation
	oauthStateString = RandomString(10)

	url := config.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	state := r.FormValue("state")
	code := r.FormValue("code")
	if err := GetOauthConfig(ctx, state, code); err != nil {
		log.Printf("oauth error: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	per, err := GetUserInfo(ctx, client)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if _, err := fmt.Fprintf(w, `<html>
<body>
Hi %v, welcome to PopulateG | A Plugin To Populate Google Docs Template
<br>
<form action="/process" method="post">
	Google Docs Template Url: <br>
		<input type="text" name="docID"> <br>
	Google Sheets Data Url: <br>
		<input type="text" name="sheetID"> <br>
	No. of Entries in Sheet, for which Doc is to generate: <br>
		<input type="number" name="ent" min="1"> <br>
	No. of Columns (or tags) in the sheet: <br>
		<input type="number" name="cols" min="1"> <br>
	<input type="submit" value="Submit">
</form>
<br>
For more info, visit: <a href="https://github.com/Shivam010/populateg">github.com/Shivam010</a>
</body>
</html>`, per.Name); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("error on welcome page: %v", err)
	}
}

func Process(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p, err := FilPopulateObject(r.FormValue("docID"), r.FormValue("sheetID"), r.FormValue("ent"), r.FormValue("cols"))
	if err != nil {
		if _, err := fmt.Fprintf(w, `<html>
<body>
Error: %v
<br>
Try Again.
<br>
<form action="/process" method="post">
	Google Docs Template Url: <br>
		<input type="text" name="docID"> <br>
	Google Sheets Data Url: <br>
		<input type="text" name="sheetID"> <br>
	No. of Entries in Sheet, for which Doc is to generate: <br>
		<input type="number" name="ent" min="1"> <br>
	No. of Columns (or tags) in the sheet: <br>
		<input type="number" name="cols" min="1"> <br>
	<input type="submit" value="Submit">
</form>
<br>
For more info, visit: <a href="https://github.com/Shivam010/populateg">github.com/Shivam010</a>
</body>
</html>`, err.Error()); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			log.Fatalf("error on welcome page: %v", err)
		}
		return
	}
	if err := p.Process(); err != nil {
		if _, err := fmt.Fprintf(w, `<html>
<body>
Error: %v <br>
Try Again <a href="/login">Log In</a>
</body>
<br>
For more info, visit: <a href="https://github.com/Shivam010/populateg">github.com/Shivam010</a>
</html>`, err.Error()); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			log.Fatalf("error on welcome page: %v", err)
		}
		return
	}
	if _, err := fmt.Fprintf(w, `<html><body>Done</body></html>`); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("error on welcome page: %v", err)
	}
}

func googleVerification(w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
	if _, err := fmt.Fprintf(w, `google-site-verification: googleccd40724f8c32619.html`); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		log.Fatalf("error on welcome page: %v", err)
	}
}