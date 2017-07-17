package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

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
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
			events, err := bt.collectEvents()
			if err != nil {
				logp.Err("Failed to collect events, got", err)
				break
			}

			for _, event := range events {
				client.PublishEvent(event)
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

func (bt *Githubbeat) collectEvents() ([]common.MapStr, error) {
	return []common.MapStr{}, nil
}
