package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := execMain(); err != nil {
		panic(err)
	}
}

func execMain() error {
	ctx := context.Background()
	ghClient := getGitHubClient(ctx)
	urls := strings.Split(os.Getenv("GITHUB_ACTION_CONFIG_URLS"), ",")

	var values [][]interface{}
	header := []interface{}{
		"id", "repo", "workflow", "branch", "started_at", "ended_at", "actor", "conclusion", "url",
	}
	values = append(values, header)

	for _, url := range urls {
		rows, err := fetchRuns(ctx, ghClient, url)
		if err != nil {
			return err
		}
		values = append(values, rows...)
	}

	ggClient, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return err
	}

	if err := updateSheet(ctx, ggClient, values); err != nil {
		return err
	}
	return nil
}

func getGitHubClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: os.Getenv("GITHUB_TOKEN"),
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func fetchRuns(ctx context.Context, client *github.Client, url string) ([][]interface{}, error) {
	ss := strings.Split(url, "/")
	org := ss[3]
	repo := ss[4]
	file := ss[len(ss)-1]

	var rows [][]interface{}

	for i := 0; ; i++ {
		opt := &github.ListWorkflowRunsOptions{
			ListOptions: github.ListOptions{
				PerPage: 100, // max
				Page:    i + 1,
			},
		}
		runs, _, err := client.Actions.ListWorkflowRunsByFileName(ctx, org, repo, file, opt)

		if err != nil {
			return nil, err
		}

		for _, r := range runs.WorkflowRuns {
			row := []interface{}{
				r.ID, r.Repository.FullName, r.Name, r.HeadBranch, r.RunStartedAt, r.UpdatedAt, r.Actor.Login, r.Conclusion, r.HTMLURL,
			}
			rows = append(rows, row)
		}

		if len(runs.WorkflowRuns) < 1 {
			break
		}
	}

	return rows, nil
}

func updateSheet(ctx context.Context, client *http.Client, values [][]interface{}) error {
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	spreadsheetId := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("TARGET_SHEET_NAME")
	valueInputOption := "USER_ENTERED"
	data := []*sheets.ValueRange{}
	vr := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          fmt.Sprintf("%s!A1", sheetName),
		Values:         values,
	}
	data = append(data, vr)

	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data:             data,
	}

	if _, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, rb).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
}
