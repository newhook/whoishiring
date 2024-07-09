package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/newhook/whoishiring/linkedin"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

//go:embed linkedin_resume.tmpl
var linkedInResumeContent string

var (
	linkedInResumeTemplate = template.Must(template.New("linkedin_resume").Parse(linkedInResumeContent))
)

func scrapeLinkedIn(ctx context.Context, link string) (string, error) {
	profile, err := linkedin.Person(ctx, link)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(profile)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var asjson map[string]any
	err = json.Unmarshal(b, &asjson)
	if err != nil {
		return "", errors.WithStack(err)
	}

	sb := &strings.Builder{}
	err = linkedInResumeTemplate.Execute(sb, asjson)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return sb.String(), nil
}
