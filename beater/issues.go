package beater

import (
	"time"

	"github.com/elastic/beats/libbeat/common"

	"github.com/google/go-github/github"
)

func (bt *Githubbeat) collectIssues(rc *repositoryClient) common.MapStr {
	if !bt.config.Issues.Enabled {
		return nil
	}

	options := &github.IssueListByRepoOptions{
		State:     bt.config.Issues.State,
		Labels:    bt.config.Issues.Labels,
		Sort:      bt.config.Issues.Sort,
		Direction: bt.config.Issues.Direction,
	}

	rawResults, err := rc.ListIssues(bt.config.Issues.Max, options)

	formatted := bt.extractIssues(rawResults, err)

	return createListMapStr(formatted, err, bt.config.Issues.List)
}

func (bt *Githubbeat) extractIssues(issues []*github.Issue, err error) []common.MapStr {

	formatIfExists := func(t *time.Time) string {
		if t == nil {
			return ""
		}

		return t.Format(time.RFC3339)
	}

	labelNames := func(labels []github.Label) []string {
		out := []string{}

		for _, label := range labels {
			out = append(out, label.GetName())
		}

		return out
	}

	out := []common.MapStr{}
	if err == nil {
		for _, issue := range issues {
			issueInfo := common.MapStr{
				"id":              issue.GetID(),
				"number":          issue.GetNumber(),
				"state":           issue.GetState(),
				"title":           issue.GetTitle(),
				"body":            issue.GetBody(),
				"user":            issue.GetUser().GetLogin(),
				"labels":          labelNames(issue.Labels),
				"assignee":        issue.GetAssignee().GetLogin(),
				"milestone":       issue.GetMilestone().GetTitle(),
				"locked":          issue.GetLocked(),
				"comment_count":   issue.GetComments(),
				"is_pull_request": issue.IsPullRequest(),
				"closed_at":       formatIfExists(issue.ClosedAt),
				"created_at":      formatIfExists(issue.CreatedAt),
				"updated_at":      formatIfExists(issue.UpdatedAt),
			}

			out = append(out, issueInfo)
		}
	}

	return out
}

func (rc *repositoryClient) ListIssues(max int, opt *github.IssueListByRepoOptions) ([]*github.Issue, error) {
	var results []*github.Issue

	for {
		list, resp, err := rc.client.Issues.ListByRepo(rc.ctx, rc.GetOwner(), rc.GetName(), opt)
		if err != nil {
			return results, err
		}

		results = append(results, list...)
		if resp.NextPage == 0 || (len(results) >= max && max > 0) {
			break
		}

		opt.Page = resp.NextPage
	}

	return results, nil
}
