package main

import (
	"context"
	"fmt"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"regexp"
	"strconv"
)

type SheetData map[string][]string

func GetOauthConfig(ctx context.Context, state, code string) error {
	if state != oauthStateString {
		return fmt.Errorf("invalid oauth state")
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("code exchange failed: %s", err.Error())
	}
	client = config.Client(ctx, token)
	return nil
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
		return "", fmt.Errorf("invalid url passed")
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
		return
	}
	p.Entries, err = strconv.ParseInt(e, 10, 64)
	if err != nil {
		return
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

func (p *PopulateObject) GetSheetData(sheetID string) (SheetData, error) {
	if client == nil {
		return nil, fmt.Errorf("client expired")
	}
	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets Service: %v", err)
	}
	c3 := fmt.Sprintf("%v%v", column(p.Columns), p.Entries)

	res, err := srv.Spreadsheets.Values.Get(sheetID, fmt.Sprintf("A1:%v", c3)).MajorDimension("COLUMNS").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get data from sheet: %v", err.Error())
	}
	mp := make(SheetData, 0)
	for _, r := range res.Values {
		in := ""
		for i, v := range r {
			if i == 0 {
				in = fmt.Sprintf("%v", v)
				mp[in] = make([]string, 0)
			} else {
				mp[in] = append(mp[in], fmt.Sprintf("%v", v))
			}
		}
	}
	return mp, nil
}

func (p *PopulateObject) CreateNewDocInDrive(docID, newTitle string) (string, error) {
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("unable to retrieve Drive Service: %v", err.Error())
	}

	file := &drive.File{Title: newTitle}
	res, err := srv.Files.Copy(docID, file).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create new copy of template: %v", err.Error())
	}
	return res.Id, nil
}

func (p *PopulateObject) UpdateNewDoc(docID string, ind int64, mp SheetData) error {
	srv, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Docs Service: %v", err.Error())
	}
	dsrv := docs.NewDocumentsService(srv)
	drs := make([]*docs.Request, 0, len(mp))
	for k, v := range mp {
		if len(v) <= int(ind) {
			continue
		}
		drs = append(drs, &docs.Request{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					MatchCase: false,
					Text:      "{{" + k + "}}",
				},
				ReplaceText: v[ind],
			},
		}, )
	}
	req := docs.BatchUpdateDocumentRequest{
		Requests: drs,
	}
	res, err := dsrv.BatchUpdate(docID, &req).Do()
	if err != nil {
		return fmt.Errorf("unable to populate document: %v", err.Error())
	}
	if res.HTTPStatusCode == 200 {
		return nil
	}
	return fmt.Errorf("something went wrong")
}

func (p *PopulateObject) Process() error {
	p.Entries++
	mp, err := p.GetSheetData(p.SheetID)
	if err != nil {
		return err
	}
	tags := make([]string, 0, len(mp))
	for t := range mp {
		tags = append(tags, t)
	}
	for i := int64(0); i+1 < p.Entries; i++ {
		newTitle := fmt.Sprintf("Doc %v", i)
		nID, err := p.CreateNewDocInDrive(p.DocID, newTitle)
		if err != nil {
			return err
		}
		if err = p.UpdateNewDoc(nID, i, mp); err != nil {
			return err
		}
	}
	return err
}
