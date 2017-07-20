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
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/google/go-github/github"

	"github.com/jlevesy/githubbeat/config"
)

// Githubbeat collects github repositories statistics
type Githubbeat struct {
	done     chan struct{}
	config   config.Config
	ghClient *github.Client
	client   publisher.Client
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

	bt.client = b.Publisher.Connect()

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
				bt.client.PublishEvent(bt.newRepoEvent(repo))
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

			bt.client.PublishEvent(bt.newRepoEvent(res))
		}(ctx, repoName)
	}
}

func (Githubbeat) newRepoEvent(repo *github.Repository) common.MapStr {
	return common.MapStr{
		"@timestamp":  common.Time(time.Now()),
		"type":        "githubbeat",
		"repo":        repo.GetName(),
		"owner":       repo.Owner.GetLogin(),
		"stargazers":  repo.GetStargazersCount(),
		"forks":       repo.GetForksCount(),
		"watchers":    repo.GetWatchersCount(),
		"open_issues": repo.GetOpenIssuesCount(),
		"subscribers": repo.GetSubscribersCount(),
		"network":     repo.GetNetworkCount(),
		"size":        repo.GetSize(),
	}
}
