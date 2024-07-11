package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/newhook/whoishiring/queries"
	"github.com/pkg/errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SearchType int

const (
	SearchType_WhoIsHiring      SearchType = iota
	SearchType_WhoWantToBeHired SearchType = iota
)

type Reader interface {
	io.ReaderAt
	io.Reader
}
type SearchTerms struct {
	Months int

	SearchType SearchType

	JobPrompt string

	LinkedIn string

	ResumeName string
	Resume     Reader
	Size       int64
}

type JobSearchResponse struct {
	Items            []queries.Item
	Comments         []int
	Parents          []int
	HNLinks          []string
	ResumeSummary    string
	SearchTerms      []string
	TotalPosts       int
	TotalItems       int
	Posts            int
	ItemsSearched    int
	Latencies        map[string]float64
	OriginalComments []int
	OriginalParents  []int
}

func JobSearch(ctx context.Context, l *slog.Logger, q *queries.Queries, search SearchTerms) (JobSearchResponse, error) {
	resp := JobSearchResponse{
		Latencies: map[string]float64{},
	}
	start := time.Now()
	recordLatency := func(step string) {
		resp.Latencies[step] = time.Since(start).Seconds()
		start = time.Now()
	}
	var clause string
	switch search.SearchType {
	case SearchType_WhoIsHiring:
		clause = whoIsHiring
	case SearchType_WhoWantToBeHired:
		clause = whoWantsToBeHired
	default:
		return resp, errors.Errorf("invalid search type: %s", search.SearchType)
	}

	resume := ""
	if search.LinkedIn != "" {
		var err error
		resume, err = scrapeLinkedIn(ctx, q, search.LinkedIn)
		if err != nil {
			return resp, err
		}
		recordLatency("scrape_linkedin")
	} else if search.Resume != nil && strings.HasSuffix(search.ResumeName, "pdf") {
		file, err := os.CreateTemp("", "*.pdf")
		if err != nil {
			return resp, errors.WithStack(err)
		}
		defer os.Remove(file.Name())
		f, err := os.OpenFile(file.Name(), os.O_RDWR, 0644)
		if err != nil {
			return resp, errors.WithStack(err)
		}
		defer f.Close()
		_, err = io.Copy(f, search.Resume)
		if err != nil {
			return resp, errors.WithStack(err)
		}

		// See "man pdftotext" for more options.
		args := []string{
			"-layout",   // Maintain (as best as possible) the original physical layout of the text.
			"-nopgbrk",  // Don't insert page breaks (form feed characters) between pages.
			file.Name(), // The input file.
			"-",         // Send the output to stdout.
		}
		cmd := exec.CommandContext(context.Background(), "pdftotext", args...)

		var buf bytes.Buffer
		cmd.Stdout = &buf

		if err := cmd.Run(); err != nil {
			return resp, errors.WithStack(err)
		}

		resume = buf.String()
	} else if search.Resume != nil {
		b, err := io.ReadAll(search.Resume)
		if err != nil {
			return resp, errors.WithStack(err)
		}
		resume = string(b)
	}

	if len(resume) > 0 {
		analyze, err := AnalyzeResume(ctx, resume)
		if err != nil {
			return resp, err
		}
		recordLatency("analyze_resume")
		resp.ResumeSummary = analyze
		if search.JobPrompt != "" {
			search.JobPrompt = fmt.Sprintf("%s\nIn addition consider the following %s", search.JobPrompt, analyze)
		} else {
			search.JobPrompt = fmt.Sprintf("I'm looking for a job that best matches this description\n%s", analyze)
		}
	}

	terms, err := GetTerms(ctx, search.JobPrompt)
	if err != nil {
		return resp, err
	}
	recordLatency("get_terms")
	resp.SearchTerms = terms

	limit := 10
	queryResults, err := VectorSearch(ctx, l, q, search.Months, *embeddingModel, clause, terms, limit)
	if err != nil {
		return resp, err
	}
	recordLatency("vector_search")
	resp.ItemsSearched = queryResults.Searched
	resp.Posts = queryResults.Posts
	resp.TotalItems = queryResults.TotalItems
	resp.TotalPosts = queryResults.TotalPosts

	for _, result := range queryResults.Results {
		resp.OriginalComments = append(resp.OriginalComments, result.Item.ID)
		resp.OriginalParents = append(resp.OriginalParents, result.Item.Parent)
	}

	type jobDescription struct {
		ID      int    `json:"id"`
		Date    string `json:"date"`
		Content string `json:"content"`
	}
	var descriptions []jobDescription
	for _, result := range queryResults.Results {
		descriptions = append(descriptions, jobDescription{
			ID:      result.ID,
			Date:    time.Unix(int64(result.Item.Time), 0).String(),
			Content: result.Item.Text,
		})
	}

	jobIDs, err := GetJobs(ctx, map[string]any{
		"Prompt": search.JobPrompt,
		"Jobs":   descriptions,
	})
	if err != nil {
		return resp, err
	}
	recordLatency("get_jobs")

	// The comments must be contained in the original query results.
	for _, id := range jobIDs {
		n, err := strconv.Atoi(id)
		if err != nil {
			return resp, err
		}
		for _, result := range queryResults.Results {
			if result.ID == n {
				resp.Comments = append(resp.Comments, result.Item.ID)
				resp.Parents = append(resp.Parents, result.Item.Parent)
				break
			}
		}
	}

	parentSet := NewSet[int]()
	for _, result := range queryResults.Results {
		parentSet.Add(result.Item.Parent)
		resp.Items = append(resp.Items, result.Item)
	}

	allParents, err := q.GetItems(ctx, parentSet.Values())
	if err != nil {
		return resp, errors.WithStack(err)
	}
	for _, parent := range allParents {
		resp.Items = append(resp.Items, parent)
	}
	return resp, nil
}
