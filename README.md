# Githubbeat

[![Build Status](https://travis-ci.org/josephlewis42/githubbeat.svg?branch=master)](https://travis-ci.org/josephlewis42/githubbeat) [![Go Report Card](https://goreportcard.com/badge/github.com/josephlewis42/githubbeat)](https://goreportcard.com/report/github.com/josephlewis42/githubbeat) 

This repository is a fork of [jlevesy's original githubbeat](github.com/jlevesy/githubbeat) that adds new capabilities while maintaining backwards formatting compatability so it can be a drop-in replacement.

## Why Githubbeat ?

Monitoring the activity of a project through externals services involved in the
lifetime of an open source project can provide important insights about its
health and the activity of its community.

Githubbeat supports the following metrics:

- Basic metrics
  - license (SPDX ID, name, SPDX key)
  - network count
  - size (bytes)
  - open issue count
  - star count
  - subscriber count
  - watcher count
  - participation (total, community, owner)

- Extended Metrics
 - branches (count, list{name, sha})
 - contributors (count, list{name, contribution count})
 - downloads (release count, total downloads, list{id, name, download count})
 - forks (count, list{the same basic metrics as above})
 - languages (count, list{name, bytes, ratio of total repository})

## How to use this ?

TODO

## Todo

- [x] Write a proof of concept
- [x] Enable CI
- [x] Open a PR to add Githubbeat to the community beats [#4723](https://github.com/elastic/beats/pull/4723).
- [ ] Write unit tests and integration tests
- [ ] Setup Delivery pipeline in order to publish new releases of githubbeat automatically through Github.
  - Targets
    - DEB 32/64
    - RPM 32/64
    - OSX
    - Docker image
- [ ] Setup a test environment using docker-compose with githubbeat, es & kibana
- [ ] Build a githubbeat kibana dashboard
- [ ] Write documentation about how to use the beat
