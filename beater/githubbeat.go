package beater

//go:generate go run gen-lists.go

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/google/go-github/github"

	"github.com/josephlewis42/githubbeat/config"
)

// Githubbeat collects github repositories statistics
type Githubbeat struct {
	done     chan struct{}
	config   config.Config
	ghClient *github.Client
	client   beat.Client
}

// New creates a new instance of a GithubBeat
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	return &Githubbeat{
		done:   make(chan struct{}),
		config: config,
	}, nil
}

// Run runs the beat
func (bt *Githubbeat) Run(b *beat.Beat) error {
	logp.Info("githubbeat is running! Hit CTRL-C to stop it.")
	logp.Info("configuration: %+v", bt.config)

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	bt.ghClient, err = newGithubClient(bt.config)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	timer := time.NewTimer(time.Millisecond)

	// Run out the timer If the beat isn't going to start immediately.
	if !bt.config.StartNow {
		<-timer.C
	}

	rootCtx, cancelRootCtx := context.WithCancel(context.Background())

	for {
		select {
		case <-bt.done:
			cancelRootCtx()
			return nil
		case <-ticker.C:
		case <-timer.C:
		}

		logp.Info("Collecting events.")
		jobCtx, jobCancel := context.WithTimeout(rootCtx, bt.config.JobTimeout)
		defer jobCancel()
		bt.collectReposEvents(jobCtx, bt.config.Repos)
		bt.collectOrgsEvents(jobCtx, bt.config.Orgs)
	}
}

// Stop stops the running beat
func (bt *Githubbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func newGithubClient(config config.Config) (*github.Client, error) {
	client, err := setupClient(config.AccessToken)
	if err != nil {
		return nil, err
	}

	baseUrl, err := config.Enterprise.GetBaseUrl()
	if err != nil {
		return nil, err
	}

	if baseUrl != nil {
		logp.Info("Using custom BaseUrl")
		client.BaseURL = baseUrl
	}

	uploadUrl, err := config.Enterprise.GetUploadUrl()
	if err != nil {
		return nil, err
	}

	if uploadUrl != nil {
		logp.Info("Using custom UploadUrl")
		client.UploadURL = uploadUrl
	}

	// Test connection
	ctx := context.Background()
	if _, _, err := client.Repositories.List(ctx, "", nil); err != nil {
		return nil, err
	}

	return client, nil
}

func setupClient(accessToken string) (*github.Client, error) {
	if accessToken == "" {
		logp.Info("Running in unauthenticated mode.")
		return github.NewClient(nil), nil
	}

	logp.Info("Running in authentcated mode.")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)

	return github.NewClient(oauth2.NewClient(ctx, ts)), nil
}

func (bt *Githubbeat) collectOrgsEvents(ctx context.Context, orgs []string) {
	for _, org := range orgs {
		go func(ctx context.Context, org string) {
			repos, _, err := bt.ghClient.Repositories.ListByOrg(ctx, org, nil)

			if err != nil {
				logp.Err("Failed to collect org repos listing, got :", err)
				return
			}

			for _, repo := range repos {
				bt.client.Publish(beat.Event{
					Timestamp: time.Now(),
					Fields:    bt.newFullRepoEvent(ctx, repo),
				})
			}
		}(ctx, org)
	}
}

func (bt *Githubbeat) collectReposEvents(ctx context.Context, repos []string) {
	for _, repoName := range repos {
		go func(ctx context.Context, repo string) {
			r := strings.Split(repo, "/")

			if len(r) != 2 {
				logp.Err("Invalid repo name format, expected [org]/[name], got: ", repo)
				return
			}

			res, _, err := bt.ghClient.Repositories.Get(ctx, r[0], r[1])

			if err != nil {
				logp.Err("Failed to collect event, got :", err)
				return
			}

			bt.client.Publish(beat.Event{
				Timestamp: time.Now(),
				Fields:    bt.newFullRepoEvent(ctx, res),
			})
		}(ctx, repoName)
	}
}

func (bt *Githubbeat) extractContributors(contributors []*github.Contributor, err error) []common.MapStr {
	users := []common.MapStr{}
	if err == nil {
		for _, contributor := range contributors {
			userInfo := common.MapStr{
				"name":          contributor.GetLogin(),
				"contributions": contributor.GetContributions(),
			}

			users = append(users, userInfo)
		}
	}

	return users
}

func (bt *Githubbeat) extractBranches(branches []*github.Branch, err error) []common.MapStr {
	// name:author pairs
	branchList := []common.MapStr{}

	if err == nil {
		for _, branch := range branches {
			branchInfo := common.MapStr{
				"name": branch.GetName(),
				"sha":  branch.Commit.GetSHA(),
			}

			branchList = append(branchList, branchInfo)
		}
	}

	return branchList
}

type collector func(repositoryClient *repositoryClient) common.MapStr

func (bt *Githubbeat) newFullRepoEvent(ctx context.Context, repo *github.Repository) common.MapStr {
	rc := NewRepositoryClient(ctx, bt.ghClient, repo)
	data := extractRepoData(repo)

	// beat metadata
	data["@timestamp"] = common.Time(time.Now())
	data["type"] = "githubbeat"

	addIf := func(key string, c collector) {
		res := c(rc)
		if res != nil {
			data[key] = res
		}
	}
	addIf("fork_list", bt.collectForks)
	addIf("contributor_list", bt.collectContributors)
	addIf("branch_list", bt.collectBranches)
	addIf("languages", bt.collectLanguages)
	addIf("participation", bt.collectParticipation)
	addIf("downloads", bt.collectDownloads)
	addIf("issues", bt.collectIssues)

	return data
}

func extractRepoData(repo *github.Repository) common.MapStr {
	license := common.MapStr{
		"key":     repo.GetLicense().GetKey(),
		"name":    repo.GetLicense().GetName(),
		"spdx_id": repo.GetLicense().GetSPDXID(),
	}

	return common.MapStr{
		"repo":        repo.GetName(),
		"owner":       repo.Owner.GetLogin(),
		"stargazers":  repo.GetStargazersCount(),
		"forks":       repo.GetForksCount(),
		"watchers":    repo.GetWatchersCount(),
		"open_issues": repo.GetOpenIssuesCount(),
		"subscribers": repo.GetSubscribersCount(),
		"network":     repo.GetNetworkCount(),
		"size":        repo.GetSize(),
		"license":     license,
	}
}

func (bt *Githubbeat) extractLanguages(langs map[string]int, err error) []common.MapStr {
	sum := 0
	for _, count := range langs {
		sum += count
	}

	out := []common.MapStr{}
	for lang, count := range langs {
		out = append(out, common.MapStr{
			"name":  lang,
			"bytes": count,
			"ratio": float64(count) / float64(sum),
		})
	}

	return out
}

func (bt *Githubbeat) extractForks(forks []*github.Repository, err error) []common.MapStr {
	forkInfo := []common.MapStr{}
	for _, repo := range forks {
		forkInfo = append(forkInfo, extractRepoData(repo))
	}

	return forkInfo
}

func (bt *Githubbeat) extractParticipation(participation *github.RepositoryParticipation, err error) common.MapStr {
	all := 0
	owner := 0

	if participation != nil {
		all = sumIntArray(participation.All)
		owner = sumIntArray(participation.Owner)
	}

	return common.MapStr{
		"all":       all,
		"owner":     owner,
		"community": all - owner,
		"period":    "year",
	}
}

func (bt *Githubbeat) extractDownloads(releases []*github.RepositoryRelease, err error) common.MapStr {
	totalDownloads := 0
	out := []common.MapStr{}
	for _, release := range releases {
		releaseDownloads := 0

		for _, asset := range release.Assets {
			releaseDownloads += asset.GetDownloadCount()
		}

		totalDownloads += releaseDownloads

		out = append(out, common.MapStr{
			"id":        release.GetID(),
			"name":      release.GetName(),
			"downloads": releaseDownloads,
		})
	}

	return common.MapStr{
		"total_downloads": totalDownloads,
		"releases":        out,
	}
}

func createListMapStr(list []common.MapStr, err error, enableList bool) common.MapStr {
	out := common.MapStr{"count": len(list), "error": err}

	if enableList {
		out["list"] = list
	}

	return out
}

func appendError(input common.MapStr, err error) common.MapStr {
	if err != nil {
		input["error"] = err.Error()
	}

	return input
}

func sumIntArray(array []int) int {
	sum := 0
	for _, i := range array {
		sum += i
	}

	return sum
}

func NewRepositoryClient(ctx context.Context, client *github.Client, repo *github.Repository) *repositoryClient {
	return &repositoryClient{
		ctx:    ctx,
		client: client,
		repo:   repo,
	}
}

type repositoryClient struct {
	ctx    context.Context
	client *github.Client
	repo   *github.Repository
}

func (rc *repositoryClient) GetOwner() string {
	return rc.repo.Owner.GetLogin()
}

func (rc *repositoryClient) GetName() string {
	return rc.repo.GetName()
}
