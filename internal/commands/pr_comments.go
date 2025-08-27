package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/dennis/noji/internal/commands/output"
	"github.com/dennis/noji/internal/config"
	"github.com/spf13/cobra"
)

// Data models for gh api responses we need

type ghPR struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
	State string `json:"state"`
}

type ghIssueComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

type ghReviewComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Path           string `json:"path"`
	DiffHunk       string `json:"diff_hunk"`
	InReplyToID    *int64 `json:"in_reply_to_id"`
	PullRequestURL string `json:"pull_request_url"`
}

type classifiedComment struct {
	Kind      string // issue|review
	ID        int64
	Author    string
	CreatedAt string
	Body      string
	URL       string
	Path      string
	ParentID  int64  // 0 if none
	Severity  string // from opencode classification
}

type prWithComments struct {
	Repo     string
	Number   int
	Title    string
	URL      string
	Author   string
	Comments []classifiedComment
	Priority string // derived from comments severities
}

func newPRCommentsCmd() *cobra.Command {
	var repo string
	var state string
	var jsonOut bool
	var excludeBots bool
	var includeDrafts bool
	var limit int
	var since string
	var doClassify bool
	var renderMD bool
	var urlsOnly bool

	cmd := &cobra.Command{
		Use:   "comments",
		Short: "List your PRs with human comments (optional severity classification)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Who am I
			me, err := whoAmI()
			if err != nil {
				return err
			}
			botRe := regexp.MustCompile(`(?i)(\[bot\]$|-bot$|^github-actions(\[bot\])?$|^dependabot(\[bot\])?$|^renovate(\[bot\]|-bot)?$|^snyk(-bot)?$|^mergify(\[bot\])?$|copilot)`)

			// Find PRs authored by me (prefilter: comments>0; optional since)
			prs, err := listMyPRs(me, repo, state, includeDrafts, limit, since)
			if err != nil {
				return err
			}
			if len(prs) == 0 {
				output.Warnf(output.ModeAuto, "No PRs found.\n")
				return nil
			}

			var results []prWithComments
			for _, pr := range prs {
				repoFull, err := repoFromPRURL(pr.HTMLURL)
				if err != nil {
					continue
				}
				// Fast probes: check for human comments with per_page=1
				hasHuman, err := hasHumanComments(repoFull, pr.Number, botRe)
				if err != nil {
					return err
				}
				if !hasHuman && excludeBots {
					// Skip heavy fetch, no human activity
					continue
				}
				// Fetch full comments only when necessary
				issues, err := fetchIssueComments(repoFull, pr.Number)
				if err != nil {
					return err
				}
				reviews, err := fetchReviewComments(repoFull, pr.Number)
				if err != nil {
					return err
				}
				var cc []classifiedComment
				for _, ic := range issues {
					if excludeBots && botRe.MatchString(ic.User.Login) {
						continue
					}
					cc = append(cc, classifiedComment{
						Kind:      "issue",
						ID:        ic.ID,
						Author:    ic.User.Login,
						CreatedAt: ic.CreatedAt,
						Body:      ic.Body,
						URL:       ic.HTMLURL,
					})
				}
				// Build threading for review comments
				idToIndex := map[int64]int{}
				for _, rc := range reviews {
					if excludeBots && botRe.MatchString(rc.User.Login) {
						continue
					}
					parent := int64(0)
					if rc.InReplyToID != nil {
						parent = *rc.InReplyToID
					}
					cc = append(cc, classifiedComment{
						Kind:      "review",
						ID:        rc.ID,
						Author:    rc.User.Login,
						CreatedAt: rc.CreatedAt,
						Body:      rc.Body,
						URL:       rc.HTMLURL,
						Path:      rc.Path,
						ParentID:  parent,
					})
					idToIndex[rc.ID] = len(cc) - 1
				}
				// Optionally classify severity per comment using opencode
				priority := "none"
				if doClassify {
					model, _ := config.GetModel()
					for i := range cc {
						sev, _ := classifyComment(model, cc[i].Body)
						cc[i].Severity = sev
					}
					// Compute PR priority: highest severity among comments
					priority = derivePriority(cc)
				}
				// Sort comments by time
				sort.Slice(cc, func(i, j int) bool { return cc[i].CreatedAt < cc[j].CreatedAt })
				results = append(results, prWithComments{
					Repo:     repoFull,
					Number:   pr.Number,
					Title:    pr.Title,
					URL:      pr.HTMLURL,
					Author:   pr.User.Login,
					Comments: cc,
					Priority: priority,
				})
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}

			if urlsOnly {
				for _, r := range results {
					if strings.TrimSpace(r.URL) != "" {
						output.Printf(output.ModeNever, "%s\n", r.URL)
					}
				}
				return nil
			}

			// Human output
			for _, r := range results {
				output.Infof(output.ModeAuto, "PR: #%d %s\n", r.Number, r.Title)
				output.Printf(output.ModeAuto, "Repo: %s\n", r.Repo)
				// Raw PR URL only (no clickable label line)
				output.Printf(output.ModeAuto, "URL:  %s\n", r.URL)
				if doClassify {
					output.Printf(output.ModeAuto, "Priority: %s\n", r.Priority)
				}
				if len(r.Comments) == 0 {
					output.Warnf(output.ModeAuto, "  (no human comments)\n\n")
					continue
				}
				// Group review replies by thread using ParentID, but print simply with indentation
				indexByID := map[int64]int{}
				for i := range r.Comments {
					indexByID[r.Comments[i].ID] = i
				}
				for _, c := range r.Comments {
					indent := "  "
					if c.Kind == "review" && c.ParentID != 0 {
						indent = "    ↳ "
					}
					sev := c.Severity
					if !doClassify || sev == "" {
						sev = "-"
					}
					author := c.Author
					if output.AuthorColorEnabled() {
						author = output.ColorizeAuthor(author)
					}
					// header line with severity and author
					header := fmt.Sprintf("%s- [%s] @%s", indent, sev, author)
					if c.Path != "" {
						header += fmt.Sprintf(" (%s)", c.Path)
					}
					output.Printf(output.ModeAuto, "%s\n", header)
					// body on its own line; render as markdown to ANSI
					if strings.TrimSpace(c.Body) != "" {
						var rendered string
						if renderMD {
							rendered = output.RenderMarkdown(c.Body)
						} else {
							rendered = oneLiner(c.Body)
						}
						for _, line := range strings.Split(strings.TrimRight(rendered, "\n"), "\n") {
							output.Printf(output.ModeAuto, "%s  %s\n", indent, line)
						}
					}
				}
				output.Printf(output.ModeAuto, "\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Limit to a single repo (OWNER/REPO). If empty, searches across accessible repos")
	cmd.Flags().StringVar(&state, "state", "open", "PR state: open|closed|all")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output JSON")
	cmd.Flags().BoolVar(&excludeBots, "no-bots", true, "Exclude bot comments")
	cmd.Flags().BoolVar(&includeDrafts, "drafts", true, "Include draft PRs")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of PRs (0=all)")
	cmd.Flags().StringVar(&since, "since", "", "Only PRs updated on/after YYYY-MM-DD")
	cmd.Flags().BoolVar(&doClassify, "classify", false, "Classify comment severity and derive PR priority (uses opencode)")
	cmd.Flags().BoolVar(&renderMD, "md", true, "Render comment bodies as Markdown to ANSI (requires a compatible terminal)")
	cmd.Flags().BoolVar(&urlsOnly, "urls", false, "Print only PR URLs (one per line)")
	return cmd
}

func oneLiner(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func whoAmI() (string, error) {
	c := exec.Command("gh", "api", "/user")
	out, err := c.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("gh api /user failed: %s", string(ee.Stderr))
		}
		return "", err
	}
	var u struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(out, &u); err != nil {
		return "", err
	}
	if u.Login == "" {
		return "", fmt.Errorf("unable to resolve authenticated user")
	}
	return u.Login, nil
}

func listMyPRs(me, repo, state string, includeDrafts bool, limit int, since string) ([]ghPR, error) {
	// Use search/issues to find PRs authored by me
	parts := []string{"is:pr", fmt.Sprintf("author:%s", me)}
	if state == "" {
		state = "open"
	}
	if state == "open" {
		parts = append(parts, "is:open")
	} else if state == "closed" {
		parts = append(parts, "is:closed")
	}
	if repo != "" {
		parts = append(parts, fmt.Sprintf("repo:%s", repo))
	}
	if !includeDrafts {
		parts = append(parts, "-is:draft")
	}
	// Performance prefilter: only PRs with any comments
	parts = append(parts, "comments:>0")
	// Optional updated since
	if strings.TrimSpace(since) != "" {
		parts = append(parts, fmt.Sprintf("updated:>=%s", since))
	}
	q := strings.Join(parts, "+")
	args := []string{"api", "-X", "GET", fmt.Sprintf("search/issues?q=%s", q)}
	if limit <= 0 || limit > 100 {
		args = append(args, "--paginate", "-Fper_page=100")
	} else {
		args = append(args, fmt.Sprintf("-Fper_page=%d", limit))
	}
	c := exec.Command("gh", args...)
	out, err := c.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gh api search issues failed: %s", string(ee.Stderr))
		}
		return nil, err
	}
	// Parse results to collect repo and pr number via URL
	var res struct {
		Items []struct {
			Number      int    `json:"number"`
			Title       string `json:"title"`
			HTMLURL     string `json:"html_url"`
			PullRequest struct {
				URL string `json:"url"`
			} `json:"pull_request"`
		} `json:"items"`
	}
	_ = json.Unmarshal(out, &res)
	var prs []ghPR
	for _, it := range res.Items {
		prs = append(prs, ghPR{Number: it.Number, Title: it.Title, HTMLURL: it.HTMLURL, User: struct {
			Login string `json:"login"`
		}{Login: me}, State: state})
	}
	return prs, nil
}

func repoFromPRURL(url string) (string, error) {
	// https://github.com/OWNER/REPO/pull/123
	parts := strings.Split(url, "/")
	for i := range parts {
		if i > 1 && parts[i-1] == "github.com" {
			if i+1 < len(parts) {
				return parts[i] + "/" + parts[i+1], nil
			}
		}
	}
	return "", fmt.Errorf("cannot parse repo from url: %s", url)
}

func hasHumanComments(repo string, prNumber int, botRe *regexp.Regexp) (bool, error) {
	// Probe latest issue comment
	path1 := fmt.Sprintf("repos/%s/issues/%d/comments?per_page=1", repo, prNumber)
	c1 := exec.Command("gh", "api", path1)
	b1, err1 := c1.Output()
	if err1 == nil {
		var one []ghIssueComment
		if json.Unmarshal(b1, &one) == nil && len(one) > 0 {
			if !botRe.MatchString(one[0].User.Login) {
				return true, nil
			}
		}
	}
	// Probe latest review comment
	path2 := fmt.Sprintf("repos/%s/pulls/%d/comments?per_page=1", repo, prNumber)
	c2 := exec.Command("gh", "api", path2)
	b2, err2 := c2.Output()
	if err2 == nil {
		var one []ghReviewComment
		if json.Unmarshal(b2, &one) == nil && len(one) > 0 {
			if !botRe.MatchString(one[0].User.Login) {
				return true, nil
			}
		}
	}
	// If both failed or only bots observed
	return false, nil
}

func fetchIssueComments(repo string, prNumber int) ([]ghIssueComment, error) {
	path := fmt.Sprintf("repos/%s/issues/%d/comments", repo, prNumber)
	c := exec.Command("gh", "api", path, "--paginate")
	out, err := c.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gh api %s failed: %s", path, string(ee.Stderr))
		}
		return nil, err
	}
	var items []ghIssueComment
	if err := json.Unmarshal(out, &items); err == nil {
		return items, nil
	}
	// Fallback for newline-delimited arrays
	for _, chunk := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		var part []ghIssueComment
		if err := json.Unmarshal([]byte(chunk), &part); err == nil {
			items = append(items, part...)
		}
	}
	return items, nil
}

func fetchReviewComments(repo string, prNumber int) ([]ghReviewComment, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments", repo, prNumber)
	c := exec.Command("gh", "api", path, "--paginate")
	out, err := c.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gh api %s failed: %s", path, string(ee.Stderr))
		}
		return nil, err
	}
	var items []ghReviewComment
	if err := json.Unmarshal(out, &items); err == nil {
		return items, nil
	}
	for _, chunk := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		var part []ghReviewComment
		if err := json.Unmarshal([]byte(chunk), &part); err == nil {
			items = append(items, part...)
		}
	}
	return items, nil
}

func classifyComment(model, body string) (string, error) {
	if strings.TrimSpace(body) == "" {
		return "info", nil
	}
	// Simple prompt to opencode for classification
	prompt := fmt.Sprintf("Classify the following GitHub PR comment by severity as one of: blocker, high, medium, low, info. Respond with just the label. Comment: %q", body)
	// Use opencode CLI to run and capture output
	args := []string{"run", "-m", model, prompt}
	cmd := exec.Command("opencode", args...)
	b, err := cmd.Output()
	if err != nil {
		return "info", nil
	}
	label := strings.ToLower(strings.TrimSpace(string(b)))
	switch label {
	case "blocker", "high", "medium", "low", "info":
		return label, nil
	default:
		return "info", nil
	}
}

func derivePriority(comments []classifiedComment) string {
	priority := "none"
	order := map[string]int{"blocker": 5, "high": 4, "medium": 3, "low": 2, "info": 1}
	max := 0
	for _, c := range comments {
		if v := order[c.Severity]; v > max {
			max = v
		}
	}
	for k, v := range order {
		if v == max {
			priority = k
		}
	}
	if priority == "none" && len(comments) > 0 {
		priority = "info"
	}
	return priority
}
