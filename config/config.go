// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

// Config contains the user-settable configuration options for
// the beat.
type Config struct {
	Period      time.Duration `config:"period"`
	JobTimeout  time.Duration `config:"period"`
	Repos       []string      `config:"repos"`
	Orgs        []string      `config:"orgs"`
	AccessToken string        `config:"access_token"`

	Forks         ListConfig     `config:"forks"`
	Contributors  ListConfig     `config:"contributors"`
	Branches      ListConfig     `config:"branches"`
	Languages     ListConfig     `config:"languages"`
	Participation ExtendedConfig `config:"participation"`
	Downloads     ListConfig     `config:"downloads"`
}

// ListConfig has configuration for metrics that have list outputs
// such as contributors.
type ListConfig struct {
	Enabled bool `config:"enabled"`
	List    bool `config:"list"`
}

// ExtendedConfig contains configuration options for metrics
// that require API calls beyond what comes in as part of
// the call to get repository information.
type ExtendedConfig struct {
	Enabled bool `config:"enabled"`
}

// DefaultConfig has the application default configurations.
// These attempt to be sane defaults, balanced between API
// call count and useful information provided.
var DefaultConfig = Config{
	Period:        30 * time.Second,
	JobTimeout:    10 * time.Second,
	Forks:         ListConfig{false, false},
	Contributors:  ListConfig{true, true},
	Branches:      ListConfig{true, false},
	Languages:     ListConfig{true, true},
	Participation: ExtendedConfig{true},
	Downloads:     ListConfig{true, false},
}
