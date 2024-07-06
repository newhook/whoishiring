package main

import (
	"context"
	_ "embed"
	"github.com/newhook/ai-hack/claude"
	"github.com/newhook/ai-hack/openai"
	"github.com/pkg/errors"
	"text/template"
)

//go:embed prompts/job_search.tmpl
var jobSearchPrompt string

//go:embed prompts/search_terms.tmpl
var searchTermsPrompt string

var (
	searchTermsTemplate = template.Must(template.New("search_terms").Parse(searchTermsPrompt))
	jobSearchTemplate   = template.Must(template.New("jobSearch").Parse(jobSearchPrompt))
)

type Completion struct {
	Model    string
	GetTerms func(ctx context.Context, context any) ([]string, error)
	GetJobs  func(ctx context.Context, context any) ([]string, error)
}

const (
	Claude = "claude"
	OpenAI = "openai"
)

var completions = map[string]Completion{
	Claude: {
		Model: Claude,
		GetTerms: func(ctx context.Context, jobPrompt any) ([]string, error) {
			var terms []string
			resp, err := claude.Completions(ctx, "terms", *fake, searchTermsTemplate, jobPrompt)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			choice := resp.Content[len(resp.Content)-1]
			if err := claude.ParseJsonResponse(choice.Text, &terms); err != nil {
				return nil, errors.WithStack(err)
			}
			return terms, nil
		},
		GetJobs: func(ctx context.Context, context any) ([]string, error) {
			var jobIDs []string
			r2, err := claude.Completions(ctx, "job_search", *fake, jobSearchTemplate, context)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			choice := r2.Content[len(r2.Content)-1]
			if err := claude.ParseJsonResponse(choice.Text, &jobIDs); err != nil {
				return nil, errors.WithStack(err)
			}
			return jobIDs, nil
		},
	},
	OpenAI: {
		Model: OpenAI,
		GetTerms: func(ctx context.Context, jobPrompt any) ([]string, error) {
			var terms []string
			resp, err := openai.Completions(ctx, "terms", *fake, searchTermsTemplate, jobPrompt)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			choice := resp.Choices[len(resp.Choices)-1]
			if err := openai.ParseJsonResponse(choice, &terms); err != nil {
				return nil, errors.WithStack(err)
			}
			return terms, nil
		},
		GetJobs: func(ctx context.Context, context any) ([]string, error) {
			var jobIDs []string
			r2, err := openai.Completions(ctx, "job_search", *fake, jobSearchTemplate, context)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			if err := openai.ParseJsonResponse(r2.Choices[len(r2.Choices)-1], &jobIDs); err != nil {
				return nil, errors.WithStack(err)
			}
			return jobIDs, nil
		},
	},
}

func GetTerms(ctx context.Context, context any) ([]string, error) {
	return completions[*completionModel].GetTerms(ctx, context)
}

func GetJobs(ctx context.Context, context any) ([]string, error) {
	return completions[*completionModel].GetJobs(ctx, context)
}
