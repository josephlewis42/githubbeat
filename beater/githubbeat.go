package beater

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/google/go-github/github"

	"github.com/jlevesy/githubbeat/config"
)

type Githubbeat struct {
	done     chan struct{}
	config   config.Config
	ghClient *github.Client
	client   publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Githubbeat{
		done:     make(chan struct{}),
		config:   config,
		ghClient: github.NewClient(nil),
	}
	return bt, nil
}

func (bt *Githubbeat) Run(b *beat.Beat) error {
	logp.Info("githubbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	rootCtx, cancelRootCtx := context.WithCancel(context.Background())
	for {
		select {
		case <-bt.done:
			cancelRootCtx()
			return nil
		case <-ticker.C:
			jobCtx, jobCancel := context.WithTimeout(rootCtx, bt.config.JobTimeout)
			bt.collectReposEvent(jobCtx, bt.config.Repos)
			jobCancel()
		}
	}
}

func (bt *Githubbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *Githubbeat) collectReposEvent(ctx context.Context, repos []string) {
	out := make(chan common.MapStr, len(repos))
	wg := sync.WaitGroup{}

	wg.Add(len(repos))

	for _, repoName := range repos {
		go func(ctx context.Context, repoName string, out chan<- common.MapStr, wg *sync.WaitGroup) {
			r := strings.Split(repoName, "/")

			if len(r) != 2 {
				logp.Err("Invalid repo name format, expected [org]/[name]")
				wg.Done()
				return
			}

			res, _, err := bt.ghClient.Repositories.Get(ctx, r[0], r[1])

			if err != nil {
				logp.Err("Failed to collect event, got :", err)
				wg.Done()
				return
			}

			out <- bt.newRepoEvent(res)
			wg.Done()
		}(ctx, repoName, out, &wg)
	}

	wg.Wait()

	close(out)

	for event := range out {
		bt.client.PublishEvent(event)
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
		"issues":      repo.GetOpenIssuesCount(),
		"subscribers": repo.GetSubscribersCount(),
		"network":     repo.GetNetworkCount(),
		"size":        repo.GetSize(),
	}
}
