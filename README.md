# PodToLog

Tool to build a Dynatrace query URL from a Kubernetes pod's name and namespace. This uses your existing cluster credentials to gather the data and build the url (eg: `oc whoami`).

## Usage

1. Run `podtolog --namespace <namespace> <podname>`

2. Open the resulting link in your browser.
