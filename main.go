package main

import (
	"context"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"github.com/MatusOllah/slogcolor"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/newhook/whoishiring/queries"
	"github.com/pkg/errors"
	slogecho "github.com/samber/slog-echo"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	db                *sql.DB
	q                 *queries.Queries
	whoIsHiring       = "Ask HN: Who is hiring?%"
	whoWantsToBeHired = "Ask HN: Who wants to be hired?%"
)

const (
	MaxWindow = 6
)

//go:embed banner
var banner string

//go:embed schema.sql
var ddl string

var (
	fake            = flag.Bool("fake", false, "use fake data")
	completionModel = flag.String("completion", Claude, "completion model")
	embeddingModel  = flag.String("embedding", OpenAI3Small, "embedding model")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	//l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	l := slog.New(slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions))
	if err := run(ctx, l); err != nil {
		l.Error("failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, l *slog.Logger) error {
	fmt.Println(banner)

	var err error
	dbPath := "./whoishiring.db"
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()
	l.Info("database opened", slog.String("path", dbPath))
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return errors.WithStack(err)
	}

	q = queries.New(db)

	if err := FetchPosts(ctx, l, q); err != nil {
		return err
	}

	for model := range embeddings {
		if err := CreateEmbeddings(ctx, l, model); err != nil {
			return err
		}
	}

	//if err := PrintTokens(ctx); err != nil {
	//	return err
	//}

	e := echo.New()
	e.Use(slogecho.New(l))
	// For debugging this one is a pain.
	//e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
	//	Skipper:      middleware.DefaultSkipper,
	//	ErrorMessage: "custom timeout error message returns to client",
	//	OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
	//		l.Error("timeout", slog.String("path", c.Path()), slog.String("error", err.Error()))
	//	},
	//	Timeout: 30 * time.Second,
	//}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},                                        // Allow all origins
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE}, // Specify allowed methods
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.POST("/jobs", func(c echo.Context) error {
		if err := c.Request().ParseMultipartForm(32 << 20); err != nil { // 32 MB max memory
			return err
		}

		monthsParam := c.FormValue("months")
		prompt := c.FormValue("prompt")
		searchType := c.FormValue("type")
		linkedin := c.FormValue("linkedin")
		form, err := c.MultipartForm()
		if err != nil {
			log.Println(err.Error())
			return err
		}
		var terms SearchTerms
		for _, files := range form.File {
			if len(files) != 1 {
				return c.String(http.StatusBadRequest, "Invalid number of files")
			}
			file := files[0]

			f, err := file.Open()
			if err != nil {
				return err
			}
			defer f.Close()

			terms.ResumeName = file.Filename
			terms.Resume = f
			terms.Size = file.Size
		}

		terms.Months, err = strconv.Atoi(monthsParam)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid months parameter")
		}

		terms.LinkedIn = linkedin
		terms.JobPrompt = prompt

		if searchType == "hiring" {
			terms.SearchType = SearchType_WhoIsHiring
		} else if searchType == "seekers" {
			terms.SearchType = SearchType_WhoWantToBeHired
		}

		resp, err := JobSearch(c.Request().Context(), l, terms)
		if err != nil {
			l.Error("job search failed", slog.String("error", err.Error()))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		var links []string
		for _, id := range resp.Comments {
			links = append(links, fmt.Sprintf("https://news.ycombinator.com/item?id=%d", id))
		}
		return c.JSON(http.StatusOK, map[string]any{
			"comments":          resp.Comments,
			"parents":           resp.Parents,
			"items":             resp.Items,
			"original_comments": resp.OriginalComments,
			"original_parents":  resp.OriginalParents,
			"hackerNewsLinks":   links,
			"resumeSummary":     resp.ResumeSummary,
			"searchTerms":       resp.SearchTerms,
			"posts":             resp.Posts,
			"itemsSearched":     resp.ItemsSearched,
		})
	})

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return errors.WithStack(err)
	}
	defer ln.Close()

	s := http.Server{
		Handler: e,
		//ReadTimeout: 30 * time.Second, // customize http.Server timeouts
	}

	l.Info("using configuration", slog.String("embedding", *embeddingModel), slog.String("completion", *completionModel))
	l.Info("starting http server", slog.String("addr", ln.Addr().String()))
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := s.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			return errors.WithStack(err)
		}
		return nil
	})

	<-ctx.Done()

	// start gracefully shutdown with a timeout of 10 seconds.
	ctx, cancelGC := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelGC()

	if err := s.Shutdown(ctx); err != nil {
		return errors.WithStack(err)
	}

	return g.Wait()
}
