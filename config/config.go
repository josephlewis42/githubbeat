// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period      time.Duration `config:"period"`
	JobTimeout  time.Duration `config:"period"`
	Repos       []string      `config:"repos"`
	Orgs        []string      `config:"orgs"`
	AccessToken string        `config:"access_token"`

	Forks         ListConfig     `config:"forks"`
	License       ExtendedConfig `config:"license"`
	Contributors  ListConfig     `config:"contributors"`
	Branches      ListConfig     `config:"branches"`
	Languages     ListConfig     `config:"languages"`
	Participation ExtendedConfig `config:"participation"`
	Downloads     ListConfig     `config:"downloads"`
}

type ListConfig struct {
	Enabled bool `config:"enabled"`
	List    bool `config:"list"`
}

type ExtendedConfig struct {
	Enabled bool `config:"enabled"`
}

var DefaultConfig = Config{
	Period:        30 * time.Second,
	JobTimeout:    10 * time.Second,
	Forks:         ListConfig{false, false},
	License:       ExtendedConfig{true},
	Contributors:  ListConfig{true, true},
	Branches:      ListConfig{true, false},
	Languages:     ListConfig{true, true},
	Participation: ExtendedConfig{true},
	Downloads:     ListConfig{true, false},
}
