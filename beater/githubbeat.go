package beater

import (
	"context"
	"fmt"
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
	counter := 1
	rootCtx, cancelRootCtx := context.WithCancel(context.Background())
	for {
		select {
		case <-bt.done:
			cancelRootCtx()
			return nil
		case <-ticker.C:
			jobCtx, jobCancel := context.WithTimeout(rootCtx, 10*time.Second)
			defer jobCancel()

			event, err := bt.collectRepoEvent(jobCtx, "containous", "traefik")

			if err != nil {
				logp.Err("Failed to collect events, got", err)
				jobCancel()
				break
			}

			bt.client.PublishEvent(event)
		}

		counter++
	}
}

func (bt *Githubbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *Githubbeat) collectRepoEvent(ctx context.Context, owner, repo string) (common.MapStr, error) {
	r, _, err := bt.ghClient.Repositories.Get(ctx, owner, repo)

	if err != nil {
		return common.MapStr{}, err
	}

	return bt.newRepoEvent(r), nil
}

func (Githubbeat) newRepoEvent(repo *github.Repository) common.MapStr {
	return common.MapStr{
		"@timestamp":   common.Time(time.Now()),
		"type":         "githubbeat",
		"repo":         repo.GetName(),
		"organization": repo.Organization.GetLogin(),
		"stargazers":   repo.GetStargazersCount(),
		"forks":        repo.GetForksCount(),
		"watchers":     repo.GetWatchersCount(),
		"issues":       repo.GetOpenIssuesCount(),
		"subscribers":  repo.GetSubscribersCount(),
		"network":      repo.GetNetworkCount(),
		"size":         repo.GetSize(),
	}
}
