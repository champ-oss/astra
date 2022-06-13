# astra

A GitHub Action which helps automate the management of GitHub pull requests for your organization

[![.github/workflows/gotest.yml](https://github.com/champ-oss/astra/actions/workflows/gotest.yml/badge.svg?branch=main)](https://github.com/champ-oss/astra/actions/workflows/gotest.yml)
[![.github/workflows/golint.yml](https://github.com/champ-oss/astra/actions/workflows/golint.yml/badge.svg?branch=main)](https://github.com/champ-oss/astra/actions/workflows/golint.yml)
[![.github/workflows/release.yml](https://github.com/champ-oss/astra/actions/workflows/release.yml/badge.svg)](https://github.com/champ-oss/astra/actions/workflows/release.yml)
[![.github/workflows/sonar.yml](https://github.com/champ-oss/astra/actions/workflows/sonar.yml/badge.svg)](https://github.com/champ-oss/astra/actions/workflows/sonar.yml)

[![SonarCloud](https://sonarcloud.io/images/project_badges/sonarcloud-black.svg)](https://sonarcloud.io/summary/new_code?id=champ-oss_astra)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=champ-oss_astra&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=champ-oss_astra)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=champ-oss_astra&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=champ-oss_astra)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=champ-oss_astra&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=champ-oss_astra)

## Features
- Scans repositories in your organization based on a name prefix
- Automatically reruns failed workflows for pull requests
- Enables auto-merge for pull requests
- Supports processing pull requests only opened by certain actors (ex: Dependabot)

## Example Usage

```yaml
on:
  schedule:
    - cron: "0 0 * * *"
  workflow_dispatch:
  push:

jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: champ-oss/astra
        with:
          dry-run: false
          debug: true
          owner: myorganization
          app-id: 123
          installation-id: 123
          pem: ${{ secrets.PEM }}
          default-branch: main
          wait-seconds-between-requests: 2
          max-run-attempts: 3
          repo-prefixes: |
            my-repo
          actors: |
            dependabot
```

## How to Set Up A GitHub App
A GitHub App is required to use this Action.

Overview: https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app

- The name of the app can be whatever makes sense to you
- You can use `https://localhost` as the Homepage URL
- All other options can be left at the default settings
- After creation, install the GitHub App into your organization
- Make note of the `App ID` and `Installation ID` which must be passed in to this Action
- Generate a private key in the app and save the key. The contents of the key should be base64 encoded when passed into this Action. Example: `cat key.pem | base64 -w 0`

### GitHub App Permissions

These permission settings are required when creating the GitHub App
If a permission is not listed below then the default setting (no access) should be used.

- Actions: Read and Write
- Administration: Read-only
- Checks: Read and Write
- Commit statuses: Read-only
- Contents: Read and Write
- Metadata:  Read-only
- Pull Requests: Read and Write
- Workflows: Read and Write

## Parameters
| Parameter | Required | Description |
| --- | --- | --- |
| owner | true | Name of GitHub organization or owner |
| app-id | true | GitHub App ID |
| installation-id | true | GitHub Installation ID |
| pem | true | GitHub App PEM file |
| repo-prefixes | true | Only repositories containing these prefixes will be processed |
| actors | false | Only pull requests by these actors will be processed |
| dry-run | false | Scan repositories but do not rerun workflows or make any changes |
| debug | false | Enable debug logging |
| default-branch | false | The name of the default branch for your repositories |
| wait-seconds-between-requests | false | Slow down requests against the GitHub API to avoid throttling |
| max-run-attempts | false | A workflow will not be restarted if it has failed this many times |


## Contributing

