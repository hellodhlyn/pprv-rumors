package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Netflix/go-env"
	"github.com/dstotijn/go-notion"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

type Environment struct {
	ApiKey      string `env:"NOTION_API_KEY"`
	RootBlockID string `env:"ROOT_BLOCK_ID"`
}

var notionClient *notion.Client
var environment Environment

const (
	PropsKeyTitle    = "Title"
	PropsKeyDate     = "Date"
	PropsKeySource   = "Source"
	PropsKeyReleased = "Released"
)

func FetchSubjects(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	doc, err := notionClient.FindBlockChildrenByID(r.Context(), environment.RootBlockID, nil)
	if err != nil {
		log.Errorf("failed to get block children: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var subjects []SubjectResponse
	for _, result := range doc.Results {
		if result.Type != notion.BlockTypeUnsupported {
			continue
		}

		doc, err := notionClient.FindDatabaseByID(r.Context(), result.ID)
		if err != nil {
			log.Errorf("failed to get database: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subject := SubjectResponse{
			ID:        result.ID,
			Title:     doc.Title[0].PlainText,
			UpdatedAt: result.LastEditedTime,
		}
		if len(doc.Title) == 2 {
			subject.Description = doc.Title[1].PlainText
		}
		subjects = append(subjects, subject)
	}

	w.Header().Set("Content-Type", "application/json; encode=utf-8")
	_ = json.NewEncoder(w).Encode(subjects)
}

func FetchSubject(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	doc, err := notionClient.FindDatabaseByID(r.Context(), p.ByName("id"))
	if err != nil {
		log.Errorf("failed to get database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subject := SubjectResponse{
		ID:        doc.ID,
		Title:     doc.Title[0].PlainText,
		UpdatedAt: &doc.LastEditedTime,
	}
	if len(doc.Title) == 2 {
		subject.Description = doc.Title[1].PlainText
	}

	w.Header().Set("Content-Type", "application/json; encode=utf-8")
	_ = json.NewEncoder(w).Encode(subject)
}

func FetchSubjectRumors(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	query, err := notionClient.QueryDatabase(r.Context(), p.ByName("id"), &notion.DatabaseQuery{
		Sorts: []notion.DatabaseQuerySort{{Property: "Date", Direction: notion.SortDirDesc}}},
	)
	if err != nil {
		log.Errorf("failed to query database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var rumors []RumorResponse
	for _, item := range query.Results {
		props := item.Properties.(notion.DatabasePageProperties)
		if props[PropsKeyReleased].Select == nil || props[PropsKeyReleased].Select.Name != "released" {
			continue
		}

		doc, err := notionClient.FindBlockChildrenByID(r.Context(), item.ID, nil)
		if err != nil {
			log.Errorf("failed to get database: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var bodies []string
		for _, result := range doc.Results {
			if len(result.Paragraph.Text) != 0 {
				bodies = append(bodies, result.Paragraph.Text[0].PlainText)
			}
		}

		rumors = append(rumors, RumorResponse{
			Title:  props[PropsKeyTitle].Title[0].PlainText,
			Date:   &props[PropsKeyDate].Date.Start.Time,
			Source: props[PropsKeySource].RichText[0].PlainText,
			Body:   strings.Join(bodies, "\n"),
		})
	}

	w.Header().Set("Content-Type", "application/json; encode=utf-8")
	_ = json.NewEncoder(w).Encode(rumors)
}

func main() {
	_, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		log.Fatal(err)
	}

	notionClient = notion.NewClient(environment.ApiKey)

	router := httprouter.New()
	router.GET("/subjects", FetchSubjects)
	router.GET("/subjects/:id", FetchSubject)
	router.GET("/subjects/:id/rumors", FetchSubjectRumors)

	log.Fatal(http.ListenAndServe(":8080", router))
}
