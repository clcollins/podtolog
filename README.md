# PodToLog

[![Go](https://github.com/clcollins/podtolog/actions/workflows/go.yml/badge.svg)](https://github.com/clcollins/podtolog/actions/workflows/go.yml)

PodToLog is a tool to build a Dynatrace query URL from a Kubernetes pod's name and optional namespace. If no namespace is supplied, the current namespace is assumed. 

PodToLog uses your existing cluster credentials to gather the data and build the url (eg: `oc whoami`).

```
# Usage

podtolog --help
Usage:
  podtolog (POD) [flags]

Flags:
  -h, --help               help for podtolog
  -n, --namespace string   namespace {default: current namespace}
```
