package commands

import (
	"encoding/json"
	"errors"
	"fmt"
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
		CreatedAt     string `json:"created_at"`
		Assignee      *struct {
			Login string `json:"login"`
		} `json:"assignee"`
		User *struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"items"`
}

type ghOrg struct {
	Login string `json:"login"`
}

func newReviewsPRCmd() *cobra.Command {
	var org string
	var limit int
	var outputJSON bool
	var inferOrgs bool

	cmd := &cobra.Command{
		Use:     "reviews",
		Short:   "List open PRs where reviews are requested from you",
		Aliases: []string{"r"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build search query
			queryParts := []string{"is:open", "is:pr", "review-requested:@me", "archived:false"}
			if org != "" {
				queryParts = append(queryParts, fmt.Sprintf("org:%s", org))
			} else if inferOrgs {
				// Try to infer organizations for the authenticated user
				orgs, err := inferUserOrgs()
				if err == nil && len(orgs) > 0 {
					for _, o := range orgs {
						queryParts = append(queryParts, fmt.Sprintf("org:%s", o))
					}
				}
			}
			query := strings.Join(queryParts, "+")

			// Base command
			// Use gh api exactly as: gh api -X GET 'search/issues?q=is:open+is:pr+review-requested:@me+archived:false' --paginate
			apiURL := fmt.Sprintf("search/issues?q=%s", query)
			ghArgs := []string{"api", "-X", "GET", apiURL, "--paginate"}

			c := exec.Command("gh", ghArgs...)
			out, err := c.Output()
			if err != nil {
				var ee *exec.ExitError
				if errors.As(err, &ee) {
					return fmt.Errorf("gh api failed: %s", string(ee.Stderr))
				}
				return err
			}

			payload := strings.TrimSpace(string(out))
			if payload == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No PRs found.")
				return nil
			}

			// gh --paginate returns concatenated JSON documents separated by newlines.
			var merged ghSearchIssuesResponse
			for _, chunk := range strings.Split(payload, "\n") {
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

			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No PRs found.")
				return nil
			}

			for _, it := range items {
				repo := strings.TrimPrefix(it.RepositoryURL, "https://api.github.com/repos/")
				requester := ""
				if it.Assignee != nil && it.Assignee.Login != "" {
					requester = it.Assignee.Login
				} else if it.User != nil {
					requester = it.User.Login
				}
				if requester == "" {
					requester = "unknown"
				}
				// Print: requester, title, created_at
				fmt.Fprintf(cmd.OutOrStdout(), "%s | %s | %s | %s#%d | %s\n", requester, it.Title, it.CreatedAt, repo, it.Number, it.HTMLURL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Filter by GitHub organization")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results (0=all)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output raw JSON")
	cmd.Flags().BoolVar(&inferOrgs, "infer-orgs", true, "Infer your org memberships if --org not provided")
	return cmd
}

// inferUserOrgs returns the list of org logins for the authenticated user using gh api
func inferUserOrgs() ([]string, error) {
	c := exec.Command("gh", "api", "/user/orgs", "--paginate")
	out, err := c.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gh api orgs failed: %s", string(ee.Stderr))
		}
		return nil, err
	}
	var orgs []ghOrg
	// gh --paginate for list endpoints returns a single JSON array or multiple concatenated arrays.
	// We'll try to decode as an array; if it fails, split by newlines and merge.
	if err := json.Unmarshal(out, &orgs); err == nil {
		res := make([]string, 0, len(orgs))
		for _, o := range orgs {
			if o.Login != "" {
				res = append(res, o.Login)
			}
		}
		return res, nil
	}
	// Fallback: newline-separated arrays
	var res []string
	for _, chunk := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		var part []ghOrg
		if err := json.Unmarshal([]byte(chunk), &part); err != nil {
			continue
		}
		for _, o := range part {
			if o.Login != "" {
				res = append(res, o.Login)
			}
		}
	}
	return res, nil
}
