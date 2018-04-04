# Githubbeat

[![Build Status](https://travis-ci.org/josephlewis42/githubbeat.svg?branch=master)](https://travis-ci.org/josephlewis42/githubbeat) 
[![Go Report Card](https://goreportcard.com/badge/github.com/josephlewis42/githubbeat)](https://goreportcard.com/report/github.com/josephlewis42/githubbeat) 
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)



This repository is a fork of [jlevesy's original githubbeat](https://github.com/jlevesy/githubbeat) that adds new capabilities while maintaining backwards formatting compatability so it can be a drop-in replacement.

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

## License

```
Copyright (c) 2016 Julien Levesy

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
