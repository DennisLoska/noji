package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type ghSearchIssuesResponse struct {
	Items []struct {
		Number        int    `json:"number"`
		Title         string `json:"title"`
		HTMLURL       string `json:"html_url"`
		RepositoryURL string `json:"repository_url"`
	} `json:"items"`
}

func newReviewsPRCmd() *cobra.Command {
	var org string
	var limit int
	var outputJSON bool

	cmd := &cobra.Command{
		Use:     "reviews",
		Short:   "List open PRs where reviews are requested from you",
		Aliases: []string{"r"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build search query
			queryParts := []string{"is:open", "is:pr", "review-requested:@me", "archived:false"}
			if org != "" {
				queryParts = append(queryParts, fmt.Sprintf("org:%s", org))
			}
			query := strings.Join(queryParts, "+")
			q := url.QueryEscape(query)

			// Base command
			apiURL := fmt.Sprintf("search/issues?q=%s&per_page=%d", q, 100)
			ghArgs := []string{"api", "-X", "GET", apiURL, "--paginate"}
			if outputJSON {
				// We'll print raw JSON of aggregated items
			} else {
				// No jq; we'll parse JSON in Go
			}

			c := exec.Command("gh", ghArgs...)
			out, err := c.Output()
			if err != nil {
				var ee *exec.ExitError
				if errors.As(err, &ee) {
					return fmt.Errorf("gh api failed: %s", string(ee.Stderr))
				}
				return err
			}

			// gh --paginate returns a concatenation of JSON docs separated by newlines.
			// We'll split and merge .items arrays.

			var merged ghSearchIssuesResponse
			for _, chunk := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if strings.TrimSpace(chunk) == "" {
					continue
				}
				var r ghSearchIssuesResponse
				if err := json.Unmarshal([]byte(chunk), &r); err != nil {
					return fmt.Errorf("parse gh api response: %w", err)
				}
				merged.Items = append(merged.Items, r.Items...)
			}

			// Apply limit if specified
			items := merged.Items
			if limit > 0 && limit < len(items) {
				items = items[:limit]
			}

			if outputJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(items)
			}

			for _, it := range items {
				repo := strings.TrimPrefix(it.RepositoryURL, "https://api.github.com/repos/")
				fmt.Fprintf(cmd.OutOrStdout(), "%s#%d %s %s\n", repo, it.Number, it.Title, it.HTMLURL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Filter by GitHub organization")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results (0=all)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output raw JSON")
	return cmd
}
