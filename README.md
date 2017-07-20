# Githubbeat

[![Build Status](https://travis-ci.org/jlevesy/githubbeat.svg?branch=master)](https://travis-ci.org/jlevesy/githubbeat)

This is currently a work in progress.

At the momment I have a working POC, but it is not well tested, packaged nor documented.
Please have a look in the [TODO](https://github.com/jlevesy/githubbeat#todo)  section to see what I plan to do ;)

## Why Githubbeat ?

Monitoring the activity of a project through externals services involved in the
lifetime of an opensource project can provide important insights about its
health and the activity of its community.

Githubbeat was built in order to monitor the activity of various github
repositories or even complete organizations through basic github stats like
open issues count or number of forks.

As a beat It ships everything directly to all supported outputs, depending of
your needs.

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

## Feature ideas

- [ ] Export total opened PRs
- [ ] Export total commit count per repository
- [ ] Export releases count per repository
- [ ] Export branches count per repository

Of course, any suggestions are welcome !
