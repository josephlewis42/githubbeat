# Githubbeat

## Why Githubbeat ?

Monitoring the activity of a project through externals services involved in the
lifetime of an opensource project can provide important insights about its
health and the activity of its community.

Githubbeat was built in order to monitor the activity of various github
repositories or even complete organizations through basic github stats like
open issues count or number of forks.

As a beat It ships everything directly to Elasticsearch or Logstash, depending
of your needs.

## How to use this ?

To run Githubbeats, simply use:

```
# -e logs to stdeer and disables syslog/file output
./githubbeat -c <path_to_your_githubbeat.yml> -e
```

## Building

This beater was built according to the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html) and thus uses the pre-defined build system.

With the appropriate dependencies this will generate a binary in the same
directory..

```
make
```

## Test

To test Githubbeat, run the following command:

```
make testsuite
```

## Todo

- [ ] Export total opened PRs
- [ ] Export total commit count per repository
- [ ] Export releases count per repository
- [ ] Export branches count per repository

Of course, any suggestions are welcome !
