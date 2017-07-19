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
}

var DefaultConfig = Config{
	Period:     30 * time.Second,
	JobTimeout: 10 * time.Second,
}
