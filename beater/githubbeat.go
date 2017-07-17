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
	done   chan struct{}
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Githubbeat{
		done:   make(chan struct{}),
		config: config,
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

			events, err := bt.collectEvents(jobCtx)

			if err != nil {
				jobCancel()
				logp.Err("Failed to collect events, got", err)
				break
			}

			for _, event := range events {
				bt.client.PublishEvent(event)
				logp.Info("Event sent")
			}
		}

		counter++
	}
}

func (bt *Githubbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *Githubbeat) collectEvents(ctx context.Context) ([]common.MapStr, error) {
	var res []common.MapStr

	client := github.NewClient(nil)

	repo, _, err := client.Repositories.Get(ctx, "containous", "traefik")

	if err != nil {
		return []common.MapStr{}, err
	}

	return append(res, bt.newRepoEvent(repo)), nil
}

func (Githubbeat) newRepoEvent(repo *github.Repository) common.MapStr {
	return common.MapStr{
		"@timestamp": common.Time(time.Now()),
		"type":       "githubbeat",
		"stargazers": repo.GetStargazersCount(),
		"forks":      repo.GetForksCount(),
		"watchers":   repo.GetWatchersCount(),
		"size":       repo.GetSize(),
	}
}
