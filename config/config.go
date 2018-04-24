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
	StartNow    bool          `config:"start_immediately"`

	Forks         PageableListConfig `config:"forks"`
	Contributors  PageableListConfig `config:"contributors"`
	Branches      PageableListConfig `config:"branches"`
	Languages     ListConfig         `config:"languages"`
	Participation ExtendedConfig     `config:"participation"`
	Downloads     PageableListConfig `config:"downloads"`
	Issues        IssuesConfig       `config:"issues"`
}

// ListConfig has configuration for metrics that have list outputs
// such as languages.
type ListConfig struct {
	Enabled bool `config:"enabled"`
	List    bool `config:"list"`
}

// PageableListConfig has configuration for metrics that have list
// outputs that require paging to fetch all of them such as contributors.
type PageableListConfig struct {
	Enabled bool `config:"enabled"`
	List    bool `config:"list"`
	Max     int  `config:"max_elements"`
}

// ExtendedConfig contains configuration options for metrics
// that require API calls beyond what comes in as part of
// the call to get repository information.
type ExtendedConfig struct {
	Enabled bool `config:"enabled"`
}

// IssuesConfig contains specific listing options for issues belonging to
// a repository.
type IssuesConfig struct {
	Enabled   bool     `config:"enabled"`
	List      bool     `config:"list"`
	Max       int      `config:"max_elements"`
	State     string   `config:"state"`
	Labels    []string `config:"labels"`
	Sort      string   `config:"sort"`
	Direction string   `config:"direction"`
}

// DefaultConfig has the application default configurations.
// These attempt to be sane defaults, balanced between API
// call count and useful information provided.
var DefaultConfig = Config{
	Period:        30 * time.Second,
	JobTimeout:    10 * time.Second,
	StartNow:      true,
	Forks:         PageableListConfig{false, false, -1},
	Contributors:  PageableListConfig{true, true, -1},
	Branches:      PageableListConfig{true, false, -1},
	Languages:     ListConfig{true, true},
	Participation: ExtendedConfig{true},
	Downloads:     PageableListConfig{true, false, -1},
	Issues:        IssuesConfig{true, true, -1, "open", []string{}, "created", "desc"},
}
