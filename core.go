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
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type SheetData map[string][]string

func GetOauthConfig(ctx context.Context, state, code string) (*oauth2.Token, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

func ParseToken(r *http.Request) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	ck, err := r.Cookie("accessToken")
	if err != nil {
		return nil, err
	}
	token.AccessToken = ck.Value
	ck, err = r.Cookie("refreshToken")
	if err != nil {
		return nil, err
	}
	token.RefreshToken = ck.Value
	ck, err = r.Cookie("tokenType")
	if err != nil {
		return nil, err
	}
	token.TokenType = ck.Value
	ck, err = r.Cookie("expiry")
	if err != nil {
		return nil, err
	}
	expiry, err := time.Parse(time.RFC3339, ck.Value)
	if err != nil {
		return nil, err
	}
	token.Expiry = expiry
	return token, nil
}

func SaveToken(w http.ResponseWriter, token *oauth2.Token) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    token.AccessToken,
		Expires:  token.Expiry,
		Secure:   HostURL != "localhost:"+PORT,
		HttpOnly: true,
		Domain:   HostURL,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    token.RefreshToken,
		Expires:  token.Expiry,
		Secure:   HostURL != "localhost:"+PORT,
		HttpOnly: true,
		Domain:   HostURL,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "tokenType",
		Value:    token.TokenType,
		Expires:  token.Expiry,
		Secure:   HostURL != "localhost:"+PORT,
		HttpOnly: true,
		Domain:   HostURL,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "expiry",
		Value:    token.Expiry.Format(time.RFC3339),
		Expires:  token.Expiry,
		Secure:   HostURL != "localhost:"+PORT,
		HttpOnly: true,
		Domain:   HostURL,
	})
}

func GetUserInfo(ctx context.Context, client *http.Client) (*Person, error) {
	srv, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, status.Error(codes.Internal, "Unable to retrieve People Service: "+err.Error())
	}
	per, err := srv.People.Get("people/me").PersonFields("names,emailAddresses").Do()
	if err != nil {
		return nil, status.Error(codes.Internal, "Unable to Get data: "+err.Error())
	}
	return &Person{
		Name:  per.Names[0].DisplayName,
		Email: per.EmailAddresses[0].Value,
	}, nil
}

type PopulateObject struct {
	DocID, SheetID   string
	Entries, Columns int64
}

func ParseUrlToID(url string) (string, error) {
	var reg = regexp.MustCompile(`(?m)https://docs.google.com/(document|spreadsheets)/(d|u/[0-9]*/d)/([^/]*)`)
	x := reg.FindAllStringSubmatch(url, 4)
	if len(x) < 1 || len(x[0]) < 4 {
		return "", fmt.Errorf("invalid Url is passed")
	}
	return x[0][3], nil
}

func FilPopulateObject(d, s, e, c string) (p *PopulateObject, err error) {
	p = &PopulateObject{}
	p.DocID, err = ParseUrlToID(d)
	if err != nil {
		return
	}
	p.SheetID, err = ParseUrlToID(s)
	if err != nil {
		return
	}
	p.Columns, err = strconv.ParseInt(c, 10, 64)
	if err != nil {
		return p, fmt.Errorf("a number is required in column")
	}
	p.Entries, err = strconv.ParseInt(e, 10, 64)
	if err != nil {
		return p, fmt.Errorf("a number is required in entries")
	}
	return
}

func column(i int64) string {
	c := ""
	for i > 0 {
		c = string('A'+((i-1)%26)) + c
		i = i / 26
	}
	return c
}

func (p *PopulateObject) GetSheetData(client *http.Client, sheetID string) (SheetData, error) {
	if client == nil {
		return nil, fmt.Errorf("client expired")
	}
	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Println("Unable to retrieve Sheets Service:", err)
		return nil, fmt.Errorf("unable to retrieve Sheet: check your access")
	}
	c3 := fmt.Sprintf("%v%v", column(p.Columns), p.Entries+1)

	res, err := srv.Spreadsheets.Values.Get(sheetID, fmt.Sprintf("A1:%v", c3)).MajorDimension("COLUMNS").Do()
	if err != nil {
		log.Println("Unable to get data from sheet:", err)
		return nil, fmt.Errorf("unable to read Sheet: check your access")
	}
	shData := make(SheetData, 0)
	for _, values := range res.Values {
		tag := ""
		for i, v := range values {
			if i == 0 {
				tag = fmt.Sprintf("%v", v)
				shData[tag] = make([]string, 0)
			} else {
				shData[tag] = append(shData[tag], fmt.Sprintf("%v", v))
			}
		}
	}
	return shData, nil
}

func (p *PopulateObject) CreateNewDocInDrive(client *http.Client, docID, newTitle string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("client expired")
	}
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Println("Unable to retrieve Drive Service:", err)
		return "", fmt.Errorf("unable to access drive: check your access")
	}

	file := &drive.File{Title: newTitle}
	res, err := srv.Files.Copy(docID, file).Do()
	if err != nil {
		log.Println("Unable to create new copy of template:", err)
		return "", fmt.Errorf("unable to create new copy of template in drive: check your access")
	}
	return res.Id, nil
}

func (p *PopulateObject) UpdateNewDoc(client *http.Client, docID string, ind int64, shData SheetData) error {
	if client == nil {
		return fmt.Errorf("client expired")
	}
	srv, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Println("Unable to retrieve Docs Service:", err)
		return fmt.Errorf("unable to populate Document: check your access")
	}
	dsrv := docs.NewDocumentsService(srv)
	drs := make([]*docs.Request, 0, len(shData))
	for tag, entries := range shData {
		if len(entries) < int(ind) {
			continue
		}
		drs = append(drs, &docs.Request{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					MatchCase: false,
					Text:      "{{" + tag + "}}",
				},
				ReplaceText: entries[ind-1],
			},
		})
	}
	if len(drs) == 0 {
		// Impossible case: though nothing to update
		return nil
	}
	req := docs.BatchUpdateDocumentRequest{
		Requests: drs,
	}
	res, err := dsrv.BatchUpdate(docID, &req).Do()
	if err != nil {
		log.Println("Unable to populate document:", err)
		return fmt.Errorf("unable to populate Document: check your access")
	}
	if res.HTTPStatusCode == 200 {
		return nil
	}
	log.Println("something went wrong. Response:", res)
	return fmt.Errorf("something went wrong")
}

type Response struct {
	// Response For Document: DocNo,
	// If any error occurred in creation of DocNo's Document,
	// then the error message is returned in ErrorMessage.
	DocNo        int64
	ErrorMessage string
}

func (p *PopulateObject) Process(client *http.Client) ([]Response, error) {

	shData, err := p.GetSheetData(client, p.SheetID)
	if err != nil {
		return nil, err
	}

	res := make([]Response, 0, p.Entries)
	for ind := int64(1); ind <= p.Entries; ind++ {
		newTitle := fmt.Sprintf("Doc %v", ind)
		if fileNames, ok := shData["fileName"]; ok {
			if len(fileNames) >= int(ind) {
				newTitle = fileNames[ind-1]
			}
		}
		nID, err := p.CreateNewDocInDrive(client, p.DocID, newTitle)
		if err != nil {
			res = append(res, Response{
				DocNo:        ind,
				ErrorMessage: err.Error(),
			})
			continue
		}
		if err = p.UpdateNewDoc(client, nID, ind, shData); err != nil {
			res = append(res, Response{
				DocNo:        ind,
				ErrorMessage: err.Error(),
			})
			continue
		}
	}
	return res, nil
}
