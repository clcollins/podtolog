# PodToLog

Tool to build a Dynatrace query URL from a Kubernetes pod's `.metadata.uid` value.

## Usage

1. Export the `PODTOLOG_HOST` environment variable with the hostname of a Dynatrace instance. (This may need some work.)

2. Run `podtolog < POD UID >`

3. Open the resulting link in your browser.

_Note:_ You may also create a `~/.config/podtolog/config` file with a single value: `host: <dynatrace instance>`, instead of setting the environment variable.
