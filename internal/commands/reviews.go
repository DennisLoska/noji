package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dennisloska/noji/internal/commands/output"
	"github.com/spf13/cobra"
)

type ghIssueItem struct {
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
}

type ghSearchIssuesResponse struct {
	Items []ghIssueItem `json:"items"`
}

type ghOrg struct {
	Login string `json:"login"`
}

func safeOneLine(s string) string {
	// collapse newlines and excessive spaces for cleaner single-line fields
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func newReviewsPRCmd() *cobra.Command {
	var org string
	var limit int
	var outputJSON bool
	var inferOrgs bool
	var noBots bool
	var botsOnly bool
	var urlsOnly bool

	cmd := &cobra.Command{
		Use:     "reviews",
		Short:   "List open PRs where reviews are requested from you",
		Aliases: []string{"r"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// compile bot login regex once per invocation
			botRe := regexp.MustCompile(`(?i)(\[bot\]|-bot$|bot$|^github-actions(\[bot\])?$|^dependabot(\[bot\])?$|^renovate(\[bot\]|-bot)?$|^snyk(-bot)?$|^mergify(\[bot\])?$|copilot)`)
			// Build search query
			queryParts := []string{"is:open", "is:pr", "archived:false"}
			// Always limit to PRs requesting my review
			queryParts = append(queryParts, "review-requested:@me")
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
			// Optimize API usage: if a small limit is requested, avoid full pagination
			ghArgs := []string{"api", "-X", "GET", apiURL}
			perPage := 0
			if limit > 0 && limit <= 100 {
				perPage = limit
			} else if limit == 0 || limit > 100 {
				// Use high per_page to reduce round-trips when paginating
				perPage = 100
				ghArgs = append(ghArgs, "--paginate")
			}
			if perPage > 0 {
				ghArgs = append(ghArgs, fmt.Sprintf("-Fper_page=%d", perPage))
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

			payload := strings.TrimSpace(string(out))
			if payload == "" {
				output.Warnf(output.ModeAuto, "No PRs found.\n")
				return nil
			}

			// gh --paginate returns concatenated JSON documents separated by newlines.
			// However, when search results fit in one page, gh returns a single JSON object
			// possibly followed by a trailing newline and then another JSON object with only
			// the 'incomplete_results' and 'total_count' fields. We'll parse robustly by
			// attempting to decode the entire payload first; if that fails, fall back to
			// splitting by newlines and decoding each chunk that looks like a JSON object.
			var merged ghSearchIssuesResponse
			// Try whole-payload decode first
			if err := json.Unmarshal([]byte(payload), &merged); err != nil {
				// Fallback: line-delimited JSON documents
				merged = ghSearchIssuesResponse{}
				for _, chunk := range strings.Split(payload, "\n") {
					chunk = strings.TrimSpace(chunk)
					if chunk == "" {
						continue
					}
					var r ghSearchIssuesResponse
					if err := json.Unmarshal([]byte(chunk), &r); err != nil {
						// ignore non-matching chunks (like {"incomplete_results":...})
						continue
					}
					merged.Items = append(merged.Items, r.Items...)
				}
			}

			// Filter author by bot vs human according to flags. The query already
			// limits to PRs requesting my review.
			items := merged.Items
			filtered := make([]ghIssueItem, 0, len(items))
			for _, it := range items {
				author := ""
				if it.User != nil {
					author = it.User.Login
				}
				isBot := botRe.MatchString(author)
				if botsOnly && !isBot {
					continue
				}
				if !botsOnly && noBots && isBot {
					continue
				}
				filtered = append(filtered, it)
			}
			itemsAny := filtered

			// Apply limit if specified
			if limit > 0 && limit < len(itemsAny) {
				itemsAny = itemsAny[:limit]
			}

			if outputJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(itemsAny)
			}

			if urlsOnly {
				for _, it := range itemsAny {
					if strings.TrimSpace(it.HTMLURL) != "" {
						output.Printf(output.ModeNever, "%s\n", it.HTMLURL)
					}
				}
				return nil
			}

			if len(itemsAny) == 0 {
				output.Warnf(output.ModeAuto, "No PRs found.\n")
				return nil
			}

			for _, it := range itemsAny {
				author := ""
				if it.User != nil && it.User.Login != "" {
					author = it.User.Login
				}
				if author == "" {
					author = "unknown"
				}
				// plain output
				authorLabel := author
				if output.AuthorColorEnabled() {
					authorLabel = output.ColorizeAuthor(authorLabel)
				}
				output.Infof(output.ModeAuto, "PR:   #%d\n", it.Number)
				output.Printf(output.ModeAuto, "Title: %s\n", safeOneLine(it.Title))
				output.Printf(output.ModeAuto, "Author: %s\n", authorLabel)
				output.Printf(output.ModeAuto, "Created: %s\n", it.CreatedAt)
				output.Printf(output.ModeAuto, "URL:   %s\n\n", it.HTMLURL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Filter by GitHub organization")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results (0=all)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output raw JSON")
	cmd.Flags().BoolVar(&inferOrgs, "infer-orgs", true, "Infer your org memberships if --org not provided")
	cmd.Flags().BoolVar(&noBots, "no-bots", true, "Exclude PRs from bot authors")
	cmd.Flags().BoolVar(&botsOnly, "bots", false, "Show only PRs from bot authors (overrides --no-bots)")
	cmd.Flags().BoolVar(&urlsOnly, "urls", false, "Print only PR URLs (one per line)")
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
