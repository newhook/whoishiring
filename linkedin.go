package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"github.com/newhook/whoishiring/linkedin"
	"github.com/newhook/whoishiring/queries"
	"github.com/pkg/errors"
	"strings"
	"text/template"
	"time"
)

//go:embed linkedin_resume.tmpl
var linkedInResumeContent string

var (
	linkedInResumeTemplate = template.Must(template.New("linkedin_resume").Parse(linkedInResumeContent))
)

func scrapeLinkedIn(ctx context.Context, q *queries.Queries, link string) (string, error) {
	var jsonBody []byte
	saved, err := q.GetLinkedInScrape(ctx, link)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.WithStack(err)
		}

		profile, err := linkedin.Person(ctx, link)
		if err != nil {
			return "", err
		}
		jsonBody, err = json.Marshal(profile)
		if err != nil {
			return "", errors.WithStack(err)
		}
		now := int(time.Now().Unix())
		err = q.InsertLinkedInScrape(ctx, queries.InsertLinkedInScrapeParams{
			Url:       link,
			Json:      string(jsonBody),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return "", errors.WithStack(err)
		}
	} else {
		jsonBody = []byte(saved.Json)
	}

	var raw map[string]any
	err = json.Unmarshal(jsonBody, &raw)
	if err != nil {
		return "", errors.WithStack(err)
	}

	sb := &strings.Builder{}
	err = linkedInResumeTemplate.Execute(sb, raw)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return sb.String(), nil
}
