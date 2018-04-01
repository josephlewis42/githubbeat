package beater

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

// New creates  a new instance of a GithubBeat
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

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ghClient, err := newGithubClient(bt.config.AccessToken)

	if err != nil {
		return err
	}

	bt.ghClient = ghClient

	ticker := time.NewTicker(bt.config.Period)

	rootCtx, cancelRootCtx := context.WithCancel(context.Background())

	for {
		select {
		case <-bt.done:
			cancelRootCtx()
			return nil
		case <-ticker.C:
			logp.Info("Collecting events.")
			jobCtx, jobCancel := context.WithTimeout(rootCtx, bt.config.JobTimeout)
			defer jobCancel()
			bt.collectReposEvents(jobCtx, bt.config.Repos)
			bt.collectOrgsEvents(jobCtx, bt.config.Orgs)
		}
	}
}

// Stop stops the running beat
func (bt *Githubbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func newGithubClient(accessToken string) (*github.Client, error) {
	if accessToken == "" {
		logp.Info("Running in unauthentcated mode.")
		return github.NewClient(nil), nil
	}

	logp.Info("Running in authentcated mode.")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)

	client := github.NewClient(oauth2.NewClient(ctx, ts))

	if _, _, err := client.Repositories.List(ctx, "", nil); err != nil {
		return nil, err
	}

	return client, nil
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

func (bt *Githubbeat) getContributions(owner, repository string, ctx context.Context) common.MapStr {
	users := []common.MapStr{}

	contributors, _, err := bt.ghClient.Repositories.ListContributors(ctx, owner, repository, nil)
	if err == nil {
		for _, contributor := range contributors {
			userInfo := common.MapStr{
				"name":          contributor.GetLogin(),
				"contributions": contributor.GetContributions(),
			}

			users = append(users, userInfo)
		}
	}

	return createListMapStr(users, err, bt.config.Contributors.List)
}

func (bt *Githubbeat) getBranches(owner, repository string, ctx context.Context) common.MapStr {
	// name:author pairs
	branchList := []common.MapStr{}

	branches, _, err := bt.ghClient.Repositories.ListBranches(ctx, owner, repository, nil)
	if err == nil {
		for _, branch := range branches {
			branchInfo := common.MapStr{
				"name": branch.GetName(),
				"sha":  branch.Commit.GetSHA(),
			}

			branchList = append(branchList, branchInfo)
		}
	}

	return createListMapStr(branchList, err, bt.config.Branches.List)
}

type collector func(owner, repository string, ctx context.Context) common.MapStr

func (bt *Githubbeat) newFullRepoEvent(ctx context.Context, repo *github.Repository) common.MapStr {

	data := bt.extractRepoData(repo)

	// beat metadata
	data["@timestamp"] = common.Time(time.Now())
	data["type"] = "githubbeat"

	// extended info
	owner := repo.Owner.GetLogin()
	repository := repo.GetName()

	addIf := func(key string, c collector, condition bool) {
		if condition {
			data[key] = c(owner, repository, ctx)
		}
	}

	addIf("fork_list", bt.collectForkInfo, bt.config.Forks.Enabled)
	addIf("contributor_list", bt.getContributions, bt.config.Contributors.Enabled)
	addIf("branch_list", bt.getBranches, bt.config.Branches.Enabled)
	addIf("languages", bt.collectLanguageInfo, bt.config.Languages.Enabled)
	addIf("participation", bt.collectParticipation, bt.config.Participation.Enabled)
	addIf("downloads", bt.collectDownloads, bt.config.Downloads.Enabled)

	return data
}

func (bt *Githubbeat) extractRepoData(repo *github.Repository) common.MapStr {
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

func (bt *Githubbeat) collectLanguageInfo(owner, repository string, ctx context.Context) common.MapStr {
	langs, _, err := bt.ghClient.Repositories.ListLanguages(ctx, owner, repository)

	// Enable totals so we can get a ratio
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

	return createListMapStr(out, err, bt.config.Languages.List)
}

func (bt *Githubbeat) collectForkInfo(owner, repository string, ctx context.Context) common.MapStr {
	forks, _, err := bt.ghClient.Repositories.ListForks(ctx, owner, repository, nil)

	forkInfo := []common.MapStr{}
	for _, repo := range forks {
		forkInfo = append(forkInfo, bt.extractRepoData(repo))
	}

	return createListMapStr(forkInfo, err, bt.config.Forks.List)
}

func (bt *Githubbeat) collectParticipation(owner, repository string, ctx context.Context) common.MapStr {
	participation, _, err := bt.ghClient.Repositories.ListParticipation(ctx, owner, repository)

	return appendError(bt.extractParticipationData(participation), err)
}

func (bt *Githubbeat) extractParticipationData(participation *github.RepositoryParticipation) common.MapStr {
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

func (bt *Githubbeat) collectDownloads(owner, repository string, ctx context.Context) common.MapStr {
	releases, _, err := bt.ghClient.Repositories.ListReleases(ctx, owner, repository, nil)

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
		"error":           err.Error(),
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
